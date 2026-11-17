// Package lyrics parses and normalizes lyric files and fetches them from
// providers. The renderer never sees a provider-specific format.
package lyrics

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/LuiisDev21/stefani/internal/domain"
)

var (
	// [mm:ss], [mm:ss.xx], [mm:ss:xxx] at the start of a line, possibly several.
	lineTagRe = regexp.MustCompile(`^((?:\[\d{1,3}:\d{2}(?:[.:]\d{1,3})?\])+)(.*)$`)
	oneTagRe  = regexp.MustCompile(`\[(\d{1,3}):(\d{2})(?:[.:](\d{1,3}))?\]`)
	// Enhanced LRC word timing: <mm:ss.xx>
	wordTagRe = regexp.MustCompile(`<(\d{1,3}):(\d{2})(?:[.:](\d{1,3}))?>`)
	metaTagRe = regexp.MustCompile(`^\[(ar|ti|al|by|offset|length|re|ve):(.*)\]$`)
)

// Parsed is the normalized result of parsing a lyric text.
type Parsed struct {
	Lines []domain.Line
	Sync  domain.SyncLevel
	// Title, Artist, Album and OffsetMs come from LRC metadata when present.
	Title    string
	Artist   string
	Album    string
	OffsetMs int64
}

// Parse reads LRC (plain, synchronized or enhanced with word timings) and
// returns normalized lines. It never fails on malformed input: bad lines are
// skipped so a single broken tag cannot hide a whole document.
func Parse(text string) Parsed {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var parsed Parsed
	type pending struct {
		startMs int64
		text    string
		words   []domain.Word
		hasWord bool
	}
	var entries []pending

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if m := metaTagRe.FindStringSubmatch(line); m != nil {
			switch m[1] {
			case "ti":
				parsed.Title = strings.TrimSpace(m[2])
			case "ar":
				parsed.Artist = strings.TrimSpace(m[2])
			case "al":
				parsed.Album = strings.TrimSpace(m[2])
			case "offset":
				if ms, err := strconv.ParseInt(strings.TrimSpace(m[2]), 10, 64); err == nil {
					parsed.OffsetMs = ms
				}
			}
			continue
		}

		m := lineTagRe.FindStringSubmatch(line)
		if m == nil {
			// Untimed text: keep it as an unsynced line once.
			entries = append(entries, pending{startMs: -1, text: strings.TrimSpace(line)})
			continue
		}
		timestamps := oneTagRe.FindAllStringSubmatch(m[1], -1)
		body := strings.TrimSpace(m[2])
		words, stripped, hasWord := extractWords(body)
		for _, ts := range timestamps {
			start := tagMillis(ts[1], ts[2], ts[3])
			entries = append(entries, pending{
				startMs: start,
				text:    stripped,
				words:   words,
				hasWord: hasWord,
			})
		}
	}

	timed := make([]pending, 0, len(entries))
	untimed := make([]pending, 0, len(entries))
	for _, e := range entries {
		if e.startMs < 0 {
			untimed = append(untimed, e)
		} else {
			timed = append(timed, e)
		}
	}
	sort.SliceStable(timed, func(i, j int) bool { return timed[i].startMs < timed[j].startMs })

	if len(timed) == 0 {
		parsed.Sync = domain.SyncUnsynced
		for _, e := range untimed {
			if e.text == "" {
				continue
			}
			parsed.Lines = append(parsed.Lines, domain.Line{
				Index: len(parsed.Lines),
				Text:  e.text,
			})
		}
		return parsed
	}

	hasWords := false
	for i, e := range timed {
		start := e.startMs + parsed.OffsetMs
		if start < 0 {
			start = 0
		}
		end := int64(0)
		if i+1 < len(timed) {
			end = timed[i+1].startMs + parsed.OffsetMs
		}
		if end <= start {
			end = start + 4000
		}
		line := domain.Line{
			Index:   len(parsed.Lines),
			Text:    e.text,
			StartMs: start,
			EndMs:   end,
		}
		if e.hasWord && len(e.words) > 0 {
			hasWords = true
			line.Words = shiftWords(e.words, parsed.OffsetMs, start, end)
		}
		parsed.Lines = append(parsed.Lines, line)
	}

	parsed.Sync = domain.SyncLine
	if hasWords {
		parsed.Sync = domain.SyncWord
	}
	return parsed
}

