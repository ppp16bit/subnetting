package ui

import (
	"math/big"
	"net/netip"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
	"github.com/ppp16bit/subnetting/internal/network"
)

type panelID string

const (
	learningPanel panelID = "learning"
	binaryPanel   panelID = "binary"
	helpPanel     panelID = "help"
	subnetPanel   panelID = "subnets"
)

type Model struct {
	input          textinput.Model
	result         *network.NetworkInfo
	err            error
	width          int
	height         int
	panels         []panelID
	styles         styleSet
	target         textinput.Model
	focus          int
	expanded       bool
	offset         *big.Int
	split          *network.Split
	splitErr       error
	page           []netip.Prefix
	viewport       viewport.Model
	subnetViewport viewport.Model
}

func NewModel() Model {
	return newModel(lipgloss.DefaultRenderer())
}

func newModel(renderer *lipgloss.Renderer) Model {
	styles := newStyles(renderer)
	ti := textinput.New()
	ti.Placeholder = "IPv4 or IPv6 CIDR, e.g. 192.168.1.0/24 or fd00:1::/64"
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 32
	ti.Prompt = ""
	ti.TextStyle = styles.inputTextStyle
	ti.PlaceholderStyle = styles.placeholderStyle
	ti.Cursor.Style = styles.cursorStyle

	target := textinput.New()
	target.Placeholder = "/64"
	target.CharLimit = 4
	target.Width = 6
	target.Prompt = "› "
	target.TextStyle = styles.inputTextStyle
	target.Cursor.Style = styles.cursorStyle
	return Model{input: ti, target: target, styles: styles, offset: new(big.Int), viewport: viewport.New(78, 20), subnetViewport: viewport.New(72, 8)}
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

	res, err := network.ParseNetwork(val)
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
	oldFocus := m.focus
	oldInput, oldTarget := m.input.Value(), m.target.Value()
	handled := false
	var reveal panelID
	resized := false
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		resized = true
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlB:
			m.panels = togglePanel(m.panels, binaryPanel)
			handled = true
		case tea.KeyTab, tea.KeyShiftTab:
			if panelEnabled(m.panels, subnetPanel) {
				step := 1
				if msg.Type == tea.KeyShiftTab {
					step = 2
				}
				m.focus = (m.focus + step) % 3
				m, cmd = m.applyFocus()
			}
			handled = true
		case tea.KeyPgUp, tea.KeyPgDown:
			if m.focus == 2 && m.split != nil && m.splitErr == nil {
				m = m.turnPage(msg.Type == tea.KeyPgDown)
			} else {
				m = m.refreshViewport()
				if msg.Type == tea.KeyPgDown {
					m.viewport.ViewDown()
				} else {
					m.viewport.ViewUp()
				}
			}
			handled = true
		case tea.KeyUp, tea.KeyDown:
			if m.focus == 2 {
				m = m.refreshSubnetViewport(m.contentWidth())
				if msg.Type == tea.KeyDown {
					m.subnetViewport.LineDown(1)
				} else {
					m.subnetViewport.LineUp(1)
				}
				handled = true
			}
		case tea.KeyHome, tea.KeyEnd:
			if m.focus == 2 && m.split != nil && m.splitErr == nil {
				m.offset = new(big.Int)
				m.subnetViewport.GotoTop()
				if msg.Type == tea.KeyEnd {
					last := new(big.Int).Sub(m.split.Count(), big.NewInt(1))
					m.offset.Mul(last.Quo(last, big.NewInt(subnetPageSize)), big.NewInt(subnetPageSize))
				}
				handled = true
			}
		case tea.KeyCtrlUp, tea.KeyCtrlDown:
			m = m.refreshViewport()
			if msg.Type == tea.KeyCtrlDown {
				m.viewport.LineDown(1)
			} else {
				m.viewport.LineUp(1)
			}
			handled = true
		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				switch msg.Runes[0] {
				case 'l', 'L':
					m.panels = togglePanel(m.panels, learningPanel)
					handled = true
				case 'b', 'B':
					address, _, _ := strings.Cut(m.input.Value(), "/")
					if addr, err := netip.ParseAddr(address); err == nil && addr.Is4() {
						m.panels = togglePanel(m.panels, binaryPanel)
						handled = true
					}
				case 's', 'S':
					m.panels = togglePanel(m.panels, subnetPanel)
					if panelEnabled(m.panels, subnetPanel) {
						reveal = subnetPanel
					}
					handled = true
					if !panelEnabled(m.panels, subnetPanel) {
						m.focus = 0
						m, cmd = m.applyFocus()
					}
					if m.target.Value() == "" && m.result != nil {
						bits := m.result.Prefix.Bits()
						target := min(bits+1, m.result.Input.BitLen())
						if m.result.IPv4 == nil && bits < 64 {
							target = 64
						}
						m.target.SetValue("/" + strconv.Itoa(target))
					}
				case 'x', 'X':
					m.expanded = !m.expanded
					if m.expanded && m.result != nil && m.result.IPv4 == nil {
						if !panelEnabled(m.panels, learningPanel) {
							m.panels = togglePanel(m.panels, learningPanel)
						}
						reveal = learningPanel
					}
					handled = true
				case '?':
					m.panels = togglePanel(m.panels, helpPanel)
					if panelEnabled(m.panels, helpPanel) {
						reveal = helpPanel
					} else {
						m.viewport.GotoTop()
					}
					handled = true
				case 'q', 'Q':
					return m, tea.Quit
				}
			}
		}
	}
	if !handled {
		switch m.focus {
		case 0:
			m.input, cmd = m.input.Update(msg)
		case 1:
			m.target, cmd = m.target.Update(msg)
		}
	}
	if m.input.Value() != oldInput {
		m = m.calculate()
		m.offset = new(big.Int)
		m.viewport.GotoTop()
		m.subnetViewport.GotoTop()
	}
	if m.target.Value() != oldTarget {
		m.offset = new(big.Int)
		m.subnetViewport.GotoTop()
	}
	m = m.calculateSplit()
	m = m.refreshViewport()
	if (m.focus != oldFocus || resized) && m.focus != 0 && panelEnabled(m.panels, subnetPanel) {
		top := lipgloss.Height(m.renderWorkspace(m.contentWidth()))
		if m.focus == 2 {
			_, _, header, _ := m.subnetChrome(m.contentWidth())
			top += lipgloss.Height(header) + 1
		}
		if top < m.viewport.YOffset || top >= m.viewport.YOffset+m.viewport.Height {
			m.viewport.SetYOffset(top)
		}
	}
	if reveal != "" {
		top := lipgloss.Height(m.renderWorkspace(m.contentWidth()))
		if reveal == helpPanel && panelEnabled(m.panels, subnetPanel) {
			top += lipgloss.Height(m.renderSubnets(m.contentWidth()))
		} else if reveal == learningPanel {
			top = 0
			if m.contentWidth() < m.sidePanelThreshold() {
				top = lipgloss.Height(m.renderFeedback(m.contentWidth()))
			}
		}
		m.viewport.SetYOffset(top)
	}
	return m, cmd
}

