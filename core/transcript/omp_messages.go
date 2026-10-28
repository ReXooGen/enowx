package transcript

import (
	"encoding/json"
	"strings"

	"github.com/enowx/enowx/proto"
)

func (p *ompParser) message(r *Reader, rec ompRecord) {
	var blocks []ompContentBlock
	var text string
	if json.Unmarshal(rec.Message.Content, &text) == nil {
		blocks = []ompContentBlock{{Type: "text", Text: text}}
	} else {
		_ = json.Unmarshal(rec.Message.Content, &blocks)
	}
	if rec.Message.Model != "" {
		r.Model = rec.Message.Model
	}
	if rec.Message.Role == "toolResult" || rec.Message.Role == "tool_result" {
		p.toolResult(r, rec, blocks)
		return
	}
	if rec.Message.Role == "user" {
		text, images := ompContent(blocks)
		if text != "" || len(images) > 0 {
			r.add(proto.TranscriptEntry{Kind: proto.TranscriptUser, Time: rec.Timestamp, Text: text, Images: images})
		}
		return
	}
	if rec.Message.Role != "assistant" {
		return
	}
	for _, b := range blocks {
		switch b.Type {
		case "thinking":
			if strings.TrimSpace(b.Thinking) != "" {
				r.add(proto.TranscriptEntry{Kind: proto.TranscriptThinking, Time: rec.Timestamp, Text: b.Thinking})
			}
		case "text":
			if strings.TrimSpace(b.Text) != "" {
				r.add(proto.TranscriptEntry{Kind: proto.TranscriptAssistant, Time: rec.Timestamp, Text: b.Text})
			}
		case "toolCall", "tool_call":
			e := proto.TranscriptEntry{Kind: proto.TranscriptTool, Time: rec.Timestamp, Text: ompToolName(b.Name), Detail: toolDetail(b.Arguments)}
			toolEdit(&e, b.Arguments)
			i := r.add(e)
			if b.ID != "" {
				r.tools[b.ID] = i
			}
		case "image":
			_, images := ompContent([]ompContentBlock{b})
			r.add(proto.TranscriptEntry{Kind: proto.TranscriptAssistant, Time: rec.Timestamp, Images: images})
		}
	}
	if rec.Message.ErrorMessage != "" {
		r.add(proto.TranscriptEntry{Kind: proto.TranscriptNotice, Time: rec.Timestamp, Text: rec.Message.ErrorMessage, Failed: true})
	}
}

func (p *ompParser) toolResult(r *Reader, rec ompRecord, blocks []ompContentBlock) {
	text, images := ompContent(blocks)
	i, ok := r.tools[rec.Message.ToolCallID]
	if !ok {
		i = r.add(proto.TranscriptEntry{Kind: proto.TranscriptTool, Time: rec.Timestamp, Text: ompToolName(rec.Message.ToolName)})
		if rec.Message.ToolCallID != "" {
			r.tools[rec.Message.ToolCallID] = i
		}
	}
	r.change(i, func(e *proto.TranscriptEntry) {
		e.Done, e.Failed, e.Result, e.Images = true, rec.Message.IsError, clip(text, 4000), images
		if !e.Time.IsZero() && rec.Timestamp.After(e.Time) {
			e.DurationMs = rec.Timestamp.Sub(e.Time).Milliseconds()
		}
	})
}

func ompContent(blocks []ompContentBlock) (string, []proto.TranscriptImage) {
	var text []string
	var images []proto.TranscriptImage
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			text = append(text, b.Text)
		}
		if b.Type == "image" && len(images) < 2 {
			images = append(images, transcriptThumbnail(b.Data))
		}
	}
	return strings.TrimSpace(strings.Join(text, "\n")), images
}

func ompToolName(name string) string {
	if display, ok := map[string]string{"bash": "Bash", "read": "Read", "write": "Write", "edit": "Edit", "grep": "Grep", "find": "Glob", "ls": "LS"}[name]; ok {
		return display
	}
	if name == "" {
		return "Tool"
	}
	return name
}
