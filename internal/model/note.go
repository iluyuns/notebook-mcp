package model

import "time"

type NoteType string

const (
	NoteTypeSummary NoteType = "summary"
	NoteTypeContext NoteType = "context"
	NoteTypeNote    NoteType = "note"
)

type Note struct {
	ID                int64          `json:"id"`
	UserID            string         `json:"user_id,omitempty"`
	NoteType          NoteType       `json:"note_type"`
	Title             string         `json:"title"`
	Content           string         `json:"content"`
	Tags              []string       `json:"tags"`
	SessionID         string         `json:"session_id"`
	SourceInstruction string         `json:"source_instruction"`
	Metadata          map[string]any `json:"metadata"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type SaveNoteRequest struct {
	NoteType          NoteType       `json:"note_type"`
	Title             string         `json:"title"`
	Content           string         `json:"content"`
	Tags              []string       `json:"tags"`
	SessionID         string         `json:"session_id"`
	SourceInstruction string         `json:"source_instruction"`
	Metadata          map[string]any `json:"metadata"`
}

type SearchNotesRequest struct {
	Keyword  string   `form:"keyword" json:"keyword"`
	NoteType NoteType `form:"note_type" json:"note_type"`
	Limit    int      `form:"limit" json:"limit"`
}
