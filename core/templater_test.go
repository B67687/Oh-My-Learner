package core

import (
	"strings"
	"testing"
)

func TestLoadSubjectPack(t *testing.T) {
	data := []byte(`
name = "Test Subject"

[[templates]]
id = "test-1"
question = "What is {{ var1 }}?"
answer = "{{ var1 }} is {{ var2 }}"

[templates.variables]
var1 = ["A", "B"]
var2 = ["X", "Y"]
`)

	templates, err := LoadSubjectPack(data, "test-subject")
	if err != nil {
		t.Fatalf("LoadSubjectPack failed: %v", err)
	}

	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}

	tmpl := templates[0]
	if tmpl.ID != "test-1" {
		t.Errorf("expected ID 'test-1', got %q", tmpl.ID)
	}
	if tmpl.SubjectID != "test-subject" {
		t.Errorf("expected SubjectID 'test-subject', got %q", tmpl.SubjectID)
	}
	if tmpl.QuestionTemplate != "What is {{ var1 }}?" {
		t.Errorf("unexpected QuestionTemplate: %q", tmpl.QuestionTemplate)
	}
	if tmpl.AnswerTemplate != "{{ var1 }} is {{ var2 }}" {
		t.Errorf("unexpected AnswerTemplate: %q", tmpl.AnswerTemplate)
	}
	if len(tmpl.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(tmpl.Variables))
	}
}

func TestLoadSubjectPack_InvalidTOML(t *testing.T) {
	_, err := LoadSubjectPack([]byte("not valid toml {{{"), "test")
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestRenderTemplate(t *testing.T) {
	tmpl := Template{
		ID:               "test-render",
		SubjectID:        "test",
		QuestionTemplate: "Q: {{ var1 }} and {{ var2 }}",
		AnswerTemplate:   "A: {{ var1 }} -> {{ var2 }}",
		Variables: map[string][]string{
			"var1": {"hello", "world"},
			"var2": {"foo", "bar"},
		},
	}

	result, err := RenderTemplate(tmpl)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if result.Question == "" {
		t.Error("expected non-empty question")
	}
	if result.Answer == "" {
		t.Error("expected non-empty answer")
	}

	// Verify no {{ }} braces remain (all variables substituted).
	if strings.Contains(result.Question, "{{") {
		t.Errorf("question still contains unsubstituted braces: %s", result.Question)
	}
	if strings.Contains(result.Answer, "{{") {
		t.Errorf("answer still contains unsubstituted braces: %s", result.Answer)
	}

	// Verify each binding value appears in the rendered output.
	for key, val := range result.Bindings {
		if !strings.Contains(result.Question, val) && !strings.Contains(result.Answer, val) {
			t.Errorf(
				"binding %s=%q not found in question or answer\n  question: %q\n  answer:   %q",
				key, val, result.Question, result.Answer,
			)
		}
	}
}

func TestRenderTemplate_MissingVariableError(t *testing.T) {
	tmpl := Template{
		ID:               "test-missing",
		QuestionTemplate: "Q: {{ missing }}",
		AnswerTemplate:   "A: ok",
		Variables: map[string][]string{
			"var1": {"hello"},
		},
	}

	_, err := RenderTemplate(tmpl)
	if err == nil {
		t.Fatal("expected error for template referencing undefined variable, got nil")
	}
}

func TestRenderTemplate_EmptyVariableError(t *testing.T) {
	tmpl := Template{
		ID:               "test-empty",
		QuestionTemplate: "Q: {{ var1 }}",
		AnswerTemplate:   "A: ok",
		Variables: map[string][]string{
			"var1": {},
		},
	}

	_, err := RenderTemplate(tmpl)
	if err == nil {
		t.Fatal("expected error for variable with no values, got nil")
	}
}

func TestRenderTemplate_SingleVariable(t *testing.T) {
	tmpl := Template{
		ID:               "test-single",
		QuestionTemplate: "What is the answer to {{ question }}?",
		AnswerTemplate:   "The answer is {{ answer }}.",
		Variables: map[string][]string{
			"question": {"life"},
			"answer":   {"42"},
		},
	}

	result, err := RenderTemplate(tmpl)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if result.Question != "What is the answer to life?" {
		t.Errorf("unexpected question: %q", result.Question)
	}
	if result.Answer != "The answer is 42." {
		t.Errorf("unexpected answer: %q", result.Answer)
	}
}

func TestRenderTemplate_WithSubjectPackRoundtrip(t *testing.T) {
	// Load a real-style TOML and verify rendering works end-to-end.
	data := []byte(`
name = "Algorithms"

[[templates]]
id = "complexity-sorting"
question = "What is the worst-case time complexity of {{ algorithm }}?"
answer = "{{ algorithm }} has a worst-case complexity of {{ complexity }}."

[templates.variables]
algorithm = ["Bubble Sort", "Merge Sort"]
complexity = ["O(n²)", "O(n log n)"]
`)

	templates, err := LoadSubjectPack(data, "algorithms")
	if err != nil {
		t.Fatalf("LoadSubjectPack failed: %v", err)
	}

	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}

	result, err := RenderTemplate(templates[0])
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	// Verify the output looks valid.
	if !strings.HasPrefix(result.Question, "What is the worst-case time complexity of ") {
		t.Errorf("question doesn't start with expected prefix: %q", result.Question)
	}
	if !strings.HasSuffix(result.Question, "?") {
		t.Errorf("question doesn't end with '?': %q", result.Question)
	}

	// Verify bindings match rendered output.
	for key, val := range result.Bindings {
		if !strings.Contains(result.Question, val) && !strings.Contains(result.Answer, val) {
			t.Errorf(
				"binding %s=%q does not appear in output\n  question: %q\n  answer:   %q",
				key, val, result.Question, result.Answer,
			)
		}
	}
}
