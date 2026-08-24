package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ppp16bit/subnetting"
)

type panelID string

const (
	learningPanel panelID = "learning"
	binaryPanel   panelID = "binary"
	helpPanel     panelID = "help"
)

type Model struct {
	input  textinput.Model
	result *subnetting.SubnetInfo
	err    error
	width  int
	height int
	panels []panelID
	styles styleSet
}

func NewModel() Model {
	return newModel(lipgloss.DefaultRenderer())
}

func newModel(renderer *lipgloss.Renderer) Model {
	styles := newStyles(renderer)
	ti := textinput.New()
	ti.Placeholder = "192.168.1.0/24"
	ti.Focus()
	ti.CharLimit = 18
	ti.Width = 32
	ti.Prompt = ""
	ti.TextStyle = styles.inputTextStyle
	ti.PlaceholderStyle = styles.placeholderStyle
	ti.Cursor.Style = styles.cursorStyle

	return Model{input: ti, styles: styles}
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

	res, err := subnetting.ParseAndCalculate(val)
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
				case 'b', 'B':
					m.panels = togglePanel(m.panels, binaryPanel)
					return m, nil
				case '?':
					m.panels = togglePanel(m.panels, helpPanel)
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
		renderHeader(m.styles, contentWidth, m.height > 0 && m.height < 27),
		m.renderInput(contentWidth),
		m.renderWorkspace(contentWidth),
	}
	if panelEnabled(m.panels, helpPanel) {
		if m.height > 0 && m.height < 32 {
			sections = append(sections, renderCompactHelp(m.styles, contentWidth))
		} else {
			sections = append(sections, renderHelp(m.styles, contentWidth), renderFooter(m.styles, contentWidth, m.panels))
		}
	} else {
		sections = append(sections, renderFooter(m.styles, contentWidth, m.panels))
	}

	app := lipgloss.JoinVertical(lipgloss.Left, sections...)
	if m.width <= 0 {
		return m.styles.renderer.NewStyle().Margin(1, 2).Render(app)
	}

	left := max(0, (m.width-lipgloss.Width(app))/2)
	top := 0
	if m.height >= 27 {
		top = 1
	}

	return m.styles.renderer.NewStyle().MarginLeft(left).MarginTop(top).Render(app)
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

func renderHeader(styles styleSet, width int, compact bool) string {
	brand := styles.logoStyle.Render("◈") + "  " + styles.wordmarkStyle.Render("SUBNET")
	badge := styles.badgeStyle.Render("IPv4")
	gap := max(1, width-lipgloss.Width(brand)-lipgloss.Width(badge))
	top := brand + strings.Repeat(" ", gap) + badge
	if compact || width < 44 {
		return top
	}

	subtitle := styles.subtitleStyle.Render("Turn an address into a clear network map.")

	return lipgloss.JoinVertical(lipgloss.Left, top, subtitle)
}

func (m Model) renderInput(width int) string {
	innerWidth := max(1, width-m.styles.panelStyle.GetHorizontalFrameSize())
	label := m.styles.eyebrowStyle.Render("NETWORK INPUT")
	heading := label
	if width >= 30 {
		live := m.styles.badgeStyle.Render("LIVE")
		gap := max(1, innerWidth-lipgloss.Width(label)-lipgloss.Width(live))
		heading = label + strings.Repeat(" ", gap) + live
	}

	field := m.styles.promptStyle.Render("› ") + m.input.View()
	field = renderFrame(m.styles.inputFrameStyle, innerWidth, field)

	if m.useCombinedLearningPanels() && m.height < 27 && width >= 78 {
		content := m.styles.eyebrowStyle.Render("NETWORK INPUT") + "  " + m.styles.promptStyle.Render("› ") + m.input.View()
		return renderFrame(m.styles.inputFrameStyle, width, content)
	}

	parts := []string{heading, "", field}
	if (m.height > 0 && m.height < 27) || m.useCombinedLearningPanels() {
		parts = []string{heading, field}
	}
	return renderFrame(m.styles.panelStyle, width, lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m Model) renderWorkspace(width int) string {
	primary := m.renderFeedback(width)
	if !m.hasSidePanels() {
		return primary
	}

	const gap = 2
	if width < 78 {
		panels := []string{primary}
		for _, id := range []panelID{learningPanel, binaryPanel} {
			if panelEnabled(m.panels, id) {
				panels = append(panels, m.renderSidePanel(id, width))
			}
		}
		return lipgloss.JoinVertical(lipgloss.Left, panels...)
	}

	primaryWidth := max(44, width*48/100)
	railWidth := width - primaryWidth - gap
	primary = m.renderFeedback(primaryWidth)
	if m.useCombinedLearningPanels() && m.result != nil {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			primary,
			strings.Repeat(" ", gap),
			renderCombinedLearningBinary(m.styles, m.result, railWidth),
		)
	}

	rail := make([]string, 0, 2)
	for _, id := range []panelID{learningPanel, binaryPanel} {
		if panelEnabled(m.panels, id) {
			rail = append(rail, m.renderSidePanel(id, railWidth))
		}
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		primary,
		strings.Repeat(" ", gap),
		lipgloss.JoinVertical(lipgloss.Left, rail...),
	)
}

