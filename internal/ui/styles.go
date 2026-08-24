package ui

import "github.com/charmbracelet/lipgloss"

type styleSet struct {
	renderer           *lipgloss.Renderer
	wordmarkStyle      lipgloss.Style
	logoStyle          lipgloss.Style
	subtitleStyle      lipgloss.Style
	eyebrowStyle       lipgloss.Style
	badgeStyle         lipgloss.Style
	successBadgeStyle  lipgloss.Style
	panelStyle         lipgloss.Style
	inputFrameStyle    lipgloss.Style
	promptStyle        lipgloss.Style
	resultPanelStyle   lipgloss.Style
	statStyle          lipgloss.Style
	statLabelStyle     lipgloss.Style
	statValueStyle     lipgloss.Style
	errorStyle         lipgloss.Style
	errorTitleStyle    lipgloss.Style
	errorDetailStyle   lipgloss.Style
	hintStyle          lipgloss.Style
	keyStyle           lipgloss.Style
	footerStyle        lipgloss.Style
	stepStyle          lipgloss.Style
	separatorStyle     lipgloss.Style
	binaryNetworkStyle lipgloss.Style
	binaryHostStyle    lipgloss.Style
	emptyPanelStyle    lipgloss.Style
	inputTextStyle     lipgloss.Style
	placeholderStyle   lipgloss.Style
	cursorStyle        lipgloss.Style
}

func newStyles(renderer *lipgloss.Renderer) styleSet {
	accent := lipgloss.Color("6")
	success := lipgloss.Color("2")
	errorColor := lipgloss.Color("1")

	return styleSet{
		renderer:      renderer,
		wordmarkStyle: renderer.NewStyle().Bold(true),
		logoStyle: renderer.NewStyle().
			Bold(true).
			Foreground(accent),
		subtitleStyle: renderer.NewStyle().Faint(true),
		eyebrowStyle: renderer.NewStyle().
			Bold(true).
			Faint(true),
		badgeStyle: renderer.NewStyle().
			Bold(true).
			Foreground(accent).
			Padding(0, 1),
		successBadgeStyle: renderer.NewStyle().
			Bold(true).
			Foreground(success).
			Padding(0, 1),
		panelStyle: renderer.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2),
		inputFrameStyle: renderer.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 1),
		promptStyle: renderer.NewStyle().
			Bold(true).
			Foreground(accent),
		resultPanelStyle: renderer.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2),
		statStyle: renderer.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderForeground(accent).
			PaddingLeft(1),
		statLabelStyle: renderer.NewStyle().
			Bold(true).
			Faint(true),
		statValueStyle: renderer.NewStyle().Bold(true),
		errorStyle: renderer.NewStyle().
			Foreground(errorColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(0, 1),
		errorTitleStyle:  renderer.NewStyle().Bold(true),
		errorDetailStyle: renderer.NewStyle().Faint(true),
		hintStyle:        renderer.NewStyle().Faint(true),
		keyStyle:         renderer.NewStyle().Bold(true),
		footerStyle:      renderer.NewStyle().Faint(true),
		stepStyle: renderer.NewStyle().
			Bold(true).
			Foreground(accent),
		separatorStyle: renderer.NewStyle().Faint(true),
		binaryNetworkStyle: renderer.NewStyle().
			Bold(true).
			Foreground(accent),
		binaryHostStyle: renderer.NewStyle().Faint(true),
		emptyPanelStyle: renderer.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2),
		inputTextStyle:   renderer.NewStyle().Bold(true),
		placeholderStyle: renderer.NewStyle().Faint(true),
		cursorStyle:      renderer.NewStyle().Foreground(accent),
	}
}
