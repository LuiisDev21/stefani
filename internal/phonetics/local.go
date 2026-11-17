package phonetics

import (
	_ "embed"
	"regexp"
	"strings"
	"sync"
)

// Local transcription
//
// CMUdict gives American English pronunciations as ARPAbet phonemes with stress
// marks. Converting them to IPA and then to a Spanish reading is deterministic,
// works offline and needs no AI: that is what lets a freshly fetched lyric show
// pronunciation immediately, even before the user configures a provider.
//
// A language model is still better for translation and for notes about context;
// this engine only covers the sounds of each word.

//go:embed data/cmudict.dict
var cmuDictData []byte

// wordPattern extracts words, keeping internal apostrophes ("don't").
var wordPattern = regexp.MustCompile(`[A-Za-z]+(?:['’][A-Za-z]+)?`)

// LocalEngine transcribes English text using the bundled CMU dictionary.
type LocalEngine struct {
	once  sync.Once
	words map[string][]string
}

// NewLocal returns a lazily loaded local engine.
func NewLocal() *LocalEngine { return &LocalEngine{} }

// Loaded reports whether the dictionary has been parsed (useful for tests).
func (e *LocalEngine) Loaded() bool { return e != nil && e.words != nil }

func (e *LocalEngine) load() {
	e.once.Do(func() {
		e.words = make(map[string][]string, 140000)
		for _, line := range strings.Split(string(cmuDictData), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, ";;;") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			word := strings.ToLower(fields[0])
			// "word(2)" marks an alternative pronunciation: keep the first one.
			if cut := strings.IndexByte(word, '('); cut > 0 {
				word = word[:cut]
			}
			if _, exists := e.words[word]; exists {
				continue
			}
			e.words[word] = fields[1:]
		}
	})
}

// Known reports whether the dictionary has a pronunciation for a word.
func (e *LocalEngine) Known(word string) bool {
	if e == nil {
		return false
	}
	e.load()
	_, ok := e.words[normalizeWord(word)]
	return ok
}

// Transcribe returns the IPA and the Spanish approximation for a phrase, plus
// the fraction of words the dictionary knew. Unknown words are left as written
// so the reader can see they were not transcribed.
func (e *LocalEngine) Transcribe(text string) (ipa string, spanish string, coverage float64) {
	if e == nil {
		return "", "", 0
	}
	e.load()
	matches := wordPattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return "", "", 0
	}
	ipaParts := make([]string, 0, len(matches))
	esParts := make([]string, 0, len(matches))
	known := 0
	for _, raw := range matches {
		word := normalizeWord(raw)
		phones, ok := e.words[word]
		if !ok {
			ipaParts = append(ipaParts, raw)
			esParts = append(esParts, raw)
			continue
		}
		known++
		wordIPA := arpabetIPA(phones)
		ipaParts = append(ipaParts, wordIPA)
		esParts = append(esParts, ipaToSpanish(wordIPA, word))
	}
	if known == 0 {
		return "", "", 0
	}
	return "/" + strings.Join(ipaParts, " ") + "/", strings.Join(esParts, " "), float64(known) / float64(len(matches))
}

// normalizeWord lowercases and strips the apostrophes the dictionary does not
// use, keeping internal ones ("don't" is in CMUdict as "don't").
func normalizeWord(word string) string {
	word = strings.ToLower(strings.TrimSpace(word))
	word = strings.Trim(word, "'’")
	return strings.ReplaceAll(word, "’", "'")
}

// phonemeUnits maps ARPAbet to IPA, including the reduced vowels.
type phonemeUnit struct {
	symbol string
	vowel  bool
	stress byte
}