func (m Model) applyFocus() (Model, tea.Cmd) {
	m.input.Blur()
	m.target.Blur()
	if m.focus == 0 {
		m.viewport.GotoTop()
		return m, m.input.Focus()
	}
	if m.focus == 1 {
		return m, m.target.Focus()
	}
	return m, nil
}

func (m Model) family() string {
	if m.result != nil {
		if m.result.IPv4 != nil {
			return "IPv4"
		}
		return "IPv6"
	}
	return "IPv4 / IPv6"
}

func (m Model) topView() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		renderHeader(m.styles, m.contentWidth(), m.height > 0 && m.height < 27, m.family()),
		m.renderInput(m.contentWidth()))
}

func (m Model) bodyView() string {
	parts := []string{m.renderWorkspace(m.contentWidth())}
	if panelEnabled(m.panels, subnetPanel) {
		parts = append(parts, m.renderSubnets(m.contentWidth()))
	}
	if panelEnabled(m.panels, helpPanel) {
		parts = append(parts, renderHelp(m.styles, m.contentWidth()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) footerView() string {
	text := "L learn  Ctrl+B bits  S subnets  X explain  ? help  Q quit"
	if m.contentWidth() < 62 {
		text = "L Ctrl+B S X ? Q | PgUp/Dn scroll"
	}
	if m.contentWidth() < 38 {
		text = "? help · Q quit"
	}
	if m.viewport.TotalLineCount() > m.viewport.Height {
		text += " · Ctrl+↑/↓ scroll"
	}
	return wrap.String(m.styles.footerStyle.Render(text), m.contentWidth())
}

func (m Model) refreshViewport() Model {
	m.input = m.sizedInput(m.contentWidth())
	m = m.refreshSubnetViewport(m.contentWidth())
	m.viewport.Width = m.contentWidth()
	m.viewport.SetContent(m.bodyView())
	m.viewport.Height = m.viewport.TotalLineCount()
	if m.height > 0 {
		m.viewport.Height = m.workspaceHeight()
	}
	m.viewport.SetYOffset(m.viewport.YOffset)
	return m
}

func (m Model) View() string {
	if m.width > 0 && (m.width < 18 || (m.height > 0 && m.height < 12)) {
		lines := strings.Split(wrap.String("Resize terminal (18×12 minimum). Q quits.", max(1, m.width)), "\n")
		if m.height > 0 {
			lines = lines[:min(len(lines), m.height)]
		}
		return strings.Join(lines, "\n")
	}
	m = m.refreshViewport()
	app := lipgloss.JoinVertical(lipgloss.Left, m.topView(), m.viewport.View(), m.footerView())
	if m.width <= 0 {
		return app
	}
	return m.styles.renderer.NewStyle().MarginLeft(max(0, (m.width-lipgloss.Width(app))/2)).Render(app)
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

func renderHeader(styles styleSet, width int, compact bool, family string) string {
	brand := styles.logoStyle.Render("◈") + "  " + styles.wordmarkStyle.Render("SUBNET")
	if width < 28 && family == "IPv4 / IPv6" {
		family = "IP"
	}
	badge := styles.badgeStyle.Render(family)
	gap := max(1, width-lipgloss.Width(brand)-lipgloss.Width(badge))
	top := brand + strings.Repeat(" ", gap) + badge
	if compact || width < 44 {
		return top
	}

	subtitle := styles.subtitleStyle.Render("Turn an address into a clear network map.")

	return lipgloss.JoinVertical(lipgloss.Left, top, subtitle)
}

func (m Model) compactInput(width int) bool { return width < 40 || (m.height > 0 && m.height < 27) }

func (m Model) sizedInput(width int) textinput.Model {
	field := m.input
	available := width - m.styles.inputFrameStyle.GetHorizontalFrameSize() - 3 // prompt and cursor
	if !m.compactInput(width) {
		available -= m.styles.panelStyle.GetHorizontalFrameSize()
	}
	field.Width = max(1, available)
	field.SetCursor(field.Position()) // Recompute horizontal scrolling after resize.
	return field
}

func (m Model) renderInput(width int) string {
	field := m.sizedInput(width)
	content := m.styles.promptStyle.Render("› ") + field.View()
	if m.compactInput(width) {
		return renderFrame(m.styles.inputFrameStyle, width, content)
	}
	innerWidth := max(1, width-m.styles.panelStyle.GetHorizontalFrameSize())
	label := m.styles.eyebrowStyle.Render("NETWORK INPUT")
	live := m.styles.badgeStyle.Render("LIVE")
	heading := label + strings.Repeat(" ", max(1, innerWidth-lipgloss.Width(label)-lipgloss.Width(live))) + live
	input := renderFrame(m.styles.inputFrameStyle, innerWidth, content)
	parts := []string{heading, "", input}
	if m.useCombinedLearningPanels() {
		parts = []string{heading, input}
	}
	return renderFrame(m.styles.panelStyle, width, lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m Model) renderWorkspace(width int) string {
	primary := m.renderFeedback(width)
	if !m.hasSidePanels() {
		return primary
	}

	const gap = 2
	if width < m.sidePanelThreshold() {
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
			renderCombinedLearningBinary(m.styles, m.result, railWidth, m.expanded),
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
		message := "Check the address and use IP/CIDR format, for example 10.0.0.1/8."
		if width < 54 {
			message = "Use IP/CIDR, for example 10.0.0.1/8."
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

	empty := m.styles.hintStyle.Render("Enter an IPv4 or IPv6 address with a CIDR prefix to see its complete range.")
	if width < 60 {
		empty = m.styles.hintStyle.Render("Enter IP/CIDR to calculate.")
	}
	return renderFrame(m.styles.emptyPanelStyle, width, "○  "+empty)
}

func (m Model) renderSidePanel(id panelID, width int) string {
	if m.result == nil {
		title := "LEARNING"
		if id == binaryPanel {
			title = "BINARY"
		}
		message := "Waiting for a valid IP/CIDR address."
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
		return renderLearning(m.styles, m.result, width, compact, m.expanded)
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

func renderHelp(styles styleSet, width int) string {
	items := []string{
		styles.keyStyle.Render("L") + styles.hintStyle.Render(" toggle learning"),
		styles.keyStyle.Render("Ctrl+B") + styles.hintStyle.Render(" toggle binary"),
		styles.keyStyle.Render("S") + styles.hintStyle.Render(" toggle subnets"),
		styles.keyStyle.Render("X") + styles.hintStyle.Render(" expanded IPv6 in learning"),
		styles.keyStyle.Render("Tab / Shift+Tab") + styles.hintStyle.Render(" focus subnet controls"),
		styles.keyStyle.Render("PgUp / PgDn") + styles.hintStyle.Render(" scroll; change page in subnet results"),
		styles.keyStyle.Render("↑ / ↓") + styles.hintStyle.Render(" scroll rows in subnet results"),
		styles.keyStyle.Render("Home / End") + styles.hintStyle.Render(" first/last subnet page in results"),
		styles.keyStyle.Render("Ctrl+↑ / Ctrl+↓") + styles.hintStyle.Render(" scroll one line"),
		styles.keyStyle.Render("?") + styles.hintStyle.Render(" close help"),
		styles.keyStyle.Render("Q") + styles.hintStyle.Render(" quit"),
	}
	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	return renderFrame(styles.panelStyle, width, lipgloss.JoinVertical(
		lipgloss.Left,
		styles.eyebrowStyle.Render("KEYBOARD HELP"),
		"",
		content,
		styles.hintStyle.Render("Learning and binary panels keep input focused. B remains an alias with IPv4 input."),
	))
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

func formatNumber(value uint64) string {
	raw := strconv.FormatUint(value, 10)
	for i := len(raw) - 3; i > 0; i -= 3 {
		raw = raw[:i] + "," + raw[i:]
	}
	return raw
}
