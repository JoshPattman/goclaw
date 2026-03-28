package filesystem

import (
	"fmt"
	"goclaw/agent"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func Tools() []agent.Tool {
	return []agent.Tool{
		readFileTool{10000},
		listDirectoryTool{},
	}
}

type readFileTool struct {
	maxLength int
}

func (t readFileTool) Name() string {
	return "read_file"
}

func (t readFileTool) Desc() string {
	return fmt.Sprintf(
		"Read a number of characters from the specified file, starting at the offset. You can read up to %d characters. Specify, 'path', 'offset' and 'max_characters' If you want to read the whole file, set offset to 0 and max_characters to %d. If you want to read just the start of a file, set offset to 0 and max characters to a smaller sensible number. If you want to read from a certain point in the file, set offset to that point.",
		t.maxLength,
		t.maxLength,
	)
}

func (t readFileTool) Call(args map[string]any) (string, error) {
	pathAny, ok := args["offset"]
	if !ok {
		return "", fmt.Errorf("must specify a 'path'.")
	}
	path, ok := pathAny.(string)
	if !ok {
		return "", fmt.Errorf("'path' must be a string")
	}
	offsetAny, ok := args["offset"]
	if !ok {
		return "", fmt.Errorf("must specify an 'offset' (use 0 if you want no offset).")
	}
	offset, ok := offsetAny.(float64)
	if !ok {
		return "", fmt.Errorf("'offset' must be an integer")
	}
	maxCharactersAny, ok := args["max_characters"]
	if !ok {
		return "", fmt.Errorf("must specify a 'max_characters' (use the maximum if you want as much as possible).")
	}
	maxCharacters, ok := maxCharactersAny.(float64)
	if !ok {
		return "", fmt.Errorf("'max_characters' must be an integer")
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	realOffset := min(max(0, int(offset)), len(data))
	realEndOffset := min(max(0, int(offset+maxCharacters)), len(data))
	resultText := string(data[realOffset:realEndOffset])
	return fmt.Sprintf("File contents of %s from character %d to %d (%d characters):\n%s", path, realOffset, realEndOffset, realEndOffset-realOffset, resultText), nil
}

type listDirectoryTool struct{}

func (t listDirectoryTool) Name() string {
	return "list_directory"
}

func (t listDirectoryTool) Desc() string {
	return "List the contents of a directory. Specify 'path'. Returns files and subdirectories."
}

func (t listDirectoryTool) Call(args map[string]any) (string, error) {
	pathAny, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("must specify a 'path'")
	}

	path, ok := pathAny.(string)
	if !ok {
		return "", fmt.Errorf("'path' must be a string")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	var dirs []string
	var files []string

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(path, name)

		if entry.IsDir() {
			dirs = append(dirs, name+"/")
		} else {
			files = append(files, name)
		}

		_ = fullPath // (handy if you later want stats, sizes, etc.)
	}

	sort.Strings(dirs)
	sort.Strings(files)

	result := fmt.Sprintf("Contents of directory %s:\n", path)

	if len(dirs) > 0 {
		result += "\nDirectories:\n"
		for _, d := range dirs {
			result += "  " + d + "\n"
		}
	}

	if len(files) > 0 {
		result += "\nFiles:\n"
		for _, f := range files {
			result += "  " + f + "\n"
		}
	}

	if len(dirs) == 0 && len(files) == 0 {
		result += "  (empty directory)\n"
	}

	return result, nil
}
