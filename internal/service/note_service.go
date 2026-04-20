package service

import (
	"context"
	"fmt"
	"strings"

	"notebook-mcp/internal/model"
	"notebook-mcp/internal/repo"
)

type NoteService struct {
	repo             *repo.NoteRepo
	defaultQuerySize int
}

func NewNoteService(repo *repo.NoteRepo, defaultQuerySize int) *NoteService {
	return &NoteService{
		repo:             repo,
		defaultQuerySize: defaultQuerySize,
	}
}

func (s *NoteService) Save(ctx context.Context, req model.SaveNoteRequest) (model.Note, error) {
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		return model.Note{}, fmt.Errorf("content is required")
	}
	if !isValidNoteType(req.NoteType) {
		return model.Note{}, fmt.Errorf("invalid note_type: %s", req.NoteType)
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}
	return s.repo.Save(ctx, req)
}

func (s *NoteService) Search(ctx context.Context, req model.SearchNotesRequest) ([]model.Note, error) {
	req.Keyword = strings.TrimSpace(req.Keyword)
	if req.Keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if req.Limit <= 0 {
		req.Limit = s.defaultQuerySize
	}
	if req.NoteType != "" && !isValidNoteType(req.NoteType) {
		return nil, fmt.Errorf("invalid note_type: %s", req.NoteType)
	}
	return s.repo.Search(ctx, req)
}

func (s *NoteService) SaveByInstruction(
	ctx context.Context,
	instruction string,
	content string,
	title string,
	sessionID string,
	tags []string,
) (model.Note, error) {
	noteType := InferNoteTypeByInstruction(instruction)
	if noteType == "" {
		return model.Note{}, fmt.Errorf("instruction not supported: %s", instruction)
	}
	return s.Save(ctx, model.SaveNoteRequest{
		NoteType:          noteType,
		Title:             title,
		Content:           content,
		SessionID:         sessionID,
		Tags:              tags,
		SourceInstruction: strings.TrimSpace(instruction),
	})
}

func InferNoteTypeByInstruction(instruction string) model.NoteType {
	t := strings.TrimSpace(instruction)
	switch {
	case strings.Contains(t, "总结"):
		return model.NoteTypeSummary
	case strings.Contains(t, "上下文"):
		return model.NoteTypeContext
	case strings.Contains(t, "笔记"):
		return model.NoteTypeNote
	default:
		return ""
	}
}

func isValidNoteType(t model.NoteType) bool {
	return t == model.NoteTypeSummary || t == model.NoteTypeContext || t == model.NoteTypeNote
}
