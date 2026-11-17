package demo

// The demo library is original material written for STEFANI and released with
// the project. Nothing here is copied from a recording: the words are new, the
// pronunciation data was authored for teaching, and the audio is synthesized,
// so word timings come from the same score that produces the sound.

// Word is one sung token and its length in beats.
type Word struct {
	Text  string
	Beats float64
}

// Line is one verse with its study data.
type Line struct {
	Text        string
	Words       []Word
	IPA         string
	Approx      string
	Translation string
	Notes       []string
}

// Song is a demo track.
type Song struct {
	Slug   string
	Title  string
	Artist string
	Album  string
	BPM    float64
	// Root is the MIDI note the melody is built around.
	Root int
	// Mode is the pitch-class offset pattern used for the melody.
	Mode []int
	// Progression lists the chord roots (scale degrees) per bar.
	Progression []int
	Lines       []Line
}

func w(text string, beats float64) Word { return Word{Text: text, Beats: beats} }

// Songs is the bundled demo catalog, in presentation order.
var Songs = []Song{
	{
		Slug:        "midnight-signals",
		Title:       "Midnight Signals",
		Artist:      "Stefani Studio",
		Album:       "Demo Sessions",
		BPM:         92,
		Root:        57, // A3
		Mode:        []int{0, 2, 3, 5, 7, 8, 10},
		Progression: []int{0, 5, 3, 4},
		Lines: []Line{
			{
				Text:        "The city hums under my feet",
				Words:       []Word{w("The", 0.5), w("city", 1), w("hums", 1), w("under", 1), w("my", 0.5), w("feet", 1.5)},
				IPA:         "/ðə ˈsɪti hʌmz ˈʌndər maɪ fiːt/",
				Approx:      "də síri jams ánder mai fiit",
				Translation: "La ciudad zumba bajo mis pies",
				Notes:       []string{"La 'th' de 'the' es suave y sonora: lengua entre los dientes, sin vibrar como la 'd' española.", "La 'y' final de 'city' suena como una 'i' corta y relajada, no como 'i' fuerte."},
			},
			{
				Text:        "Neon rivers down the street",
				Words:       []Word{w("Neon", 0.8), w("rivers", 1.2), w("down", 1), w("the", 0.5), w("street", 1.5)},
				IPA:         "/ˈniːɒn ˈrɪvərz daʊn ðə striːt/",
				Approx:      "níion rívers dáun də striit",
				Translation: "Ríos de neón calle abajo",
			},
			{
				Text:        "I keep walking through the noise",
				Words:       []Word{w("I", 0.5), w("keep", 1), w("walking", 1.2), w("through", 1), w("the", 0.5), w("noise", 1.5)},
				IPA:         "/aɪ kiːp ˈwɔːkɪŋ θruː ðə nɔɪz/",
				Approx:      "ái kiip uóokin zruu də nóis",
				Translation: "Sigo caminando entre el ruido",
				Notes:       []string{"'through' empieza con el sonido 'th' sordo: solo aire, sin voz."},
			},
			{
				Text:        "Looking for a friendly voice",
				Words:       []Word{w("Looking", 1.2), w("for", 0.8), w("a", 0.5), w("friendly", 1.2), w("voice", 1.5)},
				IPA:         "/ˈlʊkɪŋ fɔːr ə ˈfrendli vɔɪs/",
				Approx:      "lúkin foor a fréndli vóis",
				Translation: "Buscando una voz amable",
			},
			{
				Text:        "Every window holds a light",
				Words:       []Word{w("Every", 0.8), w("window", 1.2), w("holds", 1), w("a", 0.5), w("light", 1.5)},
				IPA:         "/ˈevri ˈwɪndoʊ hoʊldz ə laɪt/",
				Approx:      "évri uíndou joulds a láit",
				Translation: "Cada ventana guarda una luz",
				Notes:       []string{"'window' lleva el acento en la primera sílaba: UÍN-dou."},
			},
			{
				Text:        "Someone's waking up tonight",
				Words:       []Word{w("Someone's", 1.2), w("waking", 1), w("up", 0.5), w("tonight", 1.5)},
				IPA:         "/ˈsʌmwʌnz ˈweɪkɪŋ ʌp təˈnaɪt/",
				Approx:      "sám-uans uéikin ap tu-náit",
				Translation: "Alguien despierta esta noche",
			},
		},
	},
	{
		Slug:        "paper-boats",
		Title:       "Paper Boats",
		Artist:      "Stefani Studio",
		Album:       "Demo Sessions",
		BPM:         84,
		Root:        55, // G3
		Mode:        []int{0, 2, 4, 5, 7, 9, 11},
		Progression: []int{0, 3, 5, 4},
		Lines: []Line{
			{
				Text:        "We folded words into a boat",
				Words:       []Word{w("We", 0.5), w("folded", 1.2), w("words", 1), w("into", 0.8), w("a", 0.5), w("boat", 1.5)},
				IPA:         "/wiː ˈfoʊldɪd wɜːrdz ˈɪntuː ə boʊt/",
				Approx:      "uii fóuldid uérds íntu a bóut",
				Translation: "Doblamos palabras en un barco",
				Notes:       []string{"'words' no lleva 'o': suena 'uérds', con la vocal de 'bird'."},
			},
			{
				Text:        "Set it down and let it float",
				Words:       []Word{w("Set", 0.5), w("it", 0.5), w("down", 1), w("and", 0.5), w("let", 0.5), w("it", 0.5), w("float", 1.5)},
				IPA:         "/set ɪt daʊn ænd let ɪt floʊt/",
				Approx:      "set it dáun and let it flóut",
				Translation: "Déjalo bajar y deja que flote",
			},
			{
				Text:        "Every wave can carry news",
				Words:       []Word{w("Every", 0.8), w("wave", 1), w("can", 0.5), w("carry", 1.2), w("news", 1.5)},
				IPA:         "/ˈevri weɪv kæn ˈkæri nuːz/",
				Approx:      "évri uéiv can kári niuz",
				Translation: "Cada ola puede llevar noticias",
			},
			{
				Text:        "Every river has a muse",
				Words:       []Word{w("Every", 0.8), w("river", 1.2), w("has", 0.5), w("a", 0.5), w("muse", 1.5)},
				IPA:         "/ˈevri ˈrɪvər hæz ə mjuːz/",
				Approx:      "évri ríver jas a miuz",
				Translation: "Cada río tiene una musa",
			},
			{
				Text:        "If it sinks I'll build again",
				Words:       []Word{w("If", 0.5), w("it", 0.5), w("sinks", 1), w("I'll", 0.5), w("build", 1), w("again", 1.5)},
				IPA:         "/ɪf ɪt sɪŋks aɪl bɪld əˈɡen/",
				Approx:      "if it sinks áil bild a-guén",
				Translation: "Si se hunde, construiré otro",
				Notes:       []string{"'again' suele sonar 'a-GUEN', con el acento al final."},
			},
			{
				Text:        "Paper remembers what I pen",
				Words:       []Word{w("Paper", 0.8), w("remembers", 1.5), w("what", 0.5), w("I", 0.5), w("pen", 1.2)},
				IPA:         "/ˈpeɪpər rɪˈmembərz wʌt aɪ pen/",
				Approx:      "péiper ri-mémbers uat ái pen",
				Translation: "El papel recuerda lo que escribo",
			},
		},
	},
	{
		Slug:        "slow-morning",
		Title:       "Slow Morning",
		Artist:      "Stefani Studio",
		Album:       "Demo Sessions",
		BPM:         72,
		Root:        60, // C4
		Mode:        []int{0, 2, 4, 5, 7, 9, 11},
		Progression: []int{0, 4, 5, 3},
		Lines: []Line{
			{
				Text:        "Coffee steams against the glass",
				Words:       []Word{w("Coffee", 0.8), w("steams", 1), w("against", 1.2), w("the", 0.5), w("glass", 1.5)},
				IPA:         "/ˈkɔːfi stiːmz əˈɡenst ðə ɡlæs/",
				Approx:      "kóofi stiims a-guénst də glas",
				Translation: "El café humea contra el cristal",
				Notes:       []string{"'glass' usa una 'a' abierta, entre 'a' y 'e' española."},
			},
			{
				Text:        "Minutes move but do not pass",
				Words:       []Word{w("Minutes", 1), w("move", 1), w("but", 0.5), w("do", 0.5), w("not", 0.5), w("pass", 1.5)},
				IPA:         "/ˈmɪnɪts muːv bʌt duː nɒt pæs/",
				Approx:      "mínits muuv bat duu not pas",
				Translation: "Los minutos avanzan pero no pasan",
			},
			{
				Text:        "Sunlight leans across the floor",
				Words:       []Word{w("Sunlight", 1), w("leans", 1), w("across", 1.2), w("the", 0.5), w("floor", 1.5)},
				IPA:         "/ˈsʌnlaɪt liːnz əˈkrɔːs ðə flɔːr/",
				Approx:      "sánlait liinz a-króos də floor",
				Translation: "La luz del sol se inclina por el suelo",
			},
			{
				Text:        "I don't need to rush anymore",
				Words:       []Word{w("I", 0.5), w("don't", 0.5), w("need", 1), w("to", 0.5), w("rush", 1), w("anymore", 1.5)},
				IPA:         "/aɪ doʊnt niːd tə rʌʃ ˌeniˈmɔːr/",
				Approx:      "ái dóunt niid tu rash e-ni-móor",
				Translation: "Ya no necesito apurarme",
				Notes:       []string{"'don't' reduce la vocal a 'ou' y la 't' casi no se oye."},
			},
			{
				Text:        "Let the morning take its time",
				Words:       []Word{w("Let", 0.5), w("the", 0.5), w("morning", 1.2), w("take", 1), w("its", 0.5), w("time", 1.5)},
				IPA:         "/let ðə ˈmɔːrnɪŋ teɪk ɪts taɪm/",
				Approx:      "let də mórnin téik its táim",
				Translation: "Deja que la mañana tome su tiempo",
			},
			{
				Text:        "Every quiet hour is mine",
				Words:       []Word{w("Every", 0.8), w("quiet", 1), w("hour", 1), w("is", 0.5), w("mine", 1.5)},
				IPA:         "/ˈevri ˈkwaɪət ˈaʊər ɪz maɪn/",
				Approx:      "évri kuáiet áuer is máin",
				Translation: "Cada hora tranquila es mía",
				Notes:       []string{"'hour' empieza con vocal: la 'h' no se pronuncia."},
			},
		},
	},
}
