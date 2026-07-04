package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginLeft(2)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	resultBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#1F1F1F")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginLeft(2).
			MarginTop(1).
			Width(56)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555")).
			MarginLeft(2).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			MarginLeft(2).
			MarginTop(1)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginLeft(2).
			MarginTop(1)
)

type Model struct {
	input  textinput.Model
	result *SubnetInfo
	err    error
	width  int
	heigth int
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "192.168.1.0/24"
	ti.Focus()
	ti.CharLimit = 18
	ti.Width = 30
	ti.Prompt = "Enter IP/CIDR: "
	ti.PromptStyle = promptStyle

	return Model{
		input: ti,
	}
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
		m.heigth = msg.Height
	}

	m.input, cmd = m.input.Update(msg)
	m = m.calculate()
	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Subnetting Calculator"))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
	} else if m.result != nil {
		b.WriteString(renderResult(m.result))
	} else {
		b.WriteString(hintStyle.Render("Start typing to calculate in real time... "))
	}

	return b.String()
}

func renderResult(r *SubnetInfo) string {
	lines := []string{
		labelStyle.Render("Provided IP:    ") + fmt.Sprintf("%s/%d", r.IP, r.CIDR),
		labelStyle.Render("Subnet Mask:   ") + r.Mask,
		labelStyle.Render("Network:       ") + r.Network,
		labelStyle.Render("Broadcast:     ") + r.Broadcast,
		labelStyle.Render("First Usable:  ") + r.FirstUsable,
		labelStyle.Render("Last Usable:   ") + r.LastUsable,
		labelStyle.Render("Usable Hosts   ") + fmt.Sprintf("%d", r.UsableHosts),
	}

	content := strings.Join(lines, "\n")
	return resultBoxStyle.Render(content)
}
