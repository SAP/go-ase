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
	Precision, Scale int
	i                *big.Int
}

// NewDecimal creates a new decimal with the passed precision and scale
// and returns it.
// An error is returned if the precision/scale combination is not valid.
func NewDecimal(precision, scale int) (*Decimal, error) {
	dec := &Decimal{
		Precision: precision,
		Scale:     scale,
	}

	if err := dec.sanity(); err != nil {
		return nil, err
	}

	return dec, nil
}

// NewDecimalString creates a new decimal based on the passed string.
// If the string contains an invalid precision/scale combination an
// error is returned.
func NewDecimalString(precision, scale int, s string) (*Decimal, error) {
	dec, err := NewDecimal(precision, scale)
	if err != nil {
		return nil, fmt.Errorf("error creating decimal: %w", err)
	}

	err = dec.SetString(s)
	if err != nil {
		return nil, fmt.Errorf("error setting string: %w", err)
	}

	return dec, nil
}

func (dec Decimal) sanity() error {
	if dec.Precision > aseMaxDecimalDigits {
		return ErrDecimalPrecisionTooHigh
	}

	if dec.Precision < 0 {
		return ErrDecimalPrecisionTooLow
	}

	if dec.Scale > aseMaxDecimalDigits {
		return ErrDecimalScaleTooHigh
	}

	if dec.Scale > dec.Precision {
		return ErrDecimalScaleBiggerThanPrecision
	}

	return nil
}

func (dec Decimal) Cmp(other Decimal) bool {
	if dec.Precision != other.Precision {
		return false
	}

	if dec.Scale != other.Scale {
		return false
	}

	return dec.i.Cmp(other.i) == 0
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
	return DecimalByteSize(dec.Precision)
}

func (dec *Decimal) SetInt64(i int64) {
	if dec.i == nil {
		dec.i = &big.Int{}
	}

	dec.i.SetInt64(i)
}

// Int returns a copy of the underlying big.Int.
func (dec Decimal) Int() *big.Int {
	i := &big.Int{}
	i.Add(i, dec.i)
	return i
}

func (dec *Decimal) SetBytes(b []byte) {
	if dec.i == nil {
		dec.i = &big.Int{}
	}

	dec.i.SetBytes(b)
}

func (dec *Decimal) String() string {
	s := fmt.Sprintf("%0"+strconv.Itoa(dec.Precision)+"s", big.NewInt(0).Abs(dec.i))

	neg := ""
	if dec.IsNegative() {
		neg = "-"
	}

	right := strings.TrimRight(s[dec.Precision-dec.Scale:], "0")
	if len(right) == 0 {
		right = "0"
	}

	left := strings.TrimLeft(s[:dec.Precision-dec.Scale], "0")
	if len(left) == 0 {
		left = "0"
	}

	ret := fmt.Sprintf("%s%s.%s", neg, left, right)

	return ret
}

// Set decimal to the passed string value.
// Precision and scale are untouched.
//
// If an error is returned dec is untouched.
func (dec *Decimal) SetString(s string) error {
	// Trim spaces to avoid errors with "+0.0 " etc.pp.
	s = strings.TrimSpace(s)

	split := strings.Split(s, ".")
	left := split[0]
	right := ""
	if len(split) > 1 {
		right = split[1]
	}

	// Set underlying big.Int structure to the whole number
	i := &big.Int{}
	if _, ok := i.SetString(left+right, 10); !ok {
		return fmt.Errorf("failed to parse number %s%s", left, right)
	}

	// Multiply underlying big.Int to fit to the scale of the decimal
	if dec.Scale-len(right) > 0 {
		mul := big.NewInt(10)
		mul.Exp(mul, big.NewInt(int64(dec.Scale-len(right))), nil)
		i.Mul(i, mul)
	}

	dec.i = i
	return nil
}
