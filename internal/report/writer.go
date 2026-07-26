package report

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteAtomic writes report content atomically and refuses to overwrite existing files.
func WriteAtomic(path string, content []byte) error {
	if path == "" {
		return fmt.Errorf("output path is required")
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("output file already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check output path: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".kua-report-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		_ = tmp.Close()
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(content); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("atomic rename failed: %w", err)
	}
	cleanup = false
	return nil
}

// Render renders a document in the chosen format with optional redaction.
func Render(doc Document, opts RenderOptions) ([]byte, error) {
	if opts.Redacted {
		doc = NewRedactor(doc.AssessmentID).RedactDocument(doc)
	}

	switch opts.Format {
	case FormatJSON:
		return RenderJSON(doc)
	case FormatMarkdown:
		return RenderMarkdown(doc)
	case FormatHTML:
		return RenderHTML(doc)
	case FormatConsole, "":
		return RenderConsole(doc)
	default:
		return nil, fmt.Errorf("unsupported format: %s", opts.Format)
	}
}
