package core

import (
	"testing"
)

func TestKnowledgeTypeConstants(t *testing.T) {
	if KnowledgeDeclarative != "declarative" {
		t.Errorf("KnowledgeDeclarative = %q, want %q", KnowledgeDeclarative, "declarative")
	}
	if KnowledgeProcedural != "procedural" {
		t.Errorf("KnowledgeProcedural = %q, want %q", KnowledgeProcedural, "procedural")
	}
}

func TestTemplateKnowledgeTypeField(t *testing.T) {
	tmpl := Template{
		ID:            "test-kt",
		SubjectID:     "test",
		KnowledgeType: KnowledgeProcedural,
	}

	if tmpl.KnowledgeType != KnowledgeProcedural {
		t.Errorf("KnowledgeType = %q, want %q", tmpl.KnowledgeType, KnowledgeProcedural)
	}

	// Default should be empty (zero value).
	var empty Template
	if empty.KnowledgeType != "" {
		t.Errorf("zero-value KnowledgeType should be empty, got %q", empty.KnowledgeType)
	}
}
