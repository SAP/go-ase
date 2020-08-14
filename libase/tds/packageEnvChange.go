package tds

import "fmt"

//go:generate stringer -type=EnvChangeType
type EnvChangeType uint8

const (
	TDS_ENV_DB       EnvChangeType = 0x1
	TDS_ENV_LANG     EnvChangeType = 0x2
	TDS_ENV_CHARSET  EnvChangeType = 0x3
	TDS_ENV_PACKSIZE EnvChangeType = 0x4
)

type EnvChangePackage struct {
	members []EnvChangePackageField
}

func (pkg *EnvChangePackage) ReadFrom(ch BytesChannel) error {
	length, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}

	var n uint16 = 0
	for n < length {
		member := EnvChangePackageField{}
		i, err := member.ReadFrom(ch)
		if err != nil {
			return fmt.Errorf("error reading EnvChangePackage member: %w", err)
		}
		n += uint16(i)

		pkg.members = append(pkg.members, member)
	}

	if n > length {
		return fmt.Errorf("read too many bytes, %d instead of expected %d", n, length)
	}

	return nil
}

func (pkg EnvChangePackage) WriteTo(ch BytesChannel) error {
	err := ch.WriteUint8(byte(TDS_ENVCHANGE))
	if err != nil {
		return fmt.Errorf("error writing TDS token %s: %w", TDS_ENVCHANGE, err)
	}

	totalLength := 0
	for _, member := range pkg.members {
		totalLength += member.ByteLength()
	}

	if err := ch.WriteUint16(uint16(totalLength)); err != nil {
		return fmt.Errorf("error writing length: %w", err)
	}

	length := 0
	for _, member := range pkg.members {
		n, err := member.WriteTo(ch)
		if err != nil {
			return fmt.Errorf("error writing EnvChangePackage member: %w", err)
		}
		length += n
	}

	if length != totalLength {
		return fmt.Errorf("wrote %d bytes instead of expected %d bytes", length, totalLength)
	}

	return nil
}

func (pkg EnvChangePackage) String() string {
	s := fmt.Sprintf("%T(", pkg)

	for _, member := range pkg.members {
		s += fmt.Sprintf("%s(%s -> %s)", member.Type, member.OldValue, member.NewValue)
	}

	return s + ")"
}

type EnvChangePackageField struct {
	Type               EnvChangeType
	NewValue, OldValue string
}

func (field *EnvChangePackageField) ReadFrom(ch BytesChannel) (int, error) {
	// n is the amount of bytes read from channel
	n := 0

	typ, err := ch.Uint8()
	if err != nil {
		return n, ErrNotEnoughBytes
	}
	field.Type = EnvChangeType(typ)
	n++

	length, err := ch.Uint8()
	if err != nil {
		return n, ErrNotEnoughBytes
	}
	n++

	if length > 0 {
		field.NewValue, err = ch.String(int(length))
		if err != nil {
			return n, ErrNotEnoughBytes
		}
		n += int(length)
	}

	length, err = ch.Uint8()
	if err != nil {
		return n, ErrNotEnoughBytes
	}
	n++

	if length > 0 {
		field.OldValue, err = ch.String(int(length))
		if err != nil {
			return n, ErrNotEnoughBytes
		}
		n += int(length)
	}

	return n, nil
}

func (field EnvChangePackageField) WriteTo(ch BytesChannel) (int, error) {
	err := ch.WriteUint8(uint8(field.Type))
	if err != nil {
		return 0, fmt.Errorf("error writing type: %w", err)
	}
	n := 1

	err = ch.WriteUint8(uint8(len(field.NewValue)))
	if err != nil {
		return n, fmt.Errorf("error writing new value length: %w", err)
	}
	n++

	err = ch.WriteString(field.NewValue)
	if err != nil {
		return n, fmt.Errorf("error writing new value: %w", err)
	}
	n += len(field.NewValue)

	err = ch.WriteUint8(uint8(len(field.OldValue)))
	if err != nil {
		return n, fmt.Errorf("error writing old value length: %w", err)
	}
	n++

	err = ch.WriteString(field.OldValue)
	if err != nil {
		return n, fmt.Errorf("error writing old value: %w", err)
	}
	n += len(field.OldValue)

	return n, nil
}

func (field EnvChangePackageField) ByteLength() int {
	// type byte
	// + new value length byte + new value length
	// + old value length byte + old value length
	return 3 + len(field.NewValue) + len(field.OldValue)
}
