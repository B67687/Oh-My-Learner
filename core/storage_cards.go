package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *Storage) AddCard(tmpl Template) (CardState, error) {
	id := uuid.New().String()
	now := time.Now()
	state := CardState{
		ID:             id,
		SubjectID:      tmpl.SubjectID,
		EasinessFactor: 2.5,
		IntervalDays:   0,
		Repetition:     0,
		NextReviewAt:   now,
		CreatedAt:      now,
	}
	if err := s.InsertCard(state, tmpl); err != nil {
		return CardState{}, err
	}
	return state, nil
}

// DueCards returns all cards whose next_review_at is at or before now.
// If knowledgeType is non-empty, only cards of that type are returned.
func (s *Storage) DueCards(now time.Time, knowledgeType ...KnowledgeType) ([]CardState, error) {
	nowStr := now.Format(time.RFC3339)

	var query string
	var args []interface{}
	args = append(args, nowStr)

	if len(knowledgeType) > 0 && knowledgeType[0] != "" {
		query = `SELECT id, subject_id, easiness_factor, interval_days, repetition,
		        next_review_at, created_at
		   FROM cards
		  WHERE next_review_at <= ?
		    AND knowledge_type = ?
	   ORDER BY RANDOM()`
		args = append(args, string(knowledgeType[0]))
	} else {
		query = `SELECT id, subject_id, easiness_factor, interval_days, repetition,
		        next_review_at, created_at
		   FROM cards
		  WHERE next_review_at <= ?
	   ORDER BY RANDOM()`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query due cards: %w", err)
	}
	defer rows.Close()

	var cards []CardState
	for rows.Next() {
		var (
			c             CardState
			nextReviewStr string
			createdAtStr  string
		)
		if err := rows.Scan(
			&c.ID, &c.SubjectID, &c.EasinessFactor, &c.IntervalDays,
			&c.Repetition, &nextReviewStr, &createdAtStr,
		); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		var err error
		c.NextReviewAt, err = time.Parse(time.RFC3339, nextReviewStr)
		if err != nil {
			return nil, fmt.Errorf("parse next_review_at for card %s: %w", c.ID, err)
		}
		c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse created_at for card %s: %w", c.ID, err)
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// GetCardWithTemplate loads a card with its rendering data.
func (s *Storage) GetCardWithTemplate(cardID string) (*CardWithTemplate, error) {
	var (
		cwt           CardWithTemplate
		nextReviewStr string
		createdAtStr  string
		variablesStr  string
		typeStr       string
		ktStr         string
	)

	err := s.db.QueryRow(
		`SELECT id, subject_id, easiness_factor, interval_days, repetition,
		        next_review_at, created_at,
		        template_type, knowledge_type, template_question, template_answer, variables
		   FROM cards WHERE id = ?`, cardID,
	).Scan(
		&cwt.State.ID, &cwt.State.SubjectID,
		&cwt.State.EasinessFactor, &cwt.State.IntervalDays, &cwt.State.Repetition,
		&nextReviewStr, &createdAtStr,
		&typeStr, &ktStr, &cwt.QuestionTemplate, &cwt.AnswerTemplate,
		&variablesStr,
	)
	if err != nil {
		return nil, fmt.Errorf("get card with template: %w", err)
	}

	var errParse error
	cwt.State.NextReviewAt, errParse = time.Parse(time.RFC3339, nextReviewStr)
	if errParse != nil {
		return nil, fmt.Errorf("parse next_review_at for card %s: %w", cwt.State.ID, errParse)
	}
	cwt.State.CreatedAt, errParse = time.Parse(time.RFC3339, createdAtStr)
	if errParse != nil {
		return nil, fmt.Errorf("parse created_at for card %s: %w", cwt.State.ID, errParse)
	}
	cwt.Type = TemplateType(typeStr)
	cwt.KnowledgeType = KnowledgeType(ktStr)

	if err := json.Unmarshal([]byte(variablesStr), &cwt.Variables); err != nil {
		return nil, fmt.Errorf("unmarshal variables: %w", err)
	}

	return &cwt, nil
}

// InsertCard creates a new card from its state and template data.
func (s *Storage) InsertCard(state CardState, tmpl Template) error {
	variablesJSON, err := json.Marshal(tmpl.Variables)
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}

	if tmpl.KnowledgeType == "" {
		tmpl.KnowledgeType = KnowledgeDeclarative
	}

	_, err = s.db.Exec(
		`INSERT INTO cards
		   (id, subject_id, template_type, knowledge_type, template_question, template_answer, variables,
		    easiness_factor, interval_days, repetition, next_review_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		state.ID, state.SubjectID,
		string(tmpl.Type), string(tmpl.KnowledgeType),
		tmpl.QuestionTemplate, tmpl.AnswerTemplate,
		string(variablesJSON),
		state.EasinessFactor, state.IntervalDays, state.Repetition,
		state.NextReviewAt.Format(time.RFC3339),
		state.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// UpdateCardState persists SM-2 scheduling fields for an existing card.
// Returns an error if the card ID does not exist.
func (s *Storage) UpdateCardState(state CardState) error {
	res, err := s.db.Exec(
		`UPDATE cards
		    SET easiness_factor = ?, interval_days = ?, repetition = ?,
		        next_review_at = ?
		  WHERE id = ?`,
		state.EasinessFactor, state.IntervalDays, state.Repetition,
		state.NextReviewAt.Format(time.RFC3339),
		state.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("card %s not found", state.ID)
	}
	return nil
}

// InsertReview records a review event for audit/analytics.
func (s *Storage) InsertReview(cardID string, quality uint8, selfExplainResponse string) error {
	id := uuid.New().String()
	_, err := s.db.Exec(
		`INSERT INTO reviews (id, card_id, quality, reviewed_at, self_explain_response)
		 VALUES (?, ?, ?, ?, ?)`,
		id, cardID, quality, time.Now().Format(time.RFC3339), selfExplainResponse,
	)
	return err
}
