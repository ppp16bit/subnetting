package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	accent    = lipgloss.Color("#5EEAD4")
	accentDim = lipgloss.Color("#2DD4BF")
	text      = lipgloss.Color("#F8FAFC")
	muted     = lipgloss.Color("#94A3B8")
	subtle    = lipgloss.Color("#475569")
	surface   = lipgloss.Color("#111827")
	errorRed  = lipgloss.Color("#FB7185")
)

var (
	wordmarkStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(text)

	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(muted)

	eyebrowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(muted)

	badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			Background(lipgloss.Color("#123331")).
			Padding(0, 1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Background(surface).
			Padding(1, 2)

	inputFrameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentDim).
			Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	resultPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(subtle).
				Background(surface).
				Padding(1, 2)

	statStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderForeground(accentDim).
			PaddingLeft(1)

	statLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(muted)

	statValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(text)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorRed).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorRed).
			Padding(0, 1)

	hintStyle = lipgloss.NewStyle().
			Foreground(muted)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(text)

	footerStyle = lipgloss.NewStyle().
			Foreground(subtle)

	stepStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	separatorStyle = lipgloss.NewStyle().
			Foreground(subtle)
)

type panelID string

const (
	learningPanel panelID = "learning"
)

type Model struct {
	input  textinput.Model
	result *SubnetInfo
	err    error
	width  int
	height int
	panels []panelID
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "192.168.1.0/24"
	ti.Focus()
	ti.CharLimit = 18
	ti.Width = 32
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().Bold(true).Foreground(text)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(subtle)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(accent)

	return Model{input: ti}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) calculate() Model {
	val := strings.TrimSpace(m.input.Value())
	if val == "" {
		m.result = nil
		m.err = nil
		return m
	}

	res, err := ParseAndCalculate(val)
	if err != nil {
		m.result = nil
		m.err = err
		return m
	}

	m.result = res
	m.err = nil
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				switch msg.Runes[0] {
				case 'l', 'L':
					m.panels = togglePanel(m.panels, learningPanel)
					return m, nil
				case 'q', 'Q':
					return m, tea.Quit
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = max(6, min(msg.Width-12, 42))
	}

	m.input, cmd = m.input.Update(msg)
	m = m.calculate()
	return m, cmd
}

func (m Model) View() string {
	contentWidth := m.contentWidth()

	sections := []string{
		renderHeader(contentWidth, m.height > 0 && m.height < 27),
		m.renderInput(contentWidth),
		m.renderWorkspace(contentWidth),
	}
	sections = append(sections, renderFooter(contentWidth, m.panels))

	app := lipgloss.JoinVertical(lipgloss.Left, sections...)
	if m.width <= 0 {
		return lipgloss.NewStyle().Margin(1, 2).Render(app)
	}

	left := max(0, (m.width-lipgloss.Width(app))/2)
	top := 0
	if m.height >= 27 {
		top = 1
	}

	return lipgloss.NewStyle().MarginLeft(left).MarginTop(top).Render(app)
}

func (m Model) contentWidth() int {
	if m.width <= 0 {
		if m.hasSidePanels() {
			return 112
		}
		return 78
	}
	preferred := 78
	if m.hasSidePanels() {
		preferred = 112
	}
	return max(18, min(m.width-2, preferred))
}

func renderHeader(width int, compact bool) string {
	brand := logoStyle.Render("◈") + "  " + wordmarkStyle.Render("SUBNET")
	badge := badgeStyle.Render("IPv4")
	gap := max(1, width-lipgloss.Width(brand)-lipgloss.Width(badge))
	top := brand + strings.Repeat(" ", gap) + badge
	if compact || width < 44 {
		return top
	}

	subtitle := subtitleStyle.Render("Turn an address into a clear network map.")

	return lipgloss.JoinVertical(lipgloss.Left, top, subtitle)
}

func (m Model) renderInput(width int) string {
	innerWidth := max(1, width-panelStyle.GetHorizontalFrameSize())
	label := eyebrowStyle.Render("NETWORK INPUT")
	heading := label
	if width >= 30 {
		live := badgeStyle.Render("LIVE")
		gap := max(1, innerWidth-lipgloss.Width(label)-lipgloss.Width(live))
		heading = label + strings.Repeat(" ", gap) + live
	}

	field := promptStyle.Render("› ") + m.input.View()
	field = renderFrame(inputFrameStyle, innerWidth, field)

	parts := []string{heading, "", field}
	if m.height > 0 && m.height < 27 {
		parts = []string{heading, field}
	}
	return renderFrame(panelStyle, width, lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m Model) renderWorkspace(width int) string {
	primary := m.renderFeedback(width)
	if !m.hasSidePanels() {
		return primary
	}

	const gap = 2
	if width < 78 {
		panels := []string{primary}
		panels = append(panels, m.renderSidePanel(width))
		return lipgloss.JoinVertical(lipgloss.Left, panels...)
	}

	primaryWidth := max(44, width*48/100)
	railWidth := width - primaryWidth - gap
	primary = m.renderFeedback(primaryWidth)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		primary,
		strings.Repeat(" ", gap),
		m.renderSidePanel(railWidth),
	)
}

