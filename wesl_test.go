package wesl_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/bluescreen10/wesl-go"
)

func TestImportSyntax(t *testing.T) {
	var testCases []struct {
		Src      string `json:"src"`
		Expected bool   `json:"fails"`
	}

	r, err := os.Open("testdata/importSyntaxCases.json")
	if err != nil {
		t.Fatalf("error opening test cases: %v", err)
	}

	err = json.NewDecoder(r).Decode(&testCases)
	if err != nil {
		t.Fatalf("error parsing test cases: %v", err)
	}

	for _, test := range testCases {
		t.Run(test.Src, func(t *testing.T) {
			_, err := wesl.New("test").Parse(test.Src)
			got := err != nil

			if test.Expected != got {
				t.Errorf("parse (%s) expected (%v), got (%v)\n%v", test.Src, test.Expected, got, err)
			}
		})
	}
}

func TestImportCases(t *testing.T) {
	var testCases []struct {
		Name               string            `json:"name"`
		Srcs               map[string]string `json:"weslSrc"`
		Expected           string            `json:"expectedWgsl"`
		ExpectedUnderscore string            `json:"underscoreWgsl"`
	}

	r, err := os.Open("testdata/importCases.json")
	if err != nil {
		t.Fatalf("error opening test cases: %v", err)
	}

	err = json.NewDecoder(r).Decode(&testCases)
	if err != nil {
		t.Fatalf("error parsing test cases: %v", err)
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			var w *wesl.Translator

			for file, src := range test.Srcs {
				// skip main
				if file == "./main.wgsl" {
					continue
				}

				if w == nil {
					w, err = wesl.New(file).Parse(src)
					if err != nil {
						t.Errorf("error parsing file %s: %v", file, err)
					}
				} else {
					w, err = w.New(file).Parse(src)
					if err != nil {
						t.Errorf("error parsing file %s: %v", file, err)
					}
				}
			}

			src := test.Srcs["./main.wgsl"]
			got, err := w.Translate(src, nil)

			if err != nil {
				t.Errorf("translate failed %v", err)
			}

			if test.Expected != got {
				t.Errorf("translate (%s) expected (%s), got (%s)", src, test.Expected, got)
			}
		})
	}
}

func TestConditionalTranslation(t *testing.T) {
	var testCases []struct {
		Name               string            `json:"name"`
		Srcs               map[string]string `json:"weslSrc"`
		Expected           string            `json:"expectedWgsl"`
		ExpectedUnderscore string            `json:"underscoreWgsl"`
	}

	r, err := os.Open("testdata/conditionalTranslationCases.json")
	if err != nil {
		t.Fatalf("error opening test cases: %v", err)
	}

	err = json.NewDecoder(r).Decode(&testCases)
	if err != nil {
		t.Fatalf("error parsing test cases: %v", err)
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			var w *wesl.Translator

			for file, src := range test.Srcs {
				// skip main
				if file == "./main.wgsl" {
					continue
				}

				if w == nil {
					w, err = wesl.New(file).Parse(src)
					if err != nil {
						t.Errorf("error parsing file %s: %v", file, err)
					}
				} else {
					w, err = w.New(file).Parse(src)
					if err != nil {
						t.Errorf("error parsing file %s: %v", file, err)
					}
				}
			}

			src := test.Srcs["./main.wgsl"]
			got, err := w.Translate(src, nil)

			if err != nil {
				t.Errorf("translate failed %v", err)
			}

			if test.Expected != got {
				t.Errorf("translate (%s)\n  expected (%s)\n  got (%s)", src, test.Expected, got)
			}
		})
	}
}
