package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

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
	start := time.Now()
	rendererOnce.Do(func() {
		debugLog("MARKDOWN: Creating renderer (first time)...")
		createStart := time.Now()

		// Use a fixed "dark" style instead of WithAutoStyle() to avoid slow terminal detection
		// WithAutoStyle() does terminal capability detection which takes ~5 seconds
		// The "dark" style works well in most terminals and is much faster
		cachedRenderer, rendererErr = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(80),
		)
		createDuration := time.Since(createStart)
		if rendererErr != nil {
			debugLog(fmt.Sprintf("MARKDOWN: Renderer creation FAILED after %v: %v", createDuration, rendererErr))
		} else {
			debugLog(fmt.Sprintf("MARKDOWN: Renderer created in %v (using fixed dark style)", createDuration))
		}
	})
	totalDuration := time.Since(start)
	if totalDuration > 10*time.Millisecond {
		debugLog(fmt.Sprintf("MARKDOWN: getMarkdownRenderer took %v (cached: %v)", totalDuration, cachedRenderer != nil))
	}
	return cachedRenderer, rendererErr
}

// debugLog logs debug messages if BUYRUK_DEBUG is set
func debugLog(msg string) {
	if os.Getenv("BUYRUK_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", msg)
	}
}

// RenderMarkdown renders markdown text to formatted terminal output
// The renderer is cached for performance
func RenderMarkdown(text string) (string, error) {
	start := time.Now()
	debugLog(fmt.Sprintf("MARKDOWN: RenderMarkdown called (text length: %d)", len(text)))

	rendererStart := time.Now()
	r, err := getMarkdownRenderer()
	rendererDuration := time.Since(rendererStart)
	if rendererDuration > 10*time.Millisecond {
		debugLog(fmt.Sprintf("MARKDOWN: getMarkdownRenderer took %v", rendererDuration))
	}
	if err != nil {
		return "", fmt.Errorf("ui: failed to create markdown renderer: %w", err)
	}

	renderStart := time.Now()
	rendered, err := r.Render(text)
	renderDuration := time.Since(renderStart)
	debugLog(fmt.Sprintf("MARKDOWN: r.Render() took %v (output length: %d)", renderDuration, len(rendered)))
	if err != nil {
		return "", fmt.Errorf("ui: failed to render markdown: %w", err)
	}

	totalDuration := time.Since(start)
	debugLog(fmt.Sprintf("MARKDOWN: RenderMarkdown total time: %v", totalDuration))
	if totalDuration > 100*time.Millisecond {
		debugLog(fmt.Sprintf("MARKDOWN: ⚠️  SLOW: RenderMarkdown took %v", totalDuration))
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
