package ui

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
	"github.com/ppp16bit/subnetting/internal/network"
)

const subnetPageSize = 20

func (m Model) calculateSplit() Model {
	m.split, m.splitErr, m.page = nil, nil, nil
	if !panelEnabled(m.panels, subnetPanel) || m.result == nil {
		return m
	}
	bits, err := strconv.Atoi(strings.TrimPrefix(m.target.Value(), "/"))
	if err != nil {
		m.splitErr = fmt.Errorf("enter a child prefix between /%d and /%d", m.result.Prefix.Bits(), m.result.Input.BitLen())
		return m
	}
	m.split, m.splitErr = network.NewSplit(m.result.Prefix, bits)
	if m.splitErr == nil {
		m.page, m.splitErr = m.split.Page(m.offset, subnetPageSize)
	}
	return m
}

func (m Model) turnPage(forward bool) Model {
	next := new(big.Int).Set(m.offset)
	if forward {
		next.Add(next, big.NewInt(subnetPageSize))
	} else {
		next.Sub(next, big.NewInt(subnetPageSize))
	}
	if next.Sign() < 0 {
		next.SetInt64(0)
	}
	if next.Cmp(m.split.Count()) < 0 {
		m.offset = next
		m.subnetViewport.GotoTop()
	}
	return m
}

// Subnet chrome depends on terminal size, not the length or number of rows.
func (m Model) subnetChrome(width int) (lipgloss.Style, int, string, string) {
	style := m.styles.resultPanelStyle.Copy().Padding(0, 2)
	if width < style.GetHorizontalFrameSize()+1 {
		style = style.Padding(0)
	}
	inner := max(1, width-style.GetHorizontalFrameSize())
	title := "SUBNETS"
	if m.focus == 2 {
		title = "SUBNETS [*]"
	}
	field := "CHILD PREFIX  " + m.target.View()
	if inner < 25 {
		field = "CHILD PREFIX\n" + m.target.View()
	}
	header := wrap.String(m.styles.eyebrowStyle.Render(title)+"\n"+field, inner)
	footer := wrap.String(m.styles.hintStyle.Render("↑/↓ rows · PgUp/Dn page"), inner)
	return style, inner, header, footer
}

func (m Model) subnetRows(width int) string {
	var rows []string
	if m.result == nil {
		rows = append(rows, "Enter a valid IP/CIDR network first.")
	} else if m.splitErr != nil {
		rows = append(rows, m.styles.errorDetailStyle.Render(m.splitErr.Error()))
	} else if m.split != nil {
		rows = append(rows, "TOTAL SUBNETS  "+formatDecimal(m.split.Count().String()),
			"ADDRESSES EACH  "+formatDecimal(m.split.AddressesPerSubnet().String()))
		index := new(big.Int).Set(m.offset)
		for _, child := range m.page {
			address := child.String() // Always compressed in the working list.
			entry := index.String() + "  " + address
			if len(entry) > width {
				entry = index.String() + "\n" + address
			}
			rows = append(rows, entry)
			index.Add(index, big.NewInt(1))
		}
		if index.Cmp(m.split.Count()) < 0 {
			rows = append(rows, "More subnets available · PgDn in results")
		}
	}
	return wrap.String(strings.Join(rows, "\n"), width)
}

func (m Model) refreshSubnetViewport(width int) Model {
	style, inner, header, footer := m.subnetChrome(width)
	chrome := style.GetVerticalFrameSize() + lipgloss.Height(header) + lipgloss.Height(footer)
	m.subnetViewport.Width = inner
	m.subnetViewport.Height = max(1, min(18, m.workspaceHeight())-chrome)
	m.subnetViewport.SetContent(m.subnetRows(inner))
	m.subnetViewport.SetYOffset(m.subnetViewport.YOffset)
	return m
}

func (m Model) renderSubnets(width int) string {
	m = m.refreshSubnetViewport(width)
	style, _, header, footer := m.subnetChrome(width)
	height := style.GetVerticalFrameSize() + lipgloss.Height(header) + m.subnetViewport.Height + lipgloss.Height(footer)
	content := lipgloss.JoinVertical(lipgloss.Left, header, m.subnetViewport.View(), footer)
	return renderFrame(style.Height(height-style.GetVerticalBorderSize()).MaxHeight(height), width, content)
}
