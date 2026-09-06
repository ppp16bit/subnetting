package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ppp16bit/subnetting/internal/network"
)

func renderIPv4Binary(styles styleSet, r *network.SubnetInfo, width int) string {
	innerWidth := max(8, width-styles.resultPanelStyle.GetHorizontalFrameSize())
	if innerWidth < 35 {
		rows := []string{
			styles.eyebrowStyle.Render("BINARY AND"),
			compactBinaryRow(styles, "IP", network.BinaryIPv4(r.IP), r.CIDR, innerWidth),
			compactBinaryRow(styles, "MASK", network.BinaryIPv4(r.Mask), r.CIDR, innerWidth),
			compactBinaryRow(styles, "AND", network.BinaryIPv4(r.Network), r.CIDR, innerWidth),
		}
		return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, rows...))
	}

	rows := []string{
		styles.eyebrowStyle.Render("BINARY AND"),
		"",
		binaryRow(styles, "IP", network.BinaryIPv4(r.IP), r.CIDR, innerWidth),
		binaryRow(styles, "MASK", network.BinaryIPv4(r.Mask), r.CIDR, innerWidth),
		styles.separatorStyle.Render(strings.Repeat("─", innerWidth)),
		binaryRow(styles, "NETWORK", network.BinaryIPv4(r.Network), r.CIDR, innerWidth),
		"",
		styles.hintStyle.Render("network bits + host bits"),
	}
	return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func renderIPv4CombinedLearningBinary(styles styleSet, r *network.SubnetInfo, width int) string {
	info := network.Explain(r)
	innerWidth := max(8, width-styles.resultPanelStyle.GetHorizontalFrameSize())
	block := fmt.Sprintf("%d–%d", info.BlockStart, info.BlockEnd)
	rows := []string{styles.eyebrowStyle.Render("LEARNING + BINARY")}
	rows = append(rows, compactLearningSteps(styles, r, info, block, innerWidth)...)
	rows = append(rows,
		styles.eyebrowStyle.Render("BINARY AND"),
		compactBinaryRow(styles, "IP", network.BinaryIPv4(r.IP), r.CIDR, innerWidth),
		compactBinaryRow(styles, "MASK", network.BinaryIPv4(r.Mask), r.CIDR, innerWidth),
		compactBinaryRow(styles, "AND", network.BinaryIPv4(r.Network), r.CIDR, innerWidth),
	)
	return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func compactBinaryRow(styles styleSet, label, bits string, prefix, width int) string {
	lines := strings.Split(formatBinary(styles, bits, prefix, max(1, width-6)), "\n")
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

func renderBinary(styles styleSet, n *network.NetworkInfo, width int) string {
	if n.IPv4 != nil {
		return renderIPv4Binary(styles, n.IPv4, width)
	}
	inner := max(1, width-styles.resultPanelStyle.GetHorizontalFrameSize())
	perLine := 32
	if inner < 38 {
		perLine = 16
	}
	if inner < 22 {
		perLine = 8
	}
	if inner < 14 {
		perLine = max(1, inner-5)
	}
	bits := n.Prefix.Bits()
	ip := strings.Join(network.BinaryGroups(n.Input), "")
	network := strings.Join(network.BinaryGroups(n.First), "")
	mask := strings.Repeat("1", bits) + strings.Repeat("0", 128-bits)
	rows := []string{styles.eyebrowStyle.Render("BINARY AND · IPv6")}
	for start := 0; start < 128; start += perLine {
		end := min(128, start+perLine)
		prefix := max(0, min(end-start, bits-start))
		rows = append(rows, styles.statLabelStyle.Render(fmt.Sprintf("Bits %d–%d", start+1, end)),
			"IP   "+colorBinary(styles, ip[start:end], prefix),
			"MASK "+colorBinary(styles, mask[start:end], prefix),
			"AND  "+colorBinary(styles, network[start:end], prefix))
	}
	rows = append(rows, styles.hintStyle.Render(fmt.Sprintf("%d network bits + %d remaining address bits", bits, 128-bits)))
	return renderFrame(styles.resultPanelStyle, width, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func renderCombinedLearningBinary(styles styleSet, n *network.NetworkInfo, width int, expanded bool) string {
	if n.IPv4 != nil {
		return renderIPv4CombinedLearningBinary(styles, n.IPv4, width)
	}
	return lipgloss.JoinVertical(lipgloss.Left, renderLearning(styles, n, width, true, expanded), renderBinary(styles, n, width))
}