func (m Model) renderFeedback(width int) string {
	if m.err != nil {
		message := "Check the address and use IPv4/CIDR format, for example 10.0.0.1/8."
		if width < 54 {
			message = "Use IPv4/CIDR, for example 10.0.0.1/8."
		}
		detail := lipgloss.NewStyle().Foreground(muted).Render(m.err.Error())
		return renderFrame(errorStyle, width,
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render("Invalid network"),
				message,
				detail,
			),
		)
	}

	if m.result != nil {
		return renderResult(m.result, width)
	}

	empty := hintStyle.Render("Enter any IPv4 address with a CIDR prefix to see its complete range.")
	if width < 60 {
		empty = hintStyle.Render("Enter IPv4/CIDR to calculate.")
	}
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(subtle).
		Padding(1, 2)
	return renderFrame(style, width, "○  "+empty)
}

func (m Model) renderSidePanel(width int) string {
	if m.result == nil {
		message := "Waiting for a valid IPv4/CIDR address."
		if m.err != nil {
			message = "Fix the input to resume the live explanation."
		}
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			eyebrowStyle.Render("LEARNING"),
			"",
			hintStyle.Render(message),
		)
		return renderFrame(resultPanelStyle, width, content)
	}

	compact := m.height > 0 && m.height < 34
	return renderLearning(m.result, width, compact)
}

func (m Model) hasSidePanels() bool {
	return panelEnabled(m.panels, learningPanel)
}

type stat struct {
	label string
	value string
}

func renderResult(r *SubnetInfo, width int) string {
	title := eyebrowStyle.Render("NETWORK DETAILS")
	innerWidth := max(12, width-resultPanelStyle.GetHorizontalFrameSize())
	heading := title
	if width >= 36 {
		valid := badgeStyle.Render("VALID")
		gap := max(1, innerWidth-lipgloss.Width(title)-lipgloss.Width(valid))
		heading = title + strings.Repeat(" ", gap) + valid
	}

	stats := []stat{
		{label: "PROVIDED IP", value: r.IP},
		{label: "PREFIX", value: fmt.Sprintf("/%d", r.CIDR)},
		{label: "SUBNET MASK", value: r.Mask},
		{label: "USABLE HOSTS", value: formatNumber(r.UsableHosts)},
		{label: "NETWORK", value: r.Network},
		{label: "BROADCAST", value: r.Broadcast},
		{label: "FIRST USABLE", value: r.FirstUsable},
		{label: "LAST USABLE", value: r.LastUsable},
	}

	grid := renderStatGrid(stats, innerWidth, width >= 62)
	content := lipgloss.JoinVertical(lipgloss.Left, heading, "", grid)
	return renderFrame(resultPanelStyle, width, content)
}

func renderLearning(r *SubnetInfo, width int, compact bool) string {
	info := Explain(r)
	title := eyebrowStyle.Render("LEARNING")
	live := badgeStyle.Render("LIVE")
	innerWidth := max(8, width-resultPanelStyle.GetHorizontalFrameSize())
	gap := max(1, innerWidth-lipgloss.Width(title)-lipgloss.Width(live))
	heading := title + strings.Repeat(" ", gap) + live
	block := fmt.Sprintf("%d–%d", info.BlockStart, info.BlockEnd)

	var body string
	if compact {
		steps := compactLearningSteps(r, info, block, innerWidth)
		body = lipgloss.JoinVertical(lipgloss.Left, steps...)
	} else {
		separator := separatorStyle.Render(strings.Repeat("─", max(8, innerWidth)))
		steps := []string{
			learningStep(1, "CIDR", fmt.Sprintf("/%d  ↓  %s", r.CIDR, r.Mask)),
			separator,
			learningStep(2, fmt.Sprintf("Interesting octet #%d", info.InterestingOctet), strconv.Itoa(info.MaskOctet)),
			separator,
			learningStep(3, "Increment", fmt.Sprintf("256 - %d = %d", info.MaskOctet, info.Increment)),
			separator,
			learningStep(4, "Subnet starts", formatStarts(info.SubnetStarts)),
			separator,
			learningStep(5, fmt.Sprintf("%d belongs inside", info.IPOctet), block),
			separator,
			learningStep(6, "Network", r.Network),
		}
		body = lipgloss.JoinVertical(lipgloss.Left, steps...)
	}

	return renderFrame(resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, heading, "", body))
}

