package phonetics

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// systemPrompt states the contract once for every request. It is deliberately
// narrow: the model may only annotate the lines it receives.
func systemPrompt(accent string) string {
	return fmt.Sprintf(`Eres un profesor de ingles para estudiantes hispanohablantes y generas material de estudio para versos de canciones.

Trabajas con la variedad %s.

Reglas obligatorias:
1. Solo puedes anotar los versos que recibes. No inventes versos, no agregues lineId desconocidos, no omitas lineId.
2. "original" debe copiar el texto recibido sin cambios, sin corregir la ortografia.
3. "ipa" es la transcripcion IPA estadounidense, entre barras, con acentos primarios (ˈ) y secundarios (ˌ).
4. "spanishApprox" es una guia pedagogica para hispanohablantes: silabas separadas por espacios, acento escrito con tilde sobre la vocal fuerte, sin inventar sonidos que el ingles no tiene.
5. "translation" traduce el sentido del verso completo al espanol neutro, no palabra por palabra.
6. "notes" incluye como maximo 2 notas utiles (tipo: "stress", "link", "reduction", "sound") sobre sonidos dificiles, enlaces entre palabras o reducciones.
7. No afirmes como pronuncia el cantante: describes la pronunciacion linguistica del texto.
8. Respondes unicamente con JSON valido, sin markdown.`, accent)
}

// userPrompt renders the request with the exact line set.
func userPrompt(lines []domain.Line) string {
	var b strings.Builder
	b.WriteString("Anota exactamente estos versos.\n\nFormato de respuesta:\n")
	b.WriteString(`{"lines":[{"lineId":"...","language":"en-US","original":"...","ipa":"/.../","spanishApprox":"...","translation":"...","notes":[{"type":"stress","text":"..."}]}]}`)
	b.WriteString("\n\nVersos:\n")
	for _, line := range lines {
		fmt.Fprintf(&b, "- lineId=%q original=%q\n", line.ID, line.Text)
	}
	fmt.Fprintf(&b, "\nDevuelve exactamente %d objetos, uno por lineId, en el mismo orden.", len(lines))
	return b.String()
}

type answerLine struct {
	LineID        string        `json:"lineId"`
	Language      string        `json:"language"`
	Original      string        `json:"original"`
	IPA           string        `json:"ipa"`
	SpanishApprox string        `json:"spanishApprox"`
	Translation   string        `json:"translation"`
	Notes         []domain.Note `json:"notes"`
}

type answerDoc struct {
	Lines []answerLine `json:"lines"`
}

// validate accepts only annotations that match a requested line and preserve
// its text. Everything else is rejected with a reason the repair prompt uses.
func validate(wanted map[string]domain.Line, raw []byte, accent, model string) (
	accepted []domain.Annotation, rejected []string, feedback []string) {

	var doc answerDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		ids := make([]string, 0, len(wanted))
		for id := range wanted {
			ids = append(ids, id)
		}
		return nil, ids, []string{"la respuesta no era JSON valido con la forma esperada"}
	}

	seen := make(map[string]bool, len(doc.Lines))
	for _, item := range doc.Lines {
		line, ok := wanted[item.LineID]
		if !ok {
			feedback = append(feedback, fmt.Sprintf("lineId desconocido %q: no estaba en la peticion", item.LineID))
			continue
		}
		if seen[item.LineID] {
			feedback = append(feedback, fmt.Sprintf("lineId repetido %q", item.LineID))
			continue
		}
		seen[item.LineID] = true

		if normalize(item.Original) != normalize(line.Text) {
			rejected = append(rejected, item.LineID)
			feedback = append(feedback, fmt.Sprintf("el verso %q no coincide con el original %q", item.Original, line.Text))
			continue
		}
		ipa := normalizeIPA(item.IPA)
		if ipa == "" {
			rejected = append(rejected, item.LineID)
			feedback = append(feedback, fmt.Sprintf("falta la transcripcion IPA del verso %q", line.Text))
			continue
		}
		approx := clean(item.SpanishApprox)
		translation := clean(item.Translation)
		if approx == "" || translation == "" {
			rejected = append(rejected, item.LineID)
			feedback = append(feedback, fmt.Sprintf("faltan la aproximacion o la traduccion del verso %q", line.Text))
			continue
		}
		notes := make([]domain.Note, 0, len(item.Notes))
		for _, note := range item.Notes {
			text := limit(clean(note.Text), 240)
			if text == "" {
				continue
			}
			notes = append(notes, domain.Note{Type: limit(clean(note.Type), 24), Text: text})
			if len(notes) == 2 {
				break
			}
		}
		accepted = append(accepted, domain.Annotation{
			LineID:        item.LineID,
			Accent:        accent,
			IPA:           limit(ipa, 400),
			SpanishApprox: limit(approx, 400),
			Translation:   limit(translation, 600),
			Notes:         notes,
			Model:         model,
			Version:       EngineVersion,
			CacheKey:      cacheKey(line.Text, accent, model),
		})
	}

	for id, line := range wanted {
		if !seen[id] {
			rejected = append(rejected, id)
			feedback = append(feedback, fmt.Sprintf("falta el verso %q (lineId=%s)", line.Text, id))
		}
	}
	return accepted, rejected, feedback
}

// normalizeIPA keeps a single slash convention so the UI and cache agree.
func normalizeIPA(s string) string {
	s = clean(s)
	if s == "" {
		return ""
	}
	s = strings.Trim(s, "/ ")
	if !strings.ContainsAny(s, "ˈˌəɪʊɛɔæʌθðʃʒŋɹɑɜ") {
		// No IPA characters at all means the model returned plain spelling.
		return ""
	}
	return "/" + s + "/"
}
