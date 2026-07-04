package main

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
		return nil, 0, errors.New("only IPV4 address are suported")
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
