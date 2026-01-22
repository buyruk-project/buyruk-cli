package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

var (
	// cachedRenderer is a cached markdown renderer to avoid recreating it on every call
	cachedRenderer *glamour.TermRenderer
	rendererOnce   sync.Once
	rendererErr    error
)

// getTerminalWidth detects the terminal width or returns a default
// Priority: BUYRUK_TERM_WIDTH env var > terminal detection > default (80)
// Width is clamped between 40 and 200 to prevent issues with extreme terminal sizes
func getTerminalWidth() int {
	// Check environment variable first (for testing/override)
	if widthStr := os.Getenv("BUYRUK_TERM_WIDTH"); widthStr != "" {
		if width, err := strconv.Atoi(widthStr); err == nil && width > 0 {
			// Clamp environment variable width too
			if width < 40 {
				return 40
			}
			if width > 200 {
				return 200
			}
			return width
		}
	}

	// Try to detect terminal width from stdout
	// This will fail gracefully if stdout is not a terminal (e.g., when piping)
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		// Use terminal width, but ensure minimum of 40 and maximum of 200
		// This prevents issues with very narrow or very wide terminals
		if width < 40 {
			return 40
		}
		if width > 200 {
			return 200
		}
		return width
	}

	// Fallback to default width (used when terminal detection fails or stdout is not a terminal)
	return 80
}

// getMarkdownRenderer returns a cached markdown renderer instance
// This is thread-safe and creates the renderer only once
// Uses a fixed "dark" style to avoid slow terminal detection (WithAutoStyle takes ~5s)
// Word wrap width is detected dynamically from terminal size
func getMarkdownRenderer() (*glamour.TermRenderer, error) {
	rendererOnce.Do(func() {
		// Detect terminal width dynamically
		wordWrap := getTerminalWidth()

		// Use a fixed "dark" style instead of WithAutoStyle() to avoid slow terminal detection
		// WithAutoStyle() does terminal capability detection which takes ~5 seconds
		// The "dark" style works well in most terminals and is much faster
		cachedRenderer, rendererErr = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(wordWrap),
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
