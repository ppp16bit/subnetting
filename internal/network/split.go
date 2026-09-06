package network

import (
	"errors"
	"fmt"
	"math/big"
	"net/netip"
)

const MaxPageSize = 1000

type Split struct {
	parent netip.Prefix
	bits   int
}

func NewSplit(parent netip.Prefix, bits int) (*Split, error) {
	if !parent.IsValid() {
		return nil, errors.New("invalid parent prefix")
	}
	if bits < parent.Bits() || bits > parent.Addr().BitLen() {
		return nil, fmt.Errorf("child prefix must be between /%d and /%d", parent.Bits(), parent.Addr().BitLen())
	}
	return &Split{parent: parent.Masked(), bits: bits}, nil
}

func (s *Split) Count() *big.Int              { return powerOfTwo(s.bits - s.parent.Bits()) }
func (s *Split) AddressesPerSubnet() *big.Int { return powerOfTwo(s.parent.Addr().BitLen() - s.bits) }

func (s *Split) Child(index *big.Int) (netip.Prefix, error) {
	if index == nil || index.Sign() < 0 || index.Cmp(s.Count()) >= 0 {
		return netip.Prefix{}, errors.New("subnet index is out of range")
	}
	offset := new(big.Int).Lsh(new(big.Int).Set(index), uint(s.parent.Addr().BitLen()-s.bits))
	offset.Add(offset, addrNumber(s.parent.Addr()))
	return netip.PrefixFrom(numberAddr(offset, s.parent.Addr().BitLen()), s.bits), nil
}

func (s *Split) Page(offset *big.Int, limit int) ([]netip.Prefix, error) {
	if limit < 1 || limit > MaxPageSize {
		return nil, fmt.Errorf("limit must be between 1 and %d", MaxPageSize)
	}
	if offset == nil || offset.Sign() < 0 || offset.Cmp(s.Count()) >= 0 {
		return nil, errors.New("subnet offset is out of range")
	}
	index, count := new(big.Int).Set(offset), s.Count()
	page := make([]netip.Prefix, 0, limit)
	for len(page) < limit && index.Cmp(count) < 0 {
		child, err := s.Child(index)
		if err != nil {
			return nil, err
		}
		page = append(page, child)
		index.Add(index, big.NewInt(1))
	}
	return page, nil
}
