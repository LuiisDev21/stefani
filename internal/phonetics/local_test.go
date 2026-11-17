package phonetics

import (
	"strings"
	"testing"
)

func TestLocalTranscribesCommonWords(t *testing.T) {
	engine := NewLocal()
	ipa, spanish, coverage := engine.Transcribe("Take it slowly")
	if coverage != 1 {
		t.Fatalf("coverage = %v, want 1", coverage)
	}
	if ipa != "/teɪk ɪt ˈsloʊli/" {
		t.Fatalf("ipa = %q", ipa)
	}
	if spanish != "téik it slóuli" {
		t.Fatalf("spanish = %q", spanish)
	}
}

func TestLocalKeepsUnknownWordsVisible(t *testing.T) {
	engine := NewLocal()
	ipa, spanish, coverage := engine.Transcribe("Take the zzzqx")
	if coverage <= 0 || coverage >= 1 {
		t.Fatalf("coverage = %v", coverage)
	}
	if !strings.Contains(spanish, "zzzqx") || !strings.Contains(ipa, "zzzqx") {
		t.Fatalf("unknown word should stay visible: %q / %q", ipa, spanish)
	}
}

func TestLocalLineRespectsCoverage(t *testing.T) {
	engine := NewLocal()
	if _, _, ok := engine.Line("Take it slowly", 0.6); !ok {
		t.Fatal("a known line should pass the coverage check")
	}
	if _, _, ok := engine.Line("zzzqx qqqzz wwwwk", 0.6); ok {
		t.Fatal("an unknown line must not produce a transcription")
	}
}

func TestLocalHandlesTrickySpellings(t *testing.T) {
	engine := NewLocal()
	cases := map[string]string{
		"thought": "zót",  // th → z, ough
		"through": "zrú",  // th + r, one long vowel
		"voice":   "bóis", // v → b, oy
		"city":    "síti", // two syllables: the stress mark lands on the first
		"night":   "náit", // silent gh
	}
	for word, want := range cases {
		_, spanish, coverage := engine.Transcribe(word)
		if coverage != 1 {
			t.Fatalf("%s: coverage = %v", word, coverage)
		}
		if spanish != want {
			t.Fatalf("%s: spanish = %q, want %q", word, spanish, want)
		}
	}
}

func TestLocalNormalizesCaseAndApostrophes(t *testing.T) {
	engine := NewLocal()
	if !engine.Known("Don't") || !engine.Known("don't") {
		t.Fatal("apostrophes and case must not matter")
	}
	if !engine.Known("I'm") {
		t.Fatal("contractions should be in the dictionary")
	}
}
