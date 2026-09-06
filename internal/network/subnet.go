package network

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

type SubnetInfo struct {
	IP          string
	CIDR        int
	Mask        string
	Network     string
	Broadcast   string
	FirstUsable string
	LastUsable  string
	UsableHosts uint64
}

type LearningInfo struct {
	InterestingOctet int
	IPOctet          int
	MaskOctet        int
	Increment        int
	SubnetStarts     []int
	BlockStart       int
	BlockEnd         int
}

func ParseInput(input string) (net.IP, int, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, 0, errors.New("expected format: IP/CIDR")
	}

	ip := net.ParseIP(parts[0])
	if ip == nil {
		return nil, 0, fmt.Errorf("invalid IP address %s", parts[0])
	}

	ip = ip.To4()
	if ip == nil {
		return nil, 0, errors.New("only IPv4 addresses are supported")
	}

	cidr, err := strconv.Atoi(parts[1])
	if err != nil || cidr < 0 || cidr > 32 {
		return nil, 0, errors.New("CIDR must be an int between 0 and 32")
	}
	return ip, cidr, nil
}

func Calculate(ip net.IP, cidr int) *SubnetInfo {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok || !addr.Unmap().Is4() {
		return nil
	}
	info, err := CalculateNetwork(netip.PrefixFrom(addr.Unmap(), cidr))
	if err != nil {
		return nil
	}
	return info.IPv4
}

func ParseAndCalculate(input string) (*SubnetInfo, error) {
	ip, cidr, err := ParseInput(input)
	if err != nil {
		return nil, err
	}
	return Calculate(ip, cidr), nil
}

func Explain(info *SubnetInfo) LearningInfo {
	ip := net.ParseIP(info.IP).To4()
	mask := net.ParseIP(info.Mask).To4()

	octet := info.CIDR / 8
	if octet > 3 {
		octet = 3
	}

	ipOctet := int(ip[octet])
	maskOctet := int(mask[octet])
	increment := 256 - maskOctet
	blockStart := (ipOctet / increment) * increment
	blockEnd := min(255, blockStart+increment-1)

	starts := make([]int, 0, 256/increment)
	for start := 0; start < 256; start += increment {
		starts = append(starts, start)
	}

	return LearningInfo{
		InterestingOctet: octet + 1,
		IPOctet:          ipOctet,
		MaskOctet:        maskOctet,
		Increment:        increment,
		SubnetStarts:     starts,
		BlockStart:       blockStart,
		BlockEnd:         blockEnd,
	}
}

func BinaryIPv4(address string) string {
	ip := net.ParseIP(address).To4()
	if ip == nil {
		return ""
	}

	parts := make([]string, len(ip))
	for i, octet := range ip {
		parts[i] = fmt.Sprintf("%08b", octet)
	}
	return strings.Join(parts, ".")
}
