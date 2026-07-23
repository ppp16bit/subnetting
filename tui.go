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
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Background(surface).
			Padding(1, 2)

	inputFrameStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(accentDim).
			Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	resultPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
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
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(errorRed).
			Padding(0, 1)

	hintStyle = lipgloss.NewStyle().
			Foreground(muted)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(text)

	footerStyle = lipgloss.NewStyle().
			Foreground(subtle)
)

type Model struct {
	input  textinput.Model
	result *SubnetInfo
	err    error
	width  int
	height int
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
		m.renderFeedback(contentWidth),
		renderFooter(contentWidth),
	}

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
		return 74
	}
	return max(18, min(m.width-2, 78))
}

func renderHeader(width int, compact bool) string {
	brand := logoStyle.Render("◈") + "  " + wordmarkStyle.Render("SUBNET")
	badge := badgeStyle.Render("IPv4")
	gap := max(1, width-lipgloss.Width(brand)-lipgloss.Width(badge))
	top := brand + strings.Repeat(" ", gap) + badge
	if compact {
		return top
	}

	subtitle := subtitleStyle.Render("Turn an address into a clear network map.")

	return lipgloss.JoinVertical(lipgloss.Left, top, subtitle)
}

func (m Model) renderInput(width int) string {
	label := eyebrowStyle.Render("NETWORK INPUT")
	heading := label
	if width >= 30 {
		live := badgeStyle.Render("LIVE")
		gap := max(1, width-4-lipgloss.Width(label)-lipgloss.Width(live))
		heading = label + strings.Repeat(" ", gap) + live
	}

	field := promptStyle.Render("› ") + m.input.View()
	fieldWidth := max(10, width-6)
	field = inputFrameStyle.Width(fieldWidth).Render(field)

	return panelStyle.Width(width).Render(
		lipgloss.JoinVertical(lipgloss.Left, heading, "", field),
	)
}

func (m Model) renderFeedback(width int) string {
	if m.err != nil {
		message := "Check the address and use IPv4/CIDR format, for example 10.0.0.1/8."
		detail := lipgloss.NewStyle().Foreground(muted).Render(m.err.Error())
		return errorStyle.Width(width - 2).Render(
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
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(subtle).
		Padding(1, 2).
		Width(width).
		Render("○  " + empty)
}

type stat struct {
	label string
	value string
}

func renderResult(r *SubnetInfo, width int) string {
	title := eyebrowStyle.Render("NETWORK DETAILS")
	innerWidth := max(12, width-6)
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
	return resultPanelStyle.Width(width).Render(content)
}

func renderStatGrid(stats []stat, width int, twoColumns bool) string {
	if !twoColumns {
		rows := make([]string, 0, len(stats))
		for _, item := range stats {
			label := statLabelStyle.Render(item.label)
			value := statValueStyle.Render(item.value)
			gap := max(1, width-lipgloss.Width(label)-lipgloss.Width(value)-2)
			content := label + strings.Repeat(" ", gap) + value
			rows = append(rows, statStyle.Width(width).Render(content))
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
	return statStyle.Width(width).Render(content)
}

func renderFooter(width int) string {
	help := keyStyle.Render("esc") + footerStyle.Render(" quit")
	if width < 36 {
		return help
	}
	status := footerStyle.Render("Calculates as you type")
	gap := max(1, width-lipgloss.Width(status)-lipgloss.Width(help))
	return status + strings.Repeat(" ", gap) + help
}

func formatNumber(value uint64) string {
	raw := strconv.FormatUint(value, 10)
	for i := len(raw) - 3; i > 0; i -= 3 {
		raw = raw[:i] + "," + raw[i:]
	}
	return raw
}
