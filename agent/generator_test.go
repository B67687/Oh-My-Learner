package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCardsResponse_Valid(t *testing.T) {
	data := readFixture(t, "valid-cards.json")
	cards, err := parseCardsResponse(data)
	if err != nil {
		t.Fatalf("parseCardsResponse failed: %v", err)
	}
	if len(cards) == 0 {
		t.Fatal("expected at least 1 card")
	}
	for _, c := range cards {
		if c.Question == "" {
			t.Errorf("card has empty question")
		}
		if c.Answer == "" {
			t.Errorf("card has empty answer")
		}
		switch c.Type {
		case "standard", "code-trace", "debug-find", "explain-why":
			// OK
		default:
			t.Errorf("card has unknown type %q", c.Type)
		}
	}
}

func TestParseCardsResponse_MarkdownWrapped(t *testing.T) {
	data := readFixture(t, "markdown-wrapped.json")
	cards, err := parseCardsResponse(data)
	if err != nil {
		t.Fatalf("parseCardsResponse failed for markdown-wrapped: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards from markdown-wrapped fixture, got %d", len(cards))
	}
}

func TestParseCardsResponse_Garbage(t *testing.T) {
	data := readFixture(t, "garbage-response.json")
	_, err := parseCardsResponse(data)
	if err == nil {
		t.Fatal("expected error for garbage response, got nil")
	}
}

func TestParseCardsResponse_Empty(t *testing.T) {
	data := readFixture(t, "empty-cards.json")
	_, err := parseCardsResponse(data)
	if err == nil {
		t.Fatal("expected error for empty cards array, got nil")
	}
}

func TestParseCardsResponse_InvalidTypes(t *testing.T) {
	data := readFixture(t, "invalid-types.json")
	_, err := parseCardsResponse(data)
	if err == nil {
		t.Fatal("expected error when all cards are invalid, got nil")
	}
}

func TestParseCardsResponse_KnowledgeTypes(t *testing.T) {
	data := readFixture(t, "knowledge-types.json")
	cards, err := parseCardsResponse(data)
	if err != nil {
		t.Fatalf("parseCardsResponse failed: %v", err)
	}

	// First card should be declarative.
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].KnowledgeType != "declarative" {
		t.Errorf("card[0].KnowledgeType = %q, want 'declarative'", cards[0].KnowledgeType)
	}

	// Second card should be procedural.
	if len(cards) < 2 {
		t.Fatal("expected at least 2 cards")
	}
	if cards[1].KnowledgeType != "procedural" {
		t.Errorf("card[1].KnowledgeType = %q, want 'procedural'", cards[1].KnowledgeType)
	}

	// Third card has invalid knowledge_type — should default to declarative.
	if len(cards) < 3 {
		t.Fatal("expected at least 3 cards")
	}
	if cards[2].KnowledgeType != "declarative" {
		t.Errorf("card[2].KnowledgeType = %q, want 'declarative' (default for invalid)", cards[2].KnowledgeType)
	}
}
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}
