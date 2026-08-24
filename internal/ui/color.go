package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

func ParseColorMode(value string) (ColorMode, error) {
	mode := ColorMode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case ColorAuto, ColorAlways, ColorNever:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid color mode %q; want auto, always, or never", value)
	}
}

func newRenderer(output io.Writer, mode ColorMode, noColor bool) *lipgloss.Renderer {
	renderer := lipgloss.NewRenderer(output)
	detected := renderer.ColorProfile()
	renderer.SetColorProfile(resolveColorProfile(mode, noColor, detected))
	return renderer
}

func resolveColorProfile(mode ColorMode, noColor bool, detected termenv.Profile) termenv.Profile {
	switch mode {
	case ColorAlways:
		return termenv.ANSI
	case ColorNever:
		return termenv.Ascii
	case ColorAuto:
		if noColor {
			return termenv.Ascii
		}
		return detected
	default:
		return detected
	}
}

func noColorRequested() bool {
	value, exists := os.LookupEnv("NO_COLOR")
	return exists && value != ""
}
