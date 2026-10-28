package transcript

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

var ompUUID = regexp.MustCompile(`_([0-9a-fA-F-]+)\.jsonl$`)

// OMPFiles returns the transcripts OMP keeps for sessions run in dir, newest first.
func OMPFiles(home, dir string) []Found {
	cleanDir := filepath.Clean(dir)
	var all []Found
	root := filepath.Join(home, ".omp", "agent", "sessions")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		folder := filepath.Join(root, e.Name())
		files, err := os.ReadDir(folder)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") || strings.HasPrefix(f.Name(), ".") {
				continue
			}
			path := filepath.Join(folder, f.Name())
			if strings.EqualFold(ompCwd(path), cleanDir) {
				if info, err := f.Info(); err == nil {
					all = append(all, Found{Path: path, Changed: info.ModTime()})
				}
			}
		}
	}
	slices.SortFunc(all, func(a, b Found) int {
		return b.Changed.Compare(a.Changed)
	})
	return all
}

// OMPConversationID is the id OMP resumes the conversation at path by:
// in its file's name, else its first session record.
func OMPConversationID(path string) string {
	if m := ompUUID.FindStringSubmatch(path); len(m) > 1 {
		return m[1]
	}
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var rec struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		}
		if json.Unmarshal(scanner.Bytes(), &rec) == nil && rec.Type == "session" && rec.ID != "" {
			return rec.ID
		}
	}
	return ""
}

func ompCwd(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var rec struct {
			Type string `json:"type"`
			CWD  string `json:"cwd"`
		}
		if json.Unmarshal(scanner.Bytes(), &rec) == nil && rec.Type == "session" {
			return filepath.Clean(rec.CWD)
		}
	}
	return ""
}

// NewOMPReader follows an OMP transcript.
func NewOMPReader(path string) *Reader {
	return newReader(path, &ompParser{})
}

type ompParser struct{}

func (*ompParser) fresh() parser { return &ompParser{} }

type ompRecord struct {
	Type      string    `json:"type"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`
	Message   struct {
		Role         string          `json:"role"`
		Content      json.RawMessage `json:"content"`
		ToolCallID   string          `json:"toolCallId"`
		ToolName     string          `json:"toolName"`
		IsError      bool            `json:"isError"`
		Model        string          `json:"model"`
		ErrorMessage string          `json:"errorMessage"`
	} `json:"message"`
}

type ompContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Thinking  string          `json:"thinking"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
	Data      string          `json:"data"`
}

func (p *ompParser) parse(line []byte, r *Reader) {
	var rec ompRecord
	if json.Unmarshal(line, &rec) != nil {
		return
	}
	if rec.Type == "model_change" && rec.Model != "" {
		r.Model = rec.Model
		return
	}
	if rec.Type != "message" {
		return
	}
	p.message(r, rec)
}
