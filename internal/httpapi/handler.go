package httpapi

import (
	"net/http"

	"notebook-mcp/internal/model"
	"notebook-mcp/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.NoteService
}

func NewHandler(svc *service.NoteService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	v1.POST("/notes", h.saveNote)
	v1.GET("/notes/search", h.searchNotes)
}

func (h *Handler) saveNote(c *gin.Context) {
	var req model.SaveNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	note, err := h.svc.Save(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, note)
}

func (h *Handler) searchNotes(c *gin.Context) {
	var req model.SearchNotesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	notes, err := h.svc.Search(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": notes, "count": len(notes)})
}