// shiftWords applies the document offset and clamps word timings inside the
// line so the renderer never sees an out-of-range span.
func shiftWords(words []domain.Word, offset, lineStart, lineEnd int64) []domain.Word {
	out := make([]domain.Word, 0, len(words))
	for _, w := range words {
		start := w.StartMs + offset
		end := w.EndMs + offset
		if start < lineStart {
			start = lineStart
		}
		if end > lineEnd {
			end = lineEnd
		}
		if end <= start {
			end = start + 150
		}
		out = append(out, domain.Word{Text: w.Text, StartMs: start, EndMs: end, TimingSrc: "provider"})
	}
	return out
}

// extractWords pulls <mm:ss.xx>word tags out of a line body.
func extractWords(body string) (words []domain.Word, stripped string, ok bool) {
	matches := wordTagRe.FindAllStringSubmatchIndex(body, -1)
	if len(matches) == 0 {
		return nil, body, false
	}
	stripped = strings.TrimSpace(wordTagRe.ReplaceAllString(body, ""))
	for i, m := range matches {
		start := tagMillis(body[m[2]:m[3]], body[m[4]:m[5]], group(body, m[6], m[7]))
		if start < 0 {
			continue
		}
		end := start + 400
		if i+1 < len(matches) {
			next := matches[i+1]
			nextStart := tagMillis(body[next[2]:next[3]], body[next[4]:next[5]], group(body, next[6], next[7]))
			if nextStart > start {
				end = nextStart
			}
		}
		text := strings.TrimSpace(body[m[1]:m0(matches, i, body)])
		if text == "" {
			continue
		}
		words = append(words, domain.Word{Text: text, StartMs: start, EndMs: end, TimingSrc: "provider"})
	}
	if len(words) == 0 {
		return nil, stripped, false
	}
	return words, stripped, true
}

// m0 returns the start of the next tag (or end of body), i.e. where the current
// word's text stops.
func m0(matches [][]int, i int, body string) int {
	if i+1 < len(matches) {
		return matches[i+1][0]
	}
	return len(body)
}

func group(s string, from, to int) string {
	if from < 0 || to < 0 || to > len(s) || from > to {
		return ""
	}
	return s[from:to]
}

// tagMillis converts a minutes/seconds/fraction triple to milliseconds.
func tagMillis(min, sec, frac string) int64 {
	m, err := strconv.Atoi(min)
	if err != nil {
		return -1
	}
	s, err := strconv.Atoi(sec)
	if err != nil {
		return -1
	}
	ms := int64(m)*60000 + int64(s)*1000
	if frac != "" {
		switch len(frac) {
		case 1:
			ms += int64(atoiSafe(frac)) * 100
		case 2:
			ms += int64(atoiSafe(frac)) * 10
		default:
			ms += int64(atoiSafe(frac[:3]))
		}
	}
	return ms
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// Format renders a human-readable description of a sync level.
func Format(level domain.SyncLevel) string {
	switch level {
	case domain.SyncWord:
		return "palabra por palabra"
	case domain.SyncSyllable:
		return "silaba por silaba"
	case domain.SyncLine:
		return "por verso"
	default:
		return "sin sincronizar"
	}
}

// Validate checks internal consistency, used by tests and by the importer.
func Validate(l domain.Lyrics) error {
	var prevEnd int64
	for i, line := range l.Lines {
		if line.StartMs < prevEnd {
			return fmt.Errorf("line %d starts before line %d ends", i, i-1)
		}
		if line.EndMs < line.StartMs {
			return fmt.Errorf("line %d ends before it starts", i)
		}
		for j, w := range line.Words {
			if w.StartMs < line.StartMs || w.EndMs > line.EndMs {
				return fmt.Errorf("word %d of line %d falls outside the line", j, i)
			}
		}
		prevEnd = line.EndMs
	}
	return nil
}