var arpabetSymbols = map[string]string{
	"AA": "ɑ", "AE": "æ", "AH": "ʌ", "AO": "ɔ", "AW": "aʊ", "AY": "aɪ",
	"B": "b", "CH": "tʃ", "D": "d", "DH": "ð", "EH": "ɛ", "ER": "ɝ",
	"EY": "eɪ", "F": "f", "G": "ɡ", "HH": "h", "IH": "ɪ", "IY": "i",
	"JH": "dʒ", "K": "k", "L": "l", "M": "m", "N": "n", "NG": "ŋ",
	"OW": "oʊ", "OY": "ɔɪ", "P": "p", "R": "ɹ", "S": "s", "SH": "ʃ",
	"T": "t", "TH": "θ", "UH": "ʊ", "UW": "u", "V": "v", "W": "w",
	"Y": "j", "Z": "z", "ZH": "ʒ",
}

var vowelPhonemes = map[string]bool{
	"AA": true, "AE": true, "AH": true, "AO": true, "AW": true, "AY": true,
	"EH": true, "ER": true, "EY": true, "IH": true, "IY": true, "OW": true,
	"OY": true, "UH": true, "UW": true,
}

// arpabetIPA renders a pronunciation, placing the stress mark at the start of
// the stressed syllable (a small onset heuristic: up to two preceding
// consonants, never "ŋ").
func arpabetIPA(phones []string) string {
	units := make([]phonemeUnit, 0, len(phones))
	for _, phone := range phones {
		base := phone
		var stress byte
		if n := len(phone); n > 1 {
			switch phone[n-1] {
			case '0', '1', '2':
				stress = phone[n-1]
				base = phone[:n-1]
			}
		}
		symbol, ok := arpabetSymbols[base]
		if !ok {
			continue
		}
		if vowelPhonemes[base] {
			switch base {
			case "AH":
				if stress == '0' || stress == 0 {
					symbol = "ə"
				}
			case "ER":
				if stress == '0' || stress == 0 {
					symbol = "ɚ"
				}
			}
		}
		units = append(units, phonemeUnit{symbol: symbol, vowel: vowelPhonemes[base], stress: stress})
	}

	// A one-syllable word needs no stress mark in a phrase reading: English
	// marks every vowel in the lexicon, but "it" or "take" are read unstressed
	// next to the content word that carries the beat.
	vowels := 0
	for _, unit := range units {
		if unit.vowel {
			vowels++
		}
	}

	var out strings.Builder
	for i, unit := range units {
		if vowels > 1 && (unit.stress == '1' || unit.stress == '2') {
			mark := "ˌ"
			if unit.stress == '1' {
				mark = "ˈ"
			}
			pos := onsetStart(units, i)
			// Insert the mark right before the syllable onset: write the
			// consonants that belong to it again is not needed because the loop
			// already wrote them, so emit at the recorded position instead.
			if pos == i {
				out.WriteString(mark)
			} else {
				// Rewrite the tail: simpler to rebuild the suffix.
				s := out.String()
				keep := len(s)
				for k := pos; k < i; k++ {
					keep -= len(units[k].symbol)
				}
				out.Reset()
				out.WriteString(s[:keep])
				out.WriteString(mark)
				for k := pos; k < i; k++ {
					out.WriteString(units[k].symbol)
				}
			}
		}
		out.WriteString(unit.symbol)
	}
	return out.String()
}

// onsetStart walks back over the consonants that open a stressed syllable.
func onsetStart(units []phonemeUnit, vowelIndex int) int {
	pos := vowelIndex
	// Up to three consonants: English allows onsets like "str" (strangers).
	for steps := 0; steps < 3 && pos > 0; steps++ {
		prev := units[pos-1]
		if prev.vowel || prev.symbol == "ŋ" {
			break
		}
		pos--
	}
	return pos
}

var spanishVowels = map[rune]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true}

var accented = map[rune]rune{'a': 'á', 'e': 'é', 'i': 'í', 'o': 'ó', 'u': 'ú'}

