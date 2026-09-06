package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
)

// Frame widths include borders and padding. Wrap before applying the maximum,
// so the maximum is a guard, not a mechanism for truncating address values.
func renderFrame(style lipgloss.Style, width int, content string) string {
	// Lip Gloss v0.10 stores rules in a shared map. A value assignment alone
	// would let one panel's dimensions leak into every panel using this style.
	style = style.Copy()
	if width < style.GetHorizontalFrameSize()+1 {
		style = style.Padding(0)
	}
	inner := max(1, width-style.GetHorizontalFrameSize())
	return style.Width(max(1, width-style.GetHorizontalBorderSize())).MaxWidth(width).
		Render(wrap.String(content, inner))
}

// Retain the IPv4 side-by-side breakpoint. IPv6 needs enough room for both
// the address column and the explanation; otherwise stack them vertically.
func (m Model) sidePanelThreshold() int {
	if m.result != nil && m.result.IPv4 == nil {
		return 96
	}
	return 78
}

func (m Model) workspaceHeight() int {
	if m.height <= 0 {
		return 24
	}
	footer := wrap.String("L learn  Ctrl+B bits  S subnets  X explain  ? help  Q quit · Ctrl+↑/↓ scroll", m.contentWidth())
	return max(1, m.height-lipgloss.Height(m.topView())-lipgloss.Height(footer))
}
