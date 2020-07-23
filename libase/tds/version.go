package tds

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	libraryName = "go-ase/tds"
)

var (
	libraryVersion = Version{
		major: 0x0,
		minor: 0x1,
		sp:    0x0,
		patch: 0x0,
	}
)

type Version struct {
	major, minor, sp, patch uint8
}

func NewVersion(bs []byte) (*Version, error) {
	if len(bs) != 4 {
		return nil, fmt.Errorf("expected 4 byte array, received %d byte array: %v", len(bs), bs)
	}

	v := &Version{}
	v.major = uint8(bs[0])
	v.minor = uint8(bs[1])
	v.sp = uint8(bs[2])
	v.patch = uint8(bs[3])

	return v, nil
}

func NewVersionString(s string) (*Version, error) {
	split := strings.Split(s, ".")
	if len(split) != 4 {
		return nil, fmt.Errorf("expected 4 parts, received %d part string: %v", len(split), s)
	}

	v := &Version{}

	major, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, fmt.Errorf("error converting major to integer: %w", err)
	}
	if major > math.MaxUint8 {
		return nil, fmt.Errorf("major %d is too large for uint8 (max %d)",
			major, math.MaxUint8)
	}
	v.major = uint8(major)

	minor, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, fmt.Errorf("error converting minor to integer: %w", err)
	}
	if minor > math.MaxUint8 {
		return nil, fmt.Errorf("minor %d is too large for uint8 (max %d)",
			minor, math.MaxUint8)
	}
	v.minor = uint8(minor)

	sp, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, fmt.Errorf("error converting revision to integer: %w", err)
	}
	if sp > math.MaxUint8 {
		return nil, fmt.Errorf("revision %d is too large for uint8 (max %d)",
			sp, math.MaxUint8)
	}
	v.sp = uint8(sp)

	patch, err := strconv.Atoi(split[3])
	if err != nil {
		return nil, fmt.Errorf("error converting patch to integer: %w", err)
	}
	if patch > math.MaxUint8 {
		return nil, fmt.Errorf("patch %d is too large for uint8 (max %d)",
			patch, math.MaxUint8)
	}
	v.patch = uint8(patch)

	return v, nil
}

func (tdsv Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", tdsv.major, tdsv.minor, tdsv.sp, tdsv.patch)
}

func (tdsv Version) Bytes() []byte {
	return []byte{byte(tdsv.major), byte(tdsv.minor), byte(tdsv.sp), byte(tdsv.patch)}
}
