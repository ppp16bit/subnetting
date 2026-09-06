package network

import (
	"errors"
	"fmt"
	"math/big"
	"net/netip"
	"strconv"
	"strings"
)

type NetworkInfo struct {
	Input  netip.Addr
	Prefix netip.Prefix
	First  netip.Addr
	Last   netip.Addr
	IPv4   *SubnetInfo
}

func (n *NetworkInfo) AddressCount() *big.Int {
	if n == nil || !n.Prefix.IsValid() {
		return new(big.Int)
	}
	return powerOfTwo(n.Prefix.Addr().BitLen() - n.Prefix.Bits())
}

func powerOfTwo(bits int) *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), uint(bits))
}

func ParseNetwork(input string) (*NetworkInfo, error) {
	parts := strings.Split(strings.TrimSpace(input), "/")
	if len(parts) != 2 {
		return nil, errors.New("expected format: IP/CIDR")
	}
	addr, err := netip.ParseAddr(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid IP address %s", parts[0])
	}
	if addr.Zone() != "" {
		return nil, errors.New("IPv6 zones are not supported in network prefixes")
	}
	bits, err := strconv.Atoi(parts[1])
	if addr.Is4In6() && err == nil && bits >= 0 && bits <= 32 {
		addr = addr.Unmap()
	}
	if err != nil || bits < 0 || bits > addr.BitLen() {
		return nil, fmt.Errorf("CIDR must be an int between 0 and %d", addr.BitLen())
	}
	return CalculateNetwork(netip.PrefixFrom(addr, bits))
}

func CalculateNetwork(prefix netip.Prefix) (*NetworkInfo, error) {
	if !prefix.IsValid() {
		return nil, errors.New("invalid network prefix")
	}
	input := prefix.Addr()
	network := prefix.Masked()
	lastNumber := addrNumber(network.Addr())
	lastNumber.Add(lastNumber, new(big.Int).Sub(powerOfTwo(input.BitLen()-prefix.Bits()), big.NewInt(1)))
	last := numberAddr(lastNumber, input.BitLen())
	info := &NetworkInfo{Input: input, Prefix: network, First: network.Addr(), Last: last}
	if input.Is4() {
		mask := [4]byte{}
		for bit := 0; bit < prefix.Bits(); bit++ {
			mask[bit/8] |= 1 << uint(7-bit%8)
		}
		v4 := &SubnetInfo{IP: input.String(), CIDR: prefix.Bits(), Mask: netip.AddrFrom4(mask).String(), Network: info.First.String(), Broadcast: last.String(), FirstUsable: "N/A", LastUsable: "N/A"}
		if prefix.Bits() <= 30 {
			v4.UsableHosts = info.AddressCount().Uint64() - 2
			v4.FirstUsable, v4.LastUsable = info.First.Next().String(), last.Prev().String()
		}
		info.IPv4 = v4
	}
	return info, nil
}

func addrNumber(addr netip.Addr) *big.Int { 
	return new(big.Int).SetBytes(addr.AsSlice()) 
}

func numberAddr(value *big.Int, bits int) netip.Addr {
	if bits == 32 {
		var bytes [4]byte
		value.FillBytes(bytes[:])
		return netip.AddrFrom4(bytes)
	}
	var bytes [16]byte
	value.FillBytes(bytes[:])
	return netip.AddrFrom16(bytes)
}

func BinaryGroups(addr netip.Addr) []string {
	groups := make([]string, 0, addr.BitLen()/8)
	for _, b := range addr.AsSlice() {
		groups = append(groups, fmt.Sprintf("%08b", b))
	}
	return groups
}
