package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ppp16bit/subnetting/internal/network"
)

type stat struct {
	label string
	value string
}

func renderIPv4Result(styles styleSet, r *network.SubnetInfo, width int) string {
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

func renderStatGrid(styles styleSet, stats []stat, width int, twoColumns bool) string {
	columnWidth := (width - 2) / 2
	for _, item := range stats {
		if lipgloss.Width(item.value)+styles.statStyle.GetHorizontalFrameSize() > columnWidth {
			twoColumns = false
		}
	}
	if !twoColumns {
		rows := make([]string, 0, len(stats))
		for _, item := range stats {
			label := styles.statLabelStyle.Render(item.label)
			value := styles.statValueStyle.Render(item.value)
			gap := max(1, width-lipgloss.Width(label)-lipgloss.Width(value)-2)
			content := label + strings.Repeat(" ", gap) + value
			if lipgloss.Width(label)+lipgloss.Width(value)+3 > width {
				content = lipgloss.JoinVertical(lipgloss.Left, label, value)
			}
			rows = append(rows, renderFrame(styles.statStyle, width, content))
		}
		return lipgloss.JoinVertical(lipgloss.Left, rows...)
	}

	const columnGap = 2
	columnWidth = (width - columnGap) / 2
	rows := make([]string, 0, len(stats)/2)
	for i := 0; i < len(stats); i += 2 {
		left := renderStat(styles, stats[i], columnWidth)
		right := ""
		if i+1 < len(stats) {
			right = renderStat(styles, stats[i+1], columnWidth)
		}
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

func renderResult(styles styleSet, n *network.NetworkInfo, width int) string {
	if n.IPv4 != nil {
		return renderIPv4Result(styles, n.IPv4, width)
	}
	stats := []stat{
		{"PROVIDED IP", n.Input.String()},
		{"PREFIX", fmt.Sprintf("/%d", n.Prefix.Bits())},
		{"NETWORK", fmt.Sprintf("%s/%d", n.First.String(), n.Prefix.Bits())},
		{"TOTAL ADDRESSES", formatDecimal(n.AddressCount().String())},
		{"FIRST ADDRESS", n.First.String()},
		{"LAST ADDRESS", n.Last.String()},
	}
	inner := max(1, width-styles.resultPanelStyle.GetHorizontalFrameSize())
	grid := renderStatGrid(styles, stats, inner, width >= 100)
	return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left,
		styles.eyebrowStyle.Render("NETWORK DETAILS · IPv6"), "", grid))
}

func formatDecimal(raw string) string {
	for i := len(raw) - 3; i > 0; i -= 3 {
		raw = raw[:i] + "," + raw[i:]
	}
	return raw
}
