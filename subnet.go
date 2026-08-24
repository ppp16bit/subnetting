package subnetting

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
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

// LearningInfo contains the intermediate values used to teach the subnet
// calculation. It is derived from SubnetInfo so the explanation can never
// drift away from the calculator result.
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
	ipU32 := binary.BigEndian.Uint32(ip)
	maskU32 := ^uint32(0) << (32 - cidr)
	networkU32 := ipU32 & maskU32
	broadcastU32 := ipU32 | ^maskU32
	hostBits := 32 - cidr

	var usable uint64
	var firstUsabelStr, lastUsableStr string

	if hostBits >= 2 {
		usable = (uint64(1) << uint(hostBits)) - 2
		firstIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(firstIP, networkU32+1)
		firstUsabelStr = firstIP.String()

		lastIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(lastIP, broadcastU32-1)
		lastUsableStr = lastIP.String()
	} else {
		usable = 0
		firstUsabelStr = "N/A"
		lastUsableStr = "N/A"
	}

	maskIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(maskIP, maskU32)

	networkIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(networkIP, networkU32)

	broadcastIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(broadcastIP, broadcastU32)

	return &SubnetInfo{
		IP:          ip.String(),
		CIDR:        cidr,
		Mask:        maskIP.String(),
		Network:     networkIP.String(),
		Broadcast:   broadcastIP.String(),
		FirstUsable: firstUsabelStr,
		LastUsable:  lastUsableStr,
		UsableHosts: usable,
	}
}

func ParseAndCalculate(input string) (*SubnetInfo, error) {
	ip, cidr, err := ParseInput(input)
	if err != nil {
		return nil, err
	}
	return Calculate(ip, cidr), nil
}

// Explain returns the decimal building blocks behind the subnet result. For
// octet-boundary prefixes (for example /24), the following zero mask octet is
// used because it makes the 256 - mask shortcut and the block explicit.
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

// BinaryIPv4 renders an IPv4 address as four zero-padded binary octets.
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
