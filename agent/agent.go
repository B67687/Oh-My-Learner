// Package agent provides AI-powered card generation for Oh-My-Learner.
//
// The Agent interface abstracts card generation so the CLI can work with
// any AI provider. DeepSeekAgent is the initial implementation using
// DeepSeek V4 Flash's free API (OpenAI-compatible).
package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIGeneratedCard is a concrete card produced by the AI.
// Type must be one of: standard, code-trace, debug-find, explain-why.
// AIGeneratedCard is a concrete card produced by the AI.
// Type must be one of: standard, code-trace, debug-find, explain-why.
// KnowledgeType must be one of: declarative, procedural.
type AIGeneratedCard struct {
	Type          string `json:"type"`
	KnowledgeType string `json:"knowledge_type"`
	Question      string `json:"question"`
	Answer        string `json:"answer"`
}

// Agent generates practice cards for a given topic using AI.
type Agent interface {
	// GenerateCards returns 5-10 concrete practice cards for topic.
	// Cards use the 4 existing template types (standard, code-trace,
	// debug-find, explain-why). Returns an error if the AI service
	// is unavailable or returns invalid data.
	GenerateCards(topic string) ([]AIGeneratedCard, error)
}

// DeepSeekAgent calls the DeepSeek V4 Flash API (OpenAI-compatible).
type DeepSeekAgent struct {
	apiKey string
	client *http.Client
}

// NewDeepSeekAgent creates an agent using OML_DEEPSEEK_KEY env var.
func NewDeepSeekAgent() (*DeepSeekAgent, error) {
	apiKey := os.Getenv("OML_DEEPSEEK_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OML_DEEPSEEK_KEY environment variable not set")
	}
	return &DeepSeekAgent{
		apiKey: apiKey,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// chatMessage is one message in the OpenAI-compatible chat format.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the request body for the chat completions API.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// chatResponse is the response from the chat completions API.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

const deepseekEndpoint = "https://api.deepseek.com/v1/chat/completions"
const deepseekModel = "deepseek-chat"

// GenerateCards sends a topic to DeepSeek and parses the response.
func (a *DeepSeekAgent) GenerateCards(topic string) ([]AIGeneratedCard, error) {
	systemPrompt := `You are a university computer science tutor creating practice questions for a student.`
	systemPrompt += ` Generate 5-10 concrete practice questions about the given topic.`
	systemPrompt += ` Use exactly these 4 question types (distribute across types as appropriate):`
	systemPrompt += ` 1. "standard" — Direct Q&A (e.g., "What is paging?")`
	systemPrompt += ` 2. "code-trace" — Show code, ask what it outputs (must include a code snippet in the question)`
	systemPrompt += ` 3. "debug-find" — Show buggy code, ask what the bug is (must include a code snippet in the question)`
	systemPrompt += ` 4. "explain-why" — Ask why something works the way it does`
	systemPrompt += ` Each card also has a knowledge_type field:`
	systemPrompt += ` - "declarative" — for definitions, facts, terminology, conceptual explanations`
	systemPrompt += ` - "procedural" — for algorithms, code, problem-solving, step-by-step processes`
	systemPrompt += ` Rules:`
	systemPrompt += ` - Questions must be specific and university-level`
	systemPrompt += ` - Answers must be complete and accurate`
	systemPrompt += ` - Code snippets must be real, runnable code (use the {{code}} marker where the code belongs in the question)`
	systemPrompt += ` - Output ONLY a valid JSON array with no markdown formatting`
	systemPrompt += ` - Each element: {"type": "...", "knowledge_type": "...", "question": "...", "answer": "..."}`
	systemPrompt += ` - Skip types that don't fit the topic`
	systemPrompt += ` - Do NOT wrap the JSON in markdown code blocks`

	userPrompt := fmt.Sprintf("Create practice questions about: %s", topic)

	reqBody := chatRequest{
		Model:       deepseekModel,
		Temperature: 0.7,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	body, err := a.callAPI(reqBody)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	cards, err := parseCardsResponse(body)
	if err != nil {
		return nil, fmt.Errorf("parse AI response: %w", err)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("AI returned no valid cards for topic %q", topic)
	}

	return cards, nil
}

// callAPI sends the request to DeepSeek and returns the raw response body.
func (a *DeepSeekAgent) callAPI(req chatRequest) ([]byte, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", deepseekEndpoint, strings.NewReader(string(reqJSON)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal chat response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("API returned no choices")
	}

	content := chatResp.Choices[0].Message.Content
	return []byte(content), nil
}

// parseCardsResponse parses AI JSON content into valid cards.
// It extracts a JSON array from the response, tolerating markdown
// code block wrappers that some AI models add despite instructions.
func parseCardsResponse(data []byte) ([]AIGeneratedCard, error) {
	cleaned := strings.TrimSpace(string(data))

	// Strip markdown code fences if the AI ignored instructions.
	if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		if idx := strings.LastIndex(cleaned, "```"); idx >= 0 {
			cleaned = cleaned[:idx]
		}
		cleaned = strings.TrimSpace(cleaned)
	}

	var parsed []AIGeneratedCard
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON from AI: %w\nRaw content: %.200s", err, string(data))
	}

	// Validate each card.
	validTypes := map[string]bool{
		"standard":    true,
		"code-trace":  true,
		"debug-find":  true,
		"explain-why": true,
	}
	validKnowledgeTypes := map[string]bool{
		"declarative": true,
		"procedural":  true,
	}

	var valid []AIGeneratedCard
	for _, c := range parsed {
		if c.Question == "" || c.Answer == "" {
			continue
		}
		if !validTypes[c.Type] {
			continue
		}
		if c.KnowledgeType == "" || !validKnowledgeTypes[c.KnowledgeType] {
			// Default to declarative if not specified or invalid.
			c.KnowledgeType = "declarative"
		}
		valid = append(valid, c)
	}

	if len(valid) == 0 {
		return nil, fmt.Errorf("no valid cards after validation (parsed %d cards, all rejected)", len(parsed))
	}

	return valid, nil
}