// ipaToSpanish rewrites our own IPA output with Spanish spelling conventions.
// Longest sequences are tried first so diphthongs are not split.
var spanishRewrites = []struct{ ipa, es string }{
	{"tʃ", "ch"}, {"dʒ", "y"}, {"aɪ", "ai"}, {"aʊ", "au"}, {"ɔɪ", "oi"},
	{"eɪ", "ei"}, {"oʊ", "ou"}, {"ɝ", "er"}, {"ɚ", "er"}, {"ə", "a"},
	{"ʌ", "a"}, {"æ", "a"}, {"ɑ", "a"}, {"ɔ", "o"}, {"ɛ", "e"},
	{"ɪ", "i"}, {"i", "i"}, {"ʊ", "u"}, {"u", "u"}, {"ʃ", "sh"},
	{"ʒ", "y"}, {"θ", "z"}, {"ð", "d"}, {"ŋ", "ng"}, {"ɹ", "r"},
	{"j", "y"}, {"w", "u"}, {"ɡ", "g"}, {"v", "b"}, {"b", "b"},
	{"d", "d"}, {"f", "f"}, {"h", "j"}, {"k", "k"}, {"l", "l"},
	{"m", "m"}, {"n", "n"}, {"p", "p"}, {"s", "s"}, {"t", "t"}, {"z", "s"},
}

// functionWords are read without stress in a phrase, so they get no written
// accent: "it" stays "it", not "ít".
var functionWords = map[string]bool{
	"a": true, "an": true, "the": true, "it": true, "its": true, "of": true,
	"to": true, "in": true, "on": true, "at": true, "and": true, "or": true,
	"but": true, "as": true, "is": true, "was": true, "are": true, "were": true,
	"be": true, "been": true, "am": true, "do": true, "does": true, "did": true,
	"has": true, "have": true, "had": true, "can": true, "could": true,
	"will": true, "would": true, "shall": true, "should": true, "may": true,
	"might": true, "must": true, "my": true, "your": true, "his": true,
	"her": true, "our": true, "their": true, "them": true, "him": true,
	"us": true, "me": true, "this": true, "that": true, "these": true,
	"those": true, "for": true, "from": true, "with": true, "by": true,
	"than": true, "then": true, "so": true, "if": true, "not": true,
	"up": true, "out": true, "off": true, "we": true, "you": true, "he": true,
	"she": true, "they": true, "i": true, "i'm": true, "it's": true,
	"don't": true, "can't": true, "won't": true,
}

func ipaToSpanish(word, source string) string {
	var out strings.Builder
	pendingStress := false
	hadStress := strings.Contains(word, "ˈ")
	for i := 0; i < len(word); {
		if strings.HasPrefix(word[i:], "ˈ") {
			pendingStress = true
			i += len("ˈ")
			continue
		}
		if strings.HasPrefix(word[i:], "ˌ") {
			pendingStress = false
			i += len("ˌ")
			continue
		}
		matched := false
		for _, pair := range spanishRewrites {
			if strings.HasPrefix(word[i:], pair.ipa) {
				for _, r := range pair.es {
					if pendingStress && spanishVowels[r] && accented[r] != 0 {
						out.WriteRune(accented[r])
						pendingStress = false
						continue
					}
					out.WriteRune(r)
				}
				i += len(pair.ipa)
				matched = true
				break
			}
		}
		if !matched {
			i++
		}
	}
	result := out.String()
	// Monosyllables carry the beat but our IPA output has no stress mark for
	// them: give content words their written accent, leave function words bare.
	if !hadStress && !functionWords[source] {
		result = accentFirstVowel(result)
	}
	return result
}

// accentFirstVowel writes the accent on the first vowel of a Spanish reading.
func accentFirstVowel(word string) string {
	runes := []rune(word)
	for i, r := range runes {
		if accented[r] != 0 {
			runes[i] = accented[r]
			return string(runes)
		}
	}
	return word
}

// Line is the per-line helper the application layer uses: it only reports a
// transcription when enough words were known to be trustworthy.
func (e *LocalEngine) Line(text string, minCoverage float64) (ipa, spanish string, ok bool) {
	ipa, spanish, coverage := e.Transcribe(text)
	if coverage < minCoverage || spanish == "" {
		return "", "", false
	}
	return ipa, spanish, true
}