func compactLearningSteps(r *SubnetInfo, info LearningInfo, block string, width int) []string {
	if width < 42 {
		return []string{
			learningLine(1, "CIDR", fmt.Sprintf("/%d", r.CIDR)),
			statValueStyle.Render("   Mask " + r.Mask),
			learningLine(2, "Octet", fmt.Sprintf("#%d = %d", info.InterestingOctet, info.MaskOctet)),
			learningLine(3, "Increment", fmt.Sprintf("256-%d=%d", info.MaskOctet, info.Increment)),
			learningLine(4, "Starts", formatStarts(info.SubnetStarts)),
			learningLine(5, "Block", fmt.Sprintf("%d in %s", info.IPOctet, block)),
			learningLine(6, "Network", r.Network),
		}
	}

	return []string{
		learningLine(1, "CIDR", fmt.Sprintf("/%d → %s", r.CIDR, r.Mask)),
		learningLine(2, "Interesting octet", fmt.Sprintf("#%d = %d", info.InterestingOctet, info.MaskOctet)),
		learningLine(3, "Increment", fmt.Sprintf("256 - %d = %d", info.MaskOctet, info.Increment)),
		learningLine(4, "Subnet starts", formatStarts(info.SubnetStarts)),
		learningLine(5, "Containing block", fmt.Sprintf("%d belongs in %s", info.IPOctet, block)),
		learningLine(6, "Network", r.Network),
	}
}

func learningStep(number int, label, value string) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		stepStyle.Render(fmt.Sprintf("STEP %d · %s", number, label)),
		statValueStyle.Render(value),
	)
}

func learningLine(number int, label, value string) string {
	return stepStyle.Render(fmt.Sprintf("%d  ", number)) +
		statLabelStyle.Render(label+": ") +
		statValueStyle.Render(value)
}

func formatStarts(starts []int) string {
	values := make([]string, 0, len(starts))
	if len(starts) <= 8 {
		for _, start := range starts {
			values = append(values, strconv.Itoa(start))
		}
		return strings.Join(values, "  ")
	}

	for _, start := range starts[:4] {
		values = append(values, strconv.Itoa(start))
	}
	values = append(values, "…", strconv.Itoa(starts[len(starts)-1]))
	return strings.Join(values, "  ")
}

func renderStatGrid(stats []stat, width int, twoColumns bool) string {
	if !twoColumns {
		rows := make([]string, 0, len(stats))
		for _, item := range stats {
			label := statLabelStyle.Render(item.label)
			value := statValueStyle.Render(item.value)
			gap := max(1, width-lipgloss.Width(label)-lipgloss.Width(value)-2)
			content := label + strings.Repeat(" ", gap) + value
			rows = append(rows, renderFrame(statStyle, width, content))
		}
		return lipgloss.JoinVertical(lipgloss.Left, rows...)
	}

	const columnGap = 2
	columnWidth := (width - columnGap) / 2
	rows := make([]string, 0, len(stats)/2)
	for i := 0; i < len(stats); i += 2 {
		left := renderStat(stats[i], columnWidth)
		right := renderStat(stats[i+1], columnWidth)
		rows = append(rows, lipgloss.JoinHorizontal(
			lipgloss.Top,
			left,
			strings.Repeat(" ", columnGap),
			right,
		))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderStat(item stat, width int) string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		statLabelStyle.Render(item.label),
		statValueStyle.Render(item.value),
	)
	return renderFrame(statStyle, width, content)
}

func renderFooter(width int, panels []panelID) string {
	shortcuts := keyStyle.Render("L") + footerStyle.Render(" learn  ") +
		keyStyle.Render("Q") + footerStyle.Render(" quit")
	if width < 42 {
		return shortcuts
	}

	status := "Calculates as you type"
	if panelEnabled(panels, learningPanel) {
		status = "learning on"
	}
	statusView := footerStyle.Render(status)
	gap := max(1, width-lipgloss.Width(statusView)-lipgloss.Width(shortcuts))
	return statusView + strings.Repeat(" ", gap) + shortcuts
}

func togglePanel(panels []panelID, target panelID) []panelID {
	next := make([]panelID, 0, len(panels)+1)
	found := false
	for _, panel := range panels {
		if panel == target {
			found = true
			continue
		}
		next = append(next, panel)
	}
	if !found {
		next = append(next, target)
	}
	return next
}

func panelEnabled(panels []panelID, target panelID) bool {
	for _, panel := range panels {
		if panel == target {
			return true
		}
	}
	return false
}

func renderFrame(style lipgloss.Style, width int, content string) string {
	return style.Width(max(1, width-style.GetHorizontalBorderSize())).Render(content)
}

func formatNumber(value uint64) string {
	raw := strconv.FormatUint(value, 10)
	for i := len(raw) - 3; i > 0; i -= 3 {
		raw = raw[:i] + "," + raw[i:]
	}
	return raw
}
