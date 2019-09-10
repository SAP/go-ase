package tds

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	libraryName    = "go-ase"
	libraryVersion = "0.1.0.0"
)

type TDSVersion struct {
	major, minor, revision, patch uint8
}

func NewTDSVersion(bs []byte) (*TDSVersion, error) {
	if len(bs) != 4 {
		return nil, fmt.Errorf("expected 4 byte array, received %d byte array: %v", len(bs), bs)
	}

	v := &TDSVersion{}
	v.major = uint8(bs[0])
	v.minor = uint8(bs[1])
	v.revision = uint8(bs[2])
	v.patch = uint8(bs[3])

	return v, nil
}

func NewTDSVersionString(s string) (*TDSVersion, error) {
	split := strings.Split(s, ".")
	if len(split) != 4 {
		return nil, fmt.Errorf("expected 4 parts, received %d part string: %v", len(split), s)
	}

	v := &TDSVersion{}

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

	revision, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, fmt.Errorf("error converting revision to integer: %w", err)
	}
	if revision > math.MaxUint8 {
		return nil, fmt.Errorf("revision %d is too large for uint8 (max %d)",
			revision, math.MaxUint8)
	}
	v.revision = uint8(revision)

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

func (tdsv TDSVersion) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", tdsv.major, tdsv.minor, tdsv.revision, tdsv.patch)
}

func (tdsv TDSVersion) Bytes() []byte {
	return []byte{byte(tdsv.major), byte(tdsv.minor), byte(tdsv.revision), byte(tdsv.patch)}
}