func (m Model) renderFeedback(width int) string {
	if m.err != nil {
		message := "Check the address and use IPv4/CIDR format, for example 10.0.0.1/8."
		if width < 54 {
			message = "Use IPv4/CIDR, for example 10.0.0.1/8."
		}
		detail := m.styles.errorDetailStyle.Render(m.err.Error())
		return renderFrame(m.styles.errorStyle, width,
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.styles.errorTitleStyle.Render("Invalid network"),
				message,
				detail,
			),
		)
	}

	if m.result != nil {
		return renderResult(m.styles, m.result, width)
	}

	empty := m.styles.hintStyle.Render("Enter any IPv4 address with a CIDR prefix to see its complete range.")
	if width < 60 {
		empty = m.styles.hintStyle.Render("Enter IPv4/CIDR to calculate.")
	}
	return renderFrame(m.styles.emptyPanelStyle, width, "○  "+empty)
}

func (m Model) renderSidePanel(id panelID, width int) string {
	if m.result == nil {
		title := "LEARNING"
		if id == binaryPanel {
			title = "BINARY"
		}
		message := "Waiting for a valid IPv4/CIDR address."
		if m.err != nil {
			message = "Fix the input to resume the live explanation."
		}
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.eyebrowStyle.Render(title),
			"",
			m.styles.hintStyle.Render(message),
		)
		return renderFrame(m.styles.resultPanelStyle, width, content)
	}

	switch id {
	case learningPanel:
		compact := m.height > 0 && (m.height < 34 || panelEnabled(m.panels, binaryPanel))
		return renderLearning(m.styles, m.result, width, compact)
	case binaryPanel:
		return renderBinary(m.styles, m.result, width)
	default:
		return ""
	}
}

func (m Model) hasSidePanels() bool {
	return panelEnabled(m.panels, learningPanel) || panelEnabled(m.panels, binaryPanel)
}

func (m Model) useCombinedLearningPanels() bool {
	return m.height > 0 && m.height < 36 &&
		panelEnabled(m.panels, learningPanel) && panelEnabled(m.panels, binaryPanel)
}

type stat struct {
	label string
	value string
}

