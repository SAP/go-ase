package types

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const (
	ASEDecimalDefaultPrecision = 18
	ASEDecimalDefaultScale     = 0

	ASEMoneyPrecision = 20
	ASEMoneyScale     = 4

	ASESmallMoneyPrecision = 10
	ASESmallMoneyScale     = 4

	aseMaxDecimalDigits = 38
)

var (
	ErrDecimalPrecisionTooHigh         = fmt.Errorf("precision is set to more than %d digits", aseMaxDecimalDigits)
	ErrDecimalPrecisionTooLow          = fmt.Errorf("precision is set to less than 0 digits")
	ErrDecimalScaleTooHigh             = fmt.Errorf("scale is set to more than %d digits", aseMaxDecimalDigits)
	ErrDecimalScaleBiggerThanPrecision = fmt.Errorf("scale is bigger then precision")
)

// Number of bytes required to store the integer representation of
// a decimal.
// Source: https://github.com/thda/tds/blob/master/num.go:16
var numBytes = []int{
	1,
	2, 2, 3, 3, 4, 4, 4,
	5, 5, 6, 6, 6,
	7, 7, 8, 8, 9, 9, 9,
	10, 10, 11, 11, 11,
	12, 12, 13, 13, 14, 14, 14,
	15, 15, 16, 16, 16,
	17, 17, 18, 18, 19, 19, 19,
	20, 20, 21, 21, 21,
	22, 22, 23, 23, 24, 24, 24,
	25, 25, 26, 26, 26,
	27, 27, 28, 28, 28,
	29, 29, 29, 29, 30, 30, 30,
	31, 31, 32, 32, 32,
}

// Returns the number of bytes required to store the integer
// representation of a decimal.
// Returns -1 when the passed length is invalid.
// The passed length is the precision of the decimal.
func DecimalByteSize(length int) int {
	if length < 0 || length > len(numBytes) {
		return -1
	}
	return numBytes[length]
}

// Decimal only carries the information of Decimal, Numeric and Money
// ASE datatypes. This is only sufficient for displaying, not
// calculations.
type Decimal struct {
	precision, scale int
	i                *big.Int
}

// NewDecimal creates a new decimal with the passed precision and scale
// and returns it.
// An error is returned if the precision/scale combination is not valid.
func NewDecimal(precision, scale int) (*Decimal, error) {
	dec := &Decimal{
		precision: precision,
		scale:     scale,
	}

	if err := dec.sanity(); err != nil {
		return nil, err
	}

	return dec, nil
}

// NewDecimalString creates a new decimal based on the passed string.
// If the string contains an invalid precision/scale combination an
// error is returned.
func NewDecimalString(s string) (*Decimal, error) {
	s = strings.TrimSpace(s)

	scale := ASEDecimalDefaultScale
	if strings.Contains(s, ".") {
		split := strings.SplitN(s, ".", 2)
		scale = len(split[1])
	}

	neg := false
	if s[0] == '-' {
		neg = true
	}

	s = strings.TrimLeft(s, "+-")

	split := strings.Split(s, ".")
	prec := len(split[0]) + scale

	dec, err := NewDecimal(prec, scale)
	if err != nil {
		return nil, err
	}

	s = split[0]
	if len(split) == 2 && len(split[1]) > 0 {
		s += split[1]
	}

	i := &big.Int{}
	_, ok := i.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("Not a valid integer: %s", s)
	}

	dec.i = i
	if neg {
		dec.Negate()
	}

	return dec, nil
}

func (dec Decimal) sanity() error {
	if dec.precision > aseMaxDecimalDigits {
		return ErrDecimalPrecisionTooHigh
	}

	if dec.precision < 0 {
		return ErrDecimalPrecisionTooLow
	}

	if dec.scale > aseMaxDecimalDigits {
		return ErrDecimalScaleTooHigh
	}

	if dec.scale > dec.precision {
		return ErrDecimalScaleBiggerThanPrecision
	}

	return nil
}

func (dec Decimal) Cmp(other Decimal) bool {
	if dec.precision != other.precision {
		return false
	}

	if dec.scale != other.scale {
		return false
	}

	return dec.i.Cmp(other.i) == 0
}

func (dec Decimal) Precision() int {
	return dec.precision
}

func (dec Decimal) Scale() int {
	return dec.scale
}

func (dec Decimal) IsNegative() bool {
	return dec.i.Sign() < 0
}

func (dec *Decimal) Negate() {
	dec.i.Neg(dec.i)
}

func (dec Decimal) Bytes() []byte {
	return dec.i.Bytes()
}

func (dec Decimal) ByteSize() int {
	return DecimalByteSize(dec.precision)
}

func (dec *Decimal) SetInt64(i int64) {
	if dec.i == nil {
		dec.i = &big.Int{}
	}

	dec.i.SetInt64(i)
}

// Int returns a copy of the underlying big.Int.
func (dec Decimal) Int() *big.Int {
	return dec.i
}

func (dec *Decimal) SetBytes(b []byte) {
	if len(b) == 0 {
		if dec.i != nil {
			dec.i.SetInt64(0)
		}
		return
	}

	if dec.i == nil {
		dec.i = &big.Int{}
	}

	dec.i.SetBytes(b)
}

func (dec *Decimal) String() string {
	s := fmt.Sprintf("%0"+strconv.Itoa(dec.precision)+"s", big.NewInt(0).Abs(dec.i))

	neg := ""
	if dec.IsNegative() {
		neg = "-"
	}

	right := strings.TrimRight(s[dec.precision-dec.scale:], "0")
	if len(right) == 0 {
		right = "0"
	}

	left := strings.TrimLeft(s[:dec.precision-dec.scale], "0")
	if len(left) == 0 {
		left = "0"
	}

	ret := fmt.Sprintf("%s%s.%s", neg, left, right)

	return ret
}
