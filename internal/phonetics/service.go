// Package phonetics turns lyric lines into study material: IPA, a Spanish
// approximation and a translation. It validates every model answer against the
// lines it asked about, so invented or mismatched data never reaches the UI.
package phonetics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LuiisDev21/stefani/internal/ai"
	"github.com/LuiisDev21/stefani/internal/config"
	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/storage"
)

// Engine versions are part of the cache key: bumping any of them invalidates
// cached annotations on purpose.
const (
	PromptVersion = "phonetics-v1"
	FormatVersion = "annotations-v1"
	EngineVersion = "llm-ipa-v1"
	BatchSize     = 10
)

// ConfigFunc returns the current AI configuration and API key. Returning an
// error means the user has not configured a usable provider.
type ConfigFunc func() (config.LLM, string, error)

// Logger is the logging surface.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

// Service generates and caches annotations.
type Service struct {
	db     *storage.DB
	client *ai.Client
	cfg    ConfigFunc
	log    Logger
}

// NewService wires the phonetics service.
func NewService(db *storage.DB, client *ai.Client, cfg ConfigFunc, log Logger) *Service {
	return &Service{db: db, client: client, cfg: cfg, log: log}
}

// Progress reports annotation progress to the UI.
type Progress struct {
	Done  int
	Total int
}

// Result summarizes one annotation run.
type Result struct {
	Generated int
	Cached    int
	Failed    []string
}

// AnnotateTrack fills in missing annotations for a track. Lines already cached
// with the same engine, model and prompt version are skipped, so the call is
// cheap to repeat.
func (s *Service) AnnotateTrack(ctx context.Context, trackID, accent string, onProgress func(Progress)) (Result, error) {
	if accent == "" {
		accent = config.DefaultAccent
	}
	doc, err := s.db.Lyrics(ctx, trackID)
	if err != nil {
		return Result{}, err
	}
	if len(doc.Lines) == 0 {
		return Result{}, fmt.Errorf("this track has no lyrics to annotate")
	}
	llm, key, err := s.cfg()
	if err != nil {
		return Result{}, err
	}

	existing, err := s.db.AnnotationsByTrack(ctx, trackID)
	if err != nil {
		return Result{}, err
	}
	cached := make(map[string]domain.Annotation, len(existing))
	for _, a := range existing {
		cached[a.LineID] = a
	}

	var pending []domain.Line
	result := Result{}
	for _, line := range doc.Lines {
		if strings.TrimSpace(line.Text) == "" {
			continue
		}
		if a, ok := cached[line.ID]; ok && a.CacheKey == cacheKey(line.Text, accent, llm.Model) {
			result.Cached++
			continue
		}
		pending = append(pending, line)
	}
	if len(pending) == 0 {
		return result, nil
	}

	total := len(pending)
	done := 0
	report := func() {
		if onProgress != nil {
			onProgress(Progress{Done: done, Total: total})
		}
	}
	report()

	for start := 0; start < len(pending); start += BatchSize {
		end := start + BatchSize
		if end > len(pending) {
			end = len(pending)
		}
		batch := pending[start:end]
		items, failed, err := s.annotateBatch(ctx, llm, key, accent, batch)
		if err != nil {
			return result, err
		}
		for _, item := range items {
			if err := s.db.SaveAnnotation(ctx, item); err != nil {
				return result, err
			}
			result.Generated++
		}
		result.Failed = append(result.Failed, failed...)
		done += len(batch)
		report()
	}
	return result, nil
}

// annotateBatch asks for one batch and retries the rejected lines once with the
// validation errors attached. Whatever still fails is reported, never stored.
func (s *Service) annotateBatch(ctx context.Context, llm config.LLM, key, accent string, lines []domain.Line) ([]domain.Annotation, []string, error) {
	wanted := make(map[string]domain.Line, len(lines))
	for _, line := range lines {
		wanted[line.ID] = line
	}

	raw, err := s.client.ChatJSON(ctx, aiConfig(llm, key), systemPrompt(accent), userPrompt(lines))
	if err != nil {
		return nil, nil, err
	}
	accepted, rejected, feedback := validate(wanted, raw, accent, llm.Model)
	if len(rejected) == 0 {
		return accepted, nil, nil
	}

	// One repair attempt, restricted to the rejected lines.
	retryLines := make([]domain.Line, 0, len(rejected))
	for _, id := range rejected {
		retryLines = append(retryLines, wanted[id])
	}
	s.log.Warn("AI answer failed validation, retrying rejected lines",
		"lines", len(retryLines), "reason", strings.Join(feedback, "; "))

	rawRetry, err := s.client.ChatJSON(ctx, aiConfig(llm, key), systemPrompt(accent),
		userPrompt(retryLines)+"\n\nCorrige estos errores del intento anterior:\n- "+strings.Join(feedback, "\n- "))
	if err != nil {
		return accepted, rejected, nil
	}
	acceptedRetry, stillRejected, _ := validate(subset(wanted, retryLines), rawRetry, accent, llm.Model)
	accepted = append(accepted, acceptedRetry...)
	return accepted, stillRejected, nil
}

