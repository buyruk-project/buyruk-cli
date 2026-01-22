package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/glamour"
)

// RenderMarkdown renders markdown text to formatted terminal output
func RenderMarkdown(text string) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return "", fmt.Errorf("ui: failed to create markdown renderer: %w", err)
	}

	rendered, err := r.Render(text)
	if err != nil {
		return "", fmt.Errorf("ui: failed to render markdown: %w", err)
	}

	return rendered, nil
}

// RenderMarkdownToWriter renders markdown text directly to a writer
func RenderMarkdownToWriter(text string, w io.Writer) error {
	rendered, err := RenderMarkdown(text)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s", rendered)
	return nil
}