func renderResult(styles styleSet, r *subnetting.SubnetInfo, width int) string {
	title := styles.eyebrowStyle.Render("NETWORK DETAILS")
	innerWidth := max(12, width-styles.resultPanelStyle.GetHorizontalFrameSize())
	heading := title
	if width >= 36 {
		valid := styles.successBadgeStyle.Render("VALID")
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

	grid := renderStatGrid(styles, stats, innerWidth, width >= 62)
	content := lipgloss.JoinVertical(lipgloss.Left, heading, "", grid)
	return renderFrame(styles.resultPanelStyle, width, content)
}

func renderLearning(styles styleSet, r *subnetting.SubnetInfo, width int, compact bool) string {
	info := subnetting.Explain(r)
	title := styles.eyebrowStyle.Render("LEARNING")
	live := styles.badgeStyle.Render("LIVE")
	innerWidth := max(8, width-styles.resultPanelStyle.GetHorizontalFrameSize())
	gap := max(1, innerWidth-lipgloss.Width(title)-lipgloss.Width(live))
	heading := title + strings.Repeat(" ", gap) + live
	block := fmt.Sprintf("%d–%d", info.BlockStart, info.BlockEnd)

	var body string
	if compact {
		steps := compactLearningSteps(styles, r, info, block, innerWidth)
		body = lipgloss.JoinVertical(lipgloss.Left, steps...)
	} else {
		separator := styles.separatorStyle.Render(strings.Repeat("─", max(8, innerWidth)))
		steps := []string{
			learningStep(styles, 1, "CIDR", fmt.Sprintf("/%d  ↓  %s", r.CIDR, r.Mask)),
			separator,
			learningStep(styles, 2, fmt.Sprintf("Interesting octet #%d", info.InterestingOctet), strconv.Itoa(info.MaskOctet)),
			separator,
			learningStep(styles, 3, "Increment", fmt.Sprintf("256 - %d = %d", info.MaskOctet, info.Increment)),
			separator,
			learningStep(styles, 4, "Subnet starts", formatStarts(info.SubnetStarts)),
			separator,
			learningStep(styles, 5, fmt.Sprintf("%d belongs inside", info.IPOctet), block),
			separator,
			learningStep(styles, 6, "Network", r.Network),
		}
		body = lipgloss.JoinVertical(lipgloss.Left, steps...)
	}

	return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, heading, "", body))
}

func compactLearningSteps(styles styleSet, r *subnetting.SubnetInfo, info subnetting.LearningInfo, block string, width int) []string {
	if width < 42 {
		return []string{
			learningLine(styles, 1, "CIDR", fmt.Sprintf("/%d", r.CIDR)),
			styles.statValueStyle.Render("   Mask " + r.Mask),
			learningLine(styles, 2, "Octet", fmt.Sprintf("#%d = %d", info.InterestingOctet, info.MaskOctet)),
			learningLine(styles, 3, "Increment", fmt.Sprintf("256-%d=%d", info.MaskOctet, info.Increment)),
			learningLine(styles, 4, "Starts", formatStarts(info.SubnetStarts)),
			learningLine(styles, 5, "Block", fmt.Sprintf("%d in %s", info.IPOctet, block)),
			learningLine(styles, 6, "Network", r.Network),
		}
	}

	return []string{
		learningLine(styles, 1, "CIDR", fmt.Sprintf("/%d → %s", r.CIDR, r.Mask)),
		learningLine(styles, 2, "Interesting octet", fmt.Sprintf("#%d = %d", info.InterestingOctet, info.MaskOctet)),
		learningLine(styles, 3, "Increment", fmt.Sprintf("256 - %d = %d", info.MaskOctet, info.Increment)),
		learningLine(styles, 4, "Subnet starts", formatStarts(info.SubnetStarts)),
		learningLine(styles, 5, "Containing block", fmt.Sprintf("%d belongs in %s", info.IPOctet, block)),
		learningLine(styles, 6, "Network", r.Network),
	}
}

func learningStep(styles styleSet, number int, label, value string) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.stepStyle.Render(fmt.Sprintf("STEP %d · %s", number, label)),
		styles.statValueStyle.Render(value),
	)
}

