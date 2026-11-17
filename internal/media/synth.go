// Package media renders audio and streams files to the webview.
//
// STEFANI needs playable music that ships with the app and carries no licence
// baggage. Instead of bundling third-party recordings, the demo library is
// synthesized from a deterministic score: the same score produces the word
// timings the lyric engine uses, so audio and text are aligned by construction.
package media

import (
	"bytes"
	"encoding/binary"
	"math"
)

// Timbre selects the harmonic recipe of a note.
type Timbre int

const (
	TimbreMelody Timbre = iota
	TimbrePad
	TimbreKick
	TimbreHat
)

// Note is one scheduled sound.
type Note struct {
	StartSec float64
	DurSec   float64
	Freq     float64
	Amp      float64
	Timbre   Timbre
}

// Spec is a complete score to render.
type Spec struct {
	SampleRate int
	Duration   float64
	Notes      []Note
}

// RenderWAV mixes the score into a 16-bit mono PCM WAV file.
func RenderWAV(spec Spec) []byte {
	rate := spec.SampleRate
	if rate <= 0 {
		rate = 44100
	}
	total := int(spec.Duration * float64(rate))
	if total <= 0 {
		total = rate
	}
	buf := make([]float64, total)
	for _, n := range spec.Notes {
		mixNote(buf, rate, n)
	}

	samples := make([]int16, total)
	peak := 0.0
	for i, v := range buf {
		if v > peak {
			peak = v
		} else if -v > peak {
			peak = -v
		}
		_ = i
	}
	gain := 1.0
	if peak > 0.89 {
		gain = 0.89 / peak
	}
	for i, v := range buf {
		scaled := v * gain
		if scaled > 1 {
			scaled = 1
		} else if scaled < -1 {
			scaled = -1
		}
		samples[i] = int16(math.Round(scaled * 32767))
	}
	return encodeWAV(samples, rate)
}

func mixNote(buf []float64, rate int, n Note) {
	start := int(n.StartSec * float64(rate))
	length := int(n.DurSec * float64(rate))
	if start < 0 {
		start = 0
	}
	if length <= 0 || start >= len(buf) {
		return
	}
	if start+length > len(buf) {
		length = len(buf) - start
	}
	// A cheap deterministic noise source: hats must not depend on global state
	// and must sound identical on every run.
	noiseState := uint32(2166136261) ^ uint32(start)
	nextNoise := func() float64 {
		noiseState = noiseState*1664525 + 1013904223
		return (float64(noiseState>>9)/float64(1<<23))*2 - 1
	}

	for i := 0; i < length; i++ {
		t := float64(i) / float64(rate)
		progress := float64(i) / float64(length)
		var v float64
		switch n.Timbre {
		case TimbrePad:
			env := envelope(progress, 0.22, 0.18, 0.75)
			v = (math.Sin(2*math.Pi*n.Freq*t) +
				0.42*math.Sin(2*math.Pi*n.Freq*2*t+0.3) +
				0.18*math.Sin(2*math.Pi*n.Freq*3*t)) * env
		case TimbreKick:
			// Pitch drop gives a soft, rounded thump without samples.
			f := 105 - 60*progress
			env := math.Exp(-6.5 * progress)
			v = math.Sin(2*math.Pi*f*t) * env
		case TimbreHat:
			env := math.Exp(-14 * progress)
			v = nextNoise() * env
		default:
			env := envelope(progress, 0.02, 0.10, 0.55)
			vibrato := 1 + 0.004*math.Sin(2*math.Pi*5.2*t)
			v = (math.Sin(2*math.Pi*n.Freq*vibrato*t) +
				0.30*math.Sin(2*math.Pi*n.Freq*2*t) +
				0.10*math.Sin(2*math.Pi*n.Freq*4*t)) * env
		}
		buf[start+i] += v * n.Amp
	}
}

// envelope returns attack/sustain/release shaped gain in [0,1].
func envelope(p, attack, release, sustain float64) float64 {
	switch {
	case p < attack:
		return p / attack * (1 - sustain + sustain)
	case p > 1-release:
		return sustain * (1 - (p-(1-release))/release)
	default:
		return 1 - (1-sustain)*0.35
	}
}

func encodeWAV(samples []int16, rate int) []byte {
	var buf bytes.Buffer
	dataSize := len(samples) * 2
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))     // PCM chunk size
	binary.Write(&buf, binary.LittleEndian, uint16(1))      // PCM format
	binary.Write(&buf, binary.LittleEndian, uint16(1))      // mono
	binary.Write(&buf, binary.LittleEndian, uint32(rate))   // sample rate
	binary.Write(&buf, binary.LittleEndian, uint32(rate*2)) // byte rate
	binary.Write(&buf, binary.LittleEndian, uint16(2))      // block align
	binary.Write(&buf, binary.LittleEndian, uint16(16))     // bits per sample
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataSize))
	out := make([]byte, dataSize)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(s))
	}
	buf.Write(out)
	return buf.Bytes()
}
