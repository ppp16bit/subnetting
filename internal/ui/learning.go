package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ppp16bit/subnetting/internal/network"
)

func renderIPv4Learning(styles styleSet, r *network.SubnetInfo, width int, compact bool) string {
	info := network.Explain(r)
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

func compactLearningSteps(styles styleSet, r *network.SubnetInfo, info network.LearningInfo, block string, width int) []string {
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

func renderLearning(styles styleSet, n *network.NetworkInfo, width int, compact, expanded bool) string {
	if n.IPv4 != nil {
		return renderIPv4Learning(styles, n.IPv4, width, compact)
	}
	bits := n.Prefix.Bits()
	rows := []string{
		styles.eyebrowStyle.Render("LEARNING · IPv6"),
		learningStep(styles, 1, "Prefix", fmt.Sprintf("%d network bits; %d remaining address bits", bits, 128-bits)),
		learningStep(styles, 2, "Masking", "Keep the prefix bits; set every remaining bit to zero."),
		learningStep(styles, 3, "Network", n.Prefix.String()),
		learningStep(styles, 4, "Address count", fmt.Sprintf("2^%d = %s", 128-bits, formatDecimal(n.AddressCount().String()))),
		styles.hintStyle.Render("The range includes both endpoints. IPv6 has no broadcast. Total addresses is not a count of assignable hosts."),
	}
	if expanded {
		rows = append(rows, learningStep(styles, 5, "Expanded address", n.Input.StringExpanded()),
			learningStep(styles, 6, "Expanded network", fmt.Sprintf("%s/%d", n.First.StringExpanded(), bits)))
	}
	return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, rows...))
}
