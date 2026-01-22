package ui

import (
	"fmt"
	"io"
	"sync"

	"github.com/charmbracelet/glamour"
)

var (
	// cachedRenderer is a cached markdown renderer to avoid recreating it on every call
	cachedRenderer *glamour.TermRenderer
	rendererOnce   sync.Once
	rendererErr    error
)

// getMarkdownRenderer returns a cached markdown renderer instance
// This is thread-safe and creates the renderer only once
// Uses a fixed "dark" style to avoid slow terminal detection (WithAutoStyle takes ~5s)
func getMarkdownRenderer() (*glamour.TermRenderer, error) {
	rendererOnce.Do(func() {
		// Use a fixed "dark" style instead of WithAutoStyle() to avoid slow terminal detection
		// WithAutoStyle() does terminal capability detection which takes ~5 seconds
		// The "dark" style works well in most terminals and is much faster
		cachedRenderer, rendererErr = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(80),
		)
	})
	return cachedRenderer, rendererErr
}

// RenderMarkdown renders markdown text to formatted terminal output
// The renderer is cached for performance
func RenderMarkdown(text string) (string, error) {
	r, err := getMarkdownRenderer()
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