func subset(all map[string]domain.Line, lines []domain.Line) map[string]domain.Line {
	out := make(map[string]domain.Line, len(lines))
	for _, line := range lines {
		out[line.ID] = line
	}
	return out
}

// LookupWord returns dictionary data for one word in context. It is not stored
// until the user saves it.
func (s *Service) LookupWord(ctx context.Context, word, lineContext string) (domain.WordEntry, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return domain.WordEntry{}, fmt.Errorf("select a word first")
	}
	if len(word) > 64 {
		return domain.WordEntry{}, fmt.Errorf("that selection is too long for a word lookup")
	}
	llm, key, err := s.cfg()
	if err != nil {
		return domain.WordEntry{}, err
	}
	prompt := fmt.Sprintf(`Palabra: %q
Verso donde aparece: %q

Devuelve un objeto JSON con esta forma exacta:
{"word":"...","lemma":"...","part_of_speech":"sustantivo|verbo|adjetivo|adverbio|preposicion|pronombre|conjuncion|interjeccion|expresion","ipa":"/.../","translation":"...","meaning":"...","example":"..."}

Reglas:
- "ipa" es la transcripcion IPA estadounidense de la palabra aislada.
- "translation" es la traduccion al espanol dentro del verso.
- "meaning" explica el uso en ese contexto en una frase.
- "example" es una oracion corta en ingles que use la palabra (maximo 120 caracteres).`, word, lineContext)

	raw, err := s.client.ChatJSON(ctx, aiConfig(llm, key),
		`Eres un diccionario bilingue para estudiantes hispanohablantes de ingles. Respondes solo JSON valido.`,
		prompt)
	if err != nil {
		return domain.WordEntry{}, err
	}
	var parsed struct {
		Word         string `json:"word"`
		Lemma        string `json:"lemma"`
		PartOfSpeech string `json:"part_of_speech"`
		IPA          string `json:"ipa"`
		Translation  string `json:"translation"`
		Meaning      string `json:"meaning"`
		Example      string `json:"example"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return domain.WordEntry{}, fmt.Errorf("the AI answer was not valid JSON")
	}
	entry := domain.WordEntry{
		Word:         limit(clean(parsed.Word), 64),
		Lemma:        limit(clean(parsed.Lemma), 64),
		PartOfSpeech: limit(clean(parsed.PartOfSpeech), 32),
		IPA:          limit(clean(parsed.IPA), 120),
		Translation:  limit(clean(parsed.Translation), 240),
		Meaning:      limit(clean(parsed.Meaning), 400),
		Example:      limit(clean(parsed.Example), 240),
		Source:       "ai",
	}
	if entry.Word == "" {
		entry.Word = word
	}
	if entry.Translation == "" && entry.Meaning == "" {
		return domain.WordEntry{}, fmt.Errorf("the AI answer had no usable definition")
	}
	return entry, nil
}

// ResetCache drops generated annotations for a track.
func (s *Service) ResetCache(ctx context.Context, trackID string) error {
	return s.db.DeleteAnnotationsForTrack(ctx, trackID)
}

func aiConfig(llm config.LLM, key string) ai.Config {
	return ai.Config{
		BaseURL:     llm.BaseURL,
		Model:       llm.Model,
		APIKey:      key,
		Temperature: llm.Temperature,
		Timeout:     time.Duration(llm.TimeoutSeconds) * time.Second,
		MaxRetries:  llm.MaxRetries,
		JSONMode:    llm.JSONMode,
	}
}

// cacheKey identifies an annotation by everything that could change it.
func cacheKey(text, accent, model string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s|%s|%s",
		normalize(text), accent, EngineVersion, model, PromptVersion, FormatVersion)
	return hex.EncodeToString(h.Sum(nil))[:24]
}

func normalize(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

func limit(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max])
}

// clean removes control characters that would break the UI or the database.
func clean(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
}
