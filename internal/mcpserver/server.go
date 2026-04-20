package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"notebook-mcp/internal/model"
	"notebook-mcp/internal/service"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func New(svc *service.NoteService) *server.StreamableHTTPServer {
	s := server.NewMCPServer("private-notebook", "1.0.0", server.WithToolCapabilities(true))

	s.AddTool(
		mcp.NewTool("save_session_summary",
			mcp.WithDescription("保存本次总结"),
			mcp.WithString("content", mcp.Required(), mcp.Description("总结内容")),
			mcp.WithString("title", mcp.Description("标题")),
			mcp.WithString("session_id", mcp.Description("会话标识")),
			mcp.WithArray("tags", mcp.Description("标签"), mcp.Items(map[string]any{"type": "string"})),
		),
		saveHandler(svc, model.NoteTypeSummary),
	)
	s.AddTool(
		mcp.NewTool("save_session_context",
			mcp.WithDescription("保存本次上下文"),
			mcp.WithString("content", mcp.Required(), mcp.Description("上下文内容")),
			mcp.WithString("title", mcp.Description("标题")),
			mcp.WithString("session_id", mcp.Description("会话标识")),
			mcp.WithArray("tags", mcp.Description("标签"), mcp.Items(map[string]any{"type": "string"})),
		),
		saveHandler(svc, model.NoteTypeContext),
	)
	s.AddTool(
		mcp.NewTool("save_session_note",
			mcp.WithDescription("保存本次笔记"),
			mcp.WithString("content", mcp.Required(), mcp.Description("笔记内容")),
			mcp.WithString("title", mcp.Description("标题")),
			mcp.WithString("session_id", mcp.Description("会话标识")),
			mcp.WithArray("tags", mcp.Description("标签"), mcp.Items(map[string]any{"type": "string"})),
		),
		saveHandler(svc, model.NoteTypeNote),
	)
	s.AddTool(
		mcp.NewTool("save_by_instruction",
			mcp.WithDescription("按自然语言指令保存，示例关键词: 总结/上下文/笔记"),
			mcp.WithString("instruction", mcp.Required(), mcp.Description("触发指令，如: 帮我保存本次总结")),
			mcp.WithString("content", mcp.Required(), mcp.Description("要保存的内容")),
			mcp.WithString("title", mcp.Description("标题")),
			mcp.WithString("session_id", mcp.Description("会话标识")),
			mcp.WithArray("tags", mcp.Description("标签"), mcp.Items(map[string]any{"type": "string"})),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			note, err := svc.SaveByInstruction(
				ctx,
				req.GetString("instruction", ""),
				req.GetString("content", ""),
				req.GetString("title", ""),
				req.GetString("session_id", ""),
				readStringArray(req.GetArguments()["tags"]),
			)
			if err != nil {
				return nil, err
			}
			return resultJSON(note)
		},
	)
	s.AddTool(
		mcp.NewTool("query_notes",
			mcp.WithDescription("查询是否存在某关键词笔记"),
			mcp.WithString("keyword", mcp.Required(), mcp.Description("查询关键词")),
			mcp.WithString("note_type", mcp.Description("可选: summary/context/note")),
			mcp.WithNumber("limit", mcp.Description("返回数量，默认 20，最大 100")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			notes, err := svc.Search(ctx, model.SearchNotesRequest{
				Keyword:  req.GetString("keyword", ""),
				NoteType: model.NoteType(strings.TrimSpace(req.GetString("note_type", ""))),
				Limit:    int(req.GetFloat("limit", 20)),
			})
			if err != nil {
				return nil, err
			}
			return resultJSON(ginResult{
				Count: len(notes),
				Items: notes,
			})
		},
	)

	return server.NewStreamableHTTPServer(s)
}

type ginResult struct {
	Count int          `json:"count"`
	Items []model.Note `json:"items"`
}

func saveHandler(svc *service.NoteService, noteType model.NoteType) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		note, err := svc.Save(ctx, model.SaveNoteRequest{
			NoteType:  noteType,
			Title:     req.GetString("title", ""),
			Content:   req.GetString("content", ""),
			SessionID: req.GetString("session_id", ""),
			Tags:      readStringArray(req.GetArguments()["tags"]),
		})
		if err != nil {
			return nil, err
		}
		return resultJSON(note)
	}
}

func resultJSON(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(b)), nil
}

func readStringArray(v any) []string {
	if v == nil {
		return nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	items := make([]string, 0, len(raw))
	for _, item := range raw {
		s := strings.TrimSpace(fmt.Sprintf("%v", item))
		if s == "" {
			continue
		}
		items = append(items, s)
	}
	return items
}
