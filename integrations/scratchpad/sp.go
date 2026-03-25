package scratchpad

import (
	"errors"
	"goclaw/agent"
	"os"
	"strings"
)

type ScratchPad interface {
	Content() (string, error)
	Rewrite(oldText, newText string) error
}

func FileScratchPad(path string) ScratchPad {
	return &fileScratchPad{path}
}

type fileScratchPad struct {
	filepath string
}

func (s *fileScratchPad) Content() (string, error) {
	if err := s.ensureFile(); err != nil {
		return "", err
	}
	content, err := os.ReadFile(s.filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

var ErrOldTextNotFound = errors.New("old text was not found in the scratchpad")
var ErrOldTextAmbiguous = errors.New("old text was ambiguous in the scratchpad")

func (s *fileScratchPad) Rewrite(oldText, newText string) error {
	if err := s.ensureFile(); err != nil {
		return err
	}
	content, err := os.ReadFile(s.filepath)
	if err != nil {
		return err
	}
	n := strings.Count(string(content), oldText)
	if n == 0 {
		return ErrOldTextNotFound
	}
	if n > 1 {
		return ErrOldTextAmbiguous
	}
	newContent := strings.ReplaceAll(string(content), oldText, newText)
	err = os.WriteFile(s.filepath, []byte(newContent), os.ModePerm)
	return nil
}

func (s *fileScratchPad) ensureFile() error {
	_, err := os.Stat(s.filepath)
	if err == nil {
		return nil
	}
	f, err := os.Create(s.filepath)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func NewReadScratchPadTool(sp ScratchPad) agent.Tool {
	return &readScratchPadTool{sp: sp}
}

type readScratchPadTool struct {
	sp ScratchPad
}

func (t *readScratchPadTool) Call(map[string]any) (string, error) {
	return t.sp.Content()
}

func (t *readScratchPadTool) Name() string {
	return "read_scratchpad"
}

func (t *readScratchPadTool) Desc() string {
	return "Reads the entire scratchpad, takes no arguments."
}

func NewRewriteScratchPadTool(sp ScratchPad) agent.Tool {
	return &rewriteScratchPadTool{sp: sp}
}

type rewriteScratchPadTool struct {
	sp ScratchPad
}

func (t *rewriteScratchPadTool) Call(args map[string]any) (string, error) {
	oldText, ok := args["old_text"].(string)
	if !ok {
		return "", errors.New("missing or invalid 'old_text'")
	}

	newText, ok := args["new_text"].(string)
	if !ok {
		return "", errors.New("missing or invalid 'new_text'")
	}

	if err := t.sp.Rewrite(oldText, newText); err != nil {
		return "", err
	}

	return "scratchpad updated", nil
}

func (t *rewriteScratchPadTool) Name() string {
	return "rewrite_scratchpad"
}

func (t *rewriteScratchPadTool) Desc() string {
	return "Rewrites part of the scratchpad. Arguments: 'old_text': text to replace (may be empty if scratchpad is empty), 'new_text': replacement text"
}
