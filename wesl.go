package wesl

import (
	"errors"
	"sync"

	"github.com/bluescreen10/wesl-go/parser"
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
	_, err := parser.Parse(src)
	return t, err
}

func (t *Translator) New(name string) *Translator {
	return &Translator{name: name, common: t.common}
}

func (t *Translator) Translate(src string, params map[string]bool) (string, error) {
	return "", errors.New("not implemented")
}
