package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
	"unicode/utf16"

	"github.com/SAP/go-ase/libase/asetime"
)

func (t DataType) Bytes(endian binary.ByteOrder, value interface{}) ([]byte, error) {
	switch t {
	case MONEY, SHORTMONEY:
		bs := make([]byte, t.ByteSize())
		dec, ok := value.(*Decimal)
		if !ok {
			return nil, fmt.Errorf("expected *types.Decimal for %s, received %T", t, value)
		}
		deci := dec.Int()

		if t == MONEY {
			endian.PutUint32(bs[:4], uint32(deci.Int64()>>32))
			endian.PutUint32(bs[4:], uint32(deci.Int64()))
		} else {
			endian.PutUint32(bs, uint32(deci.Int64()))
		}

		return bs, nil
	case DATE:
		t := asetime.DurationFromDateTime(value.(time.Time))
		t -= asetime.DurationFromDateTime(asetime.Epoch1900())

		bs := make([]byte, 4)
		endian.PutUint32(bs, uint32(t.Days()))
		return bs, nil
	case TIME:
		dur := asetime.DurationFromTime(value.(time.Time))
		fract := asetime.MillisecondToFractionalSecond(dur.Microseconds())

		bs := make([]byte, 4)
		endian.PutUint32(bs, uint32(fract))
		return bs, nil
	case SHORTDATE:
		t := asetime.DurationFromDateTime(value.(time.Time))
		t -= asetime.DurationFromDateTime(asetime.Epoch1900())

		days := t.Days()
		s := asetime.ASEDuration(t.Microseconds() - days*int(asetime.Day))

		bs := make([]byte, 4)
		// TODO replace all binary.Littleendian
		binary.LittleEndian.PutUint16(bs[:2], uint16(days))
		binary.LittleEndian.PutUint16(bs[2:], uint16(s.Minutes()))
		return bs, nil
	case DATETIME:
		t := asetime.DurationFromDateTime(value.(time.Time))
		t -= asetime.DurationFromDateTime(asetime.Epoch1900())

		days := t.Days()
		s := t.Microseconds() - days*int(asetime.Day)
		s = asetime.MillisecondToFractionalSecond(s)

		bs := make([]byte, 8)
		binary.LittleEndian.PutUint32(bs[:4], uint32(days))
		binary.LittleEndian.PutUint32(bs[4:], uint32(s))
		return bs, nil
	case BIGDATETIMEN:
		dur := asetime.DurationFromDateTime(value.(time.Time))

		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(dur))
		return bs, nil
	case BIGTIMEN:
		dur := asetime.DurationFromTime(value.(time.Time))

		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(dur))
		return bs, nil
	case UNITEXT:
		// convert go string to utf16 code points
		runes := []rune(value.(string))
		utf16bytes := utf16.Encode(runes)

		// convert utf16 code points to bytes
		bs := make([]byte, len(utf16bytes)*2)
		for i := 0; i < len(utf16bytes); i++ {
			binary.LittleEndian.PutUint16(bs[i:], utf16bytes[i])
		}

		return bs, nil
	}

	buf := &bytes.Buffer{}
	err := binary.Write(buf, endian, value)
	if err != nil {
		return nil, fmt.Errorf("error writing value: %w", err)
	}
	return buf.Bytes(), nil
}
