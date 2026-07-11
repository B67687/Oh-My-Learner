package core

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"regexp"
	"text/template"

	"github.com/pelletier/go-toml/v2"
)

// subjectPack is the top-level TOML structure for a subject pack file.
type subjectPack struct {
	Name      string         `toml:"name"`
	Templates []packTemplate `toml:"templates"`
}

// packTemplate is a single template entry in the TOML file.
type packTemplate struct {
	ID        string              `toml:"id"`
	Type      string              `toml:"type"`
	Question  string              `toml:"question"`
	Answer    string              `toml:"answer"`
	Variables map[string][]string `toml:"variables"`
}

// templateVarRe converts {{ name }} to {{.name}} for Go's text/template.
var templateVarRe = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)

// LoadSubjectPack parses TOML bytes into a slice of Template.
func LoadSubjectPack(data []byte, subjectID string) ([]Template, error) {
	var pack subjectPack
	if err := toml.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("parse subject pack: %w", err)
	}

	templates := make([]Template, len(pack.Templates))
	for i, pt := range pack.Templates {
		t := TemplateType(pt.Type)
		if t == "" {
			t = TemplateStandard
		}
		templates[i] = Template{
			ID:               pt.ID,
			SubjectID:        subjectID,
			Type:             t,
			QuestionTemplate: pt.Question,
			AnswerTemplate:   pt.Answer,
			Variables:        pt.Variables,
		}
	}
	return templates, nil
}

// RenderTemplate creates a RenderedProblem by randomly selecting variable values
// and substituting them into the question and answer templates.
func RenderTemplate(tmpl Template) (*RenderedProblem, error) {
	bindings := make(map[string]string, len(tmpl.Variables))
	for key, vals := range tmpl.Variables {
		if len(vals) == 0 {
			return nil, fmt.Errorf("template %q: variable %q has no values", tmpl.ID, key)
		}
		bindings[key] = vals[rand.IntN(len(vals))]
	}

	question, err := renderText(tmpl.QuestionTemplate, bindings)
	if err != nil {
		return nil, fmt.Errorf("render question for template %q: %w", tmpl.ID, err)
	}

	answer, err := renderText(tmpl.AnswerTemplate, bindings)
	if err != nil {
		return nil, fmt.Errorf("render answer for template %q: %w", tmpl.ID, err)
	}

	return &RenderedProblem{
		Question: question,
		Answer:   answer,
		Bindings: bindings,
	}, nil
}

// renderText substitutes bindings into a template string using Go's text/template.
func renderText(tmplText string, bindings map[string]string) (string, error) {
	converted := templateVarRe.ReplaceAllString(tmplText, "{{.$1}}")

	t, err := template.New("").Option("missingkey=error").Parse(converted)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, bindings); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
