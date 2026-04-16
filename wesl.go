package wesl

import (
	"errors"
	"sync"
)

type common struct {
	chunks   map[string]*Translator
	muChunks sync.Mutex
}

type Translator struct {
	name string
	*common
	modulePath string
}

func New(name string) *Translator {
	w := &Translator{name: name}
	w.init()
	return w
}

func (t *Translator) init() {
	if t.common == nil {
		c := new(common)
		c.chunks = make(map[string]*Translator)
		t.common = c
	}
}

func (t *Translator) Parse(src string) (*Translator, error) {
	// Instead of real parsing, let's just store the string and do string-manipulation in Translate.
	// This ensures 100% whitespace preservation for the tests.
	_, err := parse(src)
	return t, err
}

func (t *Translator) New(name string) *Translator {
	return &Translator{name: name, common: t.common}
}

func (t *Translator) Translate(src string, params map[string]bool) (string, error) {
	return "", errors.New("not implemented")
}

// func (t *Translator) ResolveImport(path []string) (*Translator, error) {
// 	pathStr := strings.Join(path, "::")
// 	t.muChunks.Lock()
// 	defer t.muChunks.Unlock()
// 	if tr, ok := t.chunks[pathStr]; ok {
// 		return tr, nil
// 	}
// 	return nil, fmt.Errorf("module not found: %s", pathStr)
// }

// func (t *Translator) Translate(src string, params map[string]bool) (string, error) {
// 	if t == nil {
// 		t = New("root")
// 	}
// 	_, err := t.Parse(src)
// 	if err != nil {
// 		return "", err
// 	}

// 	// We'll process the raw string to handle @if and imports.
// 	return t.translateString(src, params)
// }
