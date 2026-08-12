package resolver

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/jillesme/tourminal/internal/tour"
)

const maxSourceSize = 2 << 20

type Kind int

const (
	Content Kind = iota
	File
	Directory
	URI
)

type ResolvedStep struct {
	Kind           Kind
	Path           string
	DisplayPath    string
	Source         string
	TargetLine     int
	SelectionStart int
	SelectionEnd   int
	Warning        string
}

func Resolve(root string, step tour.Step) (ResolvedStep, error) {
	if step.Directory != "" {
		path, err := safePath(root, step.Directory)
		if err != nil {
			return ResolvedStep{}, err
		}
		listing, err := directoryListing(path)
		if err != nil {
			return ResolvedStep{}, err
		}
		return ResolvedStep{Kind: Directory, Path: path, DisplayPath: step.Directory, Source: listing}, nil
	}
	if step.Contents != "" {
		return resolveFileContent(root, step, step.Contents)
	}
	if step.File != "" {
		path, err := safePath(root, step.File)
		if err != nil {
			return ResolvedStep{}, err
		}
		data, err := readSource(path)
		if err != nil {
			return ResolvedStep{Kind: File, Path: path, DisplayPath: step.File}, err
		}
		return resolveFileContent(root, step, string(data))
	}
	if step.URI != "" {
		return ResolvedStep{Kind: URI, DisplayPath: step.URI, Source: step.URI}, nil
	}
	return ResolvedStep{Kind: Content}, nil
}

func resolveFileContent(root string, step tour.Step, source string) (ResolvedStep, error) {
	path := step.File
	if path != "" {
		var err error
		path, err = safePath(root, path)
		if err != nil {
			return ResolvedStep{}, err
		}
	}
	lineCount := 1
	if source != "" {
		lineCount = strings.Count(source, "\n") + 1
	}
	target := step.Line
	start, end := 0, 0
	if step.Selection != nil {
		start = step.Selection.Start.Line
		end = step.Selection.End.Line
		if target == 0 {
			target = step.Selection.End.Line
		}
	}
	if target == 0 && step.Pattern != "" {
		pattern, err := regexp.Compile("(?m)" + step.Pattern)
		if err != nil {
			return ResolvedStep{}, fmt.Errorf("unsupported pattern %q: %w", step.Pattern, err)
		}
		if match := pattern.FindStringIndex(source); match != nil {
			target = strings.Count(source[:match[0]], "\n") + 1
		}
	}
	if target == 0 {
		target = lineCount
	}
	if target < 1 || target > lineCount {
		return ResolvedStep{Kind: File, Path: path, DisplayPath: step.File, Source: source, TargetLine: target},
			fmt.Errorf("line %d is outside %s (1-%d)", target, step.File, lineCount)
	}
	if start > lineCount || end > lineCount {
		return ResolvedStep{}, fmt.Errorf("selection is outside %s (1-%d)", step.File, lineCount)
	}
	return ResolvedStep{
		Kind: File, Path: path, DisplayPath: step.File, Source: source,
		TargetLine: target, SelectionStart: start, SelectionEnd: end,
	}, nil
}

func safePath(root, relative string) (string, error) {
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("absolute file path is not allowed: %s", relative)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	canonicalRoot := rootAbs
	if resolvedRoot, resolveErr := filepath.EvalSymlinks(rootAbs); resolveErr == nil {
		canonicalRoot = resolvedRoot
	}
	joined := filepath.Join(rootAbs, filepath.FromSlash(relative))
	rel, err := filepath.Rel(rootAbs, joined)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes workspace: %s", relative)
	}

	if resolved, err := filepath.EvalSymlinks(joined); err == nil {
		resolvedRel, relErr := filepath.Rel(canonicalRoot, resolved)
		if relErr != nil || resolvedRel == ".." || strings.HasPrefix(resolvedRel, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("symlink escapes workspace: %s", relative)
		}
	}
	return joined, nil
}

func readSource(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("open source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source is not a regular file: %s", path)
	}
	if info.Size() > maxSourceSize {
		return nil, fmt.Errorf("source is larger than %d MiB: %s", maxSourceSize>>20, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}
	if bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) {
		return nil, fmt.Errorf("source is binary or not UTF-8: %s", path)
	}
	return data, nil
}

func directoryListing(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("read directory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})
	const limit = 500
	var result strings.Builder
	for i, entry := range entries {
		if i == limit {
			fmt.Fprintf(&result, "… and %d more\n", len(entries)-limit)
			break
		}
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		result.WriteString(name)
		result.WriteByte('\n')
	}
	return result.String(), nil
}
