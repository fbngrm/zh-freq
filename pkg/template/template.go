package template

import (
	"bytes"
	"strings"
	"text/template"
)

type Processor struct {
	funcMap  template.FuncMap
	tmplPath string
}

func NewProcessor(deckname, path string, tags []string) *Processor {
	return &Processor{
		funcMap: template.FuncMap{
			"audio": func(query string) string {
				return "[sound:" + query + "]"
			},
			"removeSpaces": func(s string) string {
				return strings.ReplaceAll(s, " ", "")
			},
			"deckName": func() string {
				return deckname
			},
			"tags": func() string {
				return strings.Join(tags, ", ")
			},
			"join": func(s []string) string {
				return strings.Join(s, " | ")
			},
			"joinWord": func(s []string) string {
				return strings.Join(s, "")
			},
		},
		tmplPath: path + "/*.tmpl",
	}
}

func (p *Processor) Fill(a any) (string, string, error) {
	front, err := p.fill("front.tmpl", a)
	if err != nil {
		return "", "", err
	}
	back, err := p.fill("back.tmpl", a)
	if err != nil {
		return "", "", err
	}
	return front, back, err
}

func (p *Processor) fill(name string, a any) (string, error) {
	tmpl, err := template.New(name).Funcs(p.funcMap).ParseGlob(p.tmplPath)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, a)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