func learningLine(styles styleSet, number int, label, value string) string {
	return styles.stepStyle.Render(fmt.Sprintf("%d  ", number)) +
		styles.statLabelStyle.Render(label+": ") +
		styles.statValueStyle.Render(value)
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

func renderBinary(styles styleSet, r *subnetting.SubnetInfo, width int) string {
	innerWidth := max(8, width-styles.resultPanelStyle.GetHorizontalFrameSize())
	if innerWidth < 35 {
		rows := []string{
			styles.eyebrowStyle.Render("BINARY AND"),
			compactBinaryRow(styles, "IP", subnetting.BinaryIPv4(r.IP), r.CIDR, innerWidth),
			compactBinaryRow(styles, "MASK", subnetting.BinaryIPv4(r.Mask), r.CIDR, innerWidth),
			compactBinaryRow(styles, "AND", subnetting.BinaryIPv4(r.Network), r.CIDR, innerWidth),
		}
		return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, rows...))
	}

	rows := []string{
		styles.eyebrowStyle.Render("BINARY AND"),
		"",
		binaryRow(styles, "IP", subnetting.BinaryIPv4(r.IP), r.CIDR, innerWidth),
		binaryRow(styles, "MASK", subnetting.BinaryIPv4(r.Mask), r.CIDR, innerWidth),
		styles.separatorStyle.Render(strings.Repeat("─", innerWidth)),
		binaryRow(styles, "NETWORK", subnetting.BinaryIPv4(r.Network), r.CIDR, innerWidth),
		"",
		styles.hintStyle.Render("network bits + host bits"),
	}
	return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func renderCombinedLearningBinary(styles styleSet, r *subnetting.SubnetInfo, width int) string {
	info := subnetting.Explain(r)
	innerWidth := max(8, width-styles.resultPanelStyle.GetHorizontalFrameSize())
	block := fmt.Sprintf("%d–%d", info.BlockStart, info.BlockEnd)
	rows := []string{styles.eyebrowStyle.Render("LEARNING + BINARY")}
	rows = append(rows, compactLearningSteps(styles, r, info, block, innerWidth)...)
	rows = append(rows,
		styles.eyebrowStyle.Render("BINARY AND"),
		compactBinaryRow(styles, "IP", subnetting.BinaryIPv4(r.IP), r.CIDR, innerWidth),
		compactBinaryRow(styles, "MASK", subnetting.BinaryIPv4(r.Mask), r.CIDR, innerWidth),
		compactBinaryRow(styles, "AND", subnetting.BinaryIPv4(r.Network), r.CIDR, innerWidth),
	)
	return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func compactBinaryRow(styles styleSet, label, bits string, prefix, width int) string {
	lines := strings.Split(formatBinary(styles, bits, prefix, width), "\n")
	if len(lines) == 0 {
		return styles.statLabelStyle.Render(label)
	}
	indent := strings.Repeat(" ", 6)
	lines[0] = styles.statLabelStyle.Render(fmt.Sprintf("%-6s", label)) + lines[0]
	for i := 1; i < len(lines); i++ {
		lines[i] = indent + lines[i]
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func binaryRow(styles styleSet, label, bits string, prefix, width int) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.statLabelStyle.Render(label),
		formatBinary(styles, bits, prefix, width),
	)
}

func formatBinary(styles styleSet, bits string, prefix, width int) string {
	octets := strings.Split(bits, ".")
	perLine := 4
	if width < 35 {
		perLine = 2
	}
	if width < 17 {
		perLine = 1
	}

	lines := make([]string, 0, (len(octets)+perLine-1)/perLine)
	for start := 0; start < len(octets); start += perLine {
		end := min(len(octets), start+perLine)
		line := strings.Join(octets[start:end], ".")
		lineBits := (end - start) * 8
		linePrefix := max(0, min(lineBits, prefix-start*8))
		lines = append(lines, colorBinary(styles, line, linePrefix))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func colorBinary(styles styleSet, bits string, prefix int) string {
	plainPosition := 0
	var networkBits, hostBits strings.Builder
	for _, char := range bits {
		if char == '.' {
			if plainPosition <= prefix {
				networkBits.WriteRune(char)
			} else {
				hostBits.WriteRune(char)
			}
			continue
		}
		if plainPosition < prefix {
			networkBits.WriteRune(char)
		} else {
			hostBits.WriteRune(char)
		}
		plainPosition++
	}
	return styles.binaryNetworkStyle.Render(networkBits.String()) + styles.binaryHostStyle.Render(hostBits.String())
}

func renderStatGrid(styles styleSet, stats []stat, width int, twoColumns bool) string {
	if !twoColumns {
		rows := make([]string, 0, len(stats))
		for _, item := range stats {
			label := styles.statLabelStyle.Render(item.label)
			value := styles.statValueStyle.Render(item.value)
			gap := max(1, width-lipgloss.Width(label)-lipgloss.Width(value)-2)
			content := label + strings.Repeat(" ", gap) + value
			rows = append(rows, renderFrame(styles.statStyle, width, content))
		}
		return lipgloss.JoinVertical(lipgloss.Left, rows...)
	}

	const columnGap = 2
	columnWidth := (width - columnGap) / 2
	rows := make([]string, 0, len(stats)/2)
	for i := 0; i < len(stats); i += 2 {
		left := renderStat(styles, stats[i], columnWidth)
		right := renderStat(styles, stats[i+1], columnWidth)
		rows = append(rows, lipgloss.JoinHorizontal(
			lipgloss.Top,
			left,
			strings.Repeat(" ", columnGap),
			right,
		))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderStat(styles styleSet, item stat, width int) string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.statLabelStyle.Render(item.label),
		styles.statValueStyle.Render(item.value),
	)
	return renderFrame(styles.statStyle, width, content)
}

func renderHelp(styles styleSet, width int) string {
	items := []string{
		styles.keyStyle.Render("L") + styles.hintStyle.Render(" toggle learning"),
		styles.keyStyle.Render("B") + styles.hintStyle.Render(" toggle binary"),
		styles.keyStyle.Render("?") + styles.hintStyle.Render(" close help"),
		styles.keyStyle.Render("Q") + styles.hintStyle.Render(" quit"),
	}
	content := strings.Join(items, "   ")
	if width < 70 {
		content = lipgloss.JoinVertical(lipgloss.Left, items...)
	}
	return renderFrame(styles.panelStyle, width, lipgloss.JoinVertical(
		lipgloss.Left,
		styles.eyebrowStyle.Render("KEYBOARD HELP"),
		"",
		content,
		styles.hintStyle.Render("Input stays focused while panels open and close."),
	))
}

func renderCompactHelp(styles styleSet, width int) string {
	if width < 50 {
		return styles.keyStyle.Render("L") + styles.footerStyle.Render(" learn  ") +
			styles.keyStyle.Render("B") + styles.footerStyle.Render(" bits  ") +
			styles.keyStyle.Render("?") + styles.footerStyle.Render(" close  ") +
			styles.keyStyle.Render("Q") + styles.footerStyle.Render(" quit")
	}
	return styles.keyStyle.Render("L") + styles.footerStyle.Render(" toggle learning   ") +
		styles.keyStyle.Render("B") + styles.footerStyle.Render(" toggle binary   ") +
		styles.keyStyle.Render("?") + styles.footerStyle.Render(" close help   ") +
		styles.keyStyle.Render("Q/Esc") + styles.footerStyle.Render(" quit")
}

func renderFooter(styles styleSet, width int, panels []panelID) string {
	if width < 34 {
		return styles.keyStyle.Render("L") + styles.footerStyle.Render("  ") +
			styles.keyStyle.Render("B") + styles.footerStyle.Render("  ") +
			styles.keyStyle.Render("?") + styles.footerStyle.Render("  ") +
			styles.keyStyle.Render("Q") + styles.footerStyle.Render(" quit")
	}
	shortcuts := styles.keyStyle.Render("L") + styles.footerStyle.Render(" learn  ") +
		styles.keyStyle.Render("B") + styles.footerStyle.Render(" binary  ") +
		styles.keyStyle.Render("?") + styles.footerStyle.Render(" help  ") +
		styles.keyStyle.Render("Q") + styles.footerStyle.Render(" quit")
	if width < 54 {
		return shortcuts
	}

	status := "Calculates as you type"
	active := make([]string, 0, 2)
	if panelEnabled(panels, learningPanel) {
		active = append(active, "learning")
	}
	if panelEnabled(panels, binaryPanel) {
		active = append(active, "binary")
	}
	if len(active) > 0 {
		status = strings.Join(active, " + ") + " on"
	}
	statusView := styles.footerStyle.Render(status)
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
