package cli

import (
	"fmt"
	"math/big"
	"net/netip"
	"strconv"
	"strings"

	"github.com/ppp16bit/subnetting/internal/network"
)

func calculateOutput(o options) (string, error) {
	n, err := network.ParseNetwork(o.network)
	if err != nil {
		return "", err
	}
	address := func(a netip.Addr) string {
		if o.expanded {
			return a.StringExpanded()
		}
		return a.String()
	}
	prefix := func(p netip.Prefix) string { return fmt.Sprintf("%s/%d", address(p.Addr()), p.Bits()) }
	var out strings.Builder
	row := func(label, value string) { fmt.Fprintf(&out, "%-20s %s\n", label+":", value) }
	family := "IPv6"
	if n.IPv4 != nil {
		family = "IPv4"
	}
	row("Address family", family)
	row("Provided IP", address(n.Input))
	row("Prefix", fmt.Sprintf("/%d", n.Prefix.Bits()))
	row("Network", prefix(n.Prefix))
	if v4 := n.IPv4; v4 != nil {
		row("Subnet mask", v4.Mask)
		row("Broadcast", v4.Broadcast)
		row("First usable", v4.FirstUsable)
		row("Last usable", v4.LastUsable)
		row("Usable hosts", strconv.FormatUint(v4.UsableHosts, 10))
	} else {
		row("First address", address(n.First))
		row("Last address", address(n.Last))
		row("Total addresses", n.AddressCount().String())
	}
	if o.split != "" {
		bits, err := strconv.Atoi(strings.TrimPrefix(o.split, "/"))
		if err != nil {
			return "", fmt.Errorf("invalid child prefix %q", o.split)
		}
		split, err := network.NewSplit(n.Prefix, bits)
		if err != nil {
			return "", err
		}
		page, err := split.Page(o.offset, o.limit)
		if err != nil {
			return "", err
		}
		out.WriteByte('\n')
		row("Total subnets", split.Count().String())
		row("Addresses per subnet", split.AddressesPerSubnet().String())
		index := new(big.Int).Set(o.offset)
		for _, child := range page {
			fmt.Fprintf(&out, "%s  %s\n", index, prefix(child))
			index.Add(index, big.NewInt(1))
		}
		if index.Cmp(split.Count()) < 0 {
			fmt.Fprintf(&out, "More subnets available; continue with --offset %s\n", index)
		}
	}
	return out.String(), nil
}
