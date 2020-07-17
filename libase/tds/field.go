// The types in this file are format fields for TDS_PARAMFMT,
// TDS_PARAMFMT2, TDS_ROWFMT, TDS_ROWFMT2 and data fields for
// TDS_PARAMS and TDS_ROW.
//
// To make handling and identification of the different data types
// easier in Go each data type has their own field structure, created by
// embedding abstract types.
//
// Note: The data types are embedded instead of aliased to provide the
// methods of the embedded types - aliasing does not provide access to
// the methods of the source type.

package tds

import "fmt"

// Both Param- and RowFmtStatus are uints communicated using
// TDS_PARAMFMT* and TDS_ROWFMT*. Depending on the token they have
// slightly different meanings.
//
// fmtStatus is only used for legibility in the code when e.g. checking
// for column status or if null types are allowed. The methods setting
// and getting status convert it from and to fmtStatus.
type fmtStatus uint

const (
	tdsFmtColumnStatus fmtStatus = 0x8
	tdsFmtNullAllowed  fmtStatus = 0x20
)

//go:generate stringer -type=ParamFmtStatus
type ParamFmtStatus uint

const (
	TDS_PARAM_NOSTATUS     ParamFmtStatus = 0x0
	TDS_PARAM_RETURN       ParamFmtStatus = 0x1
	TDS_PARAM_COLUMNSTATUS ParamFmtStatus = 0x8
	TDS_PARAM_NULLALLOWED  ParamFmtStatus = 0x20
)

//go:generate stringer -type=RowFmtStatus
type RowFmtStatus uint

const (
	TDS_ROW_NOSTATUS     RowFmtStatus = 0x0
	TDS_ROW_HIDDEN       RowFmtStatus = 0x1
	TDS_ROW_KEY          RowFmtStatus = 0x2
	TDS_ROW_VERSION      RowFmtStatus = 0x4
	TDS_ROW_COLUMNSTATUS RowFmtStatus = 0x8
	TDS_ROW_UPDATEABLE   RowFmtStatus = 0x10
	TDS_ROW_NULLALLOWED  RowFmtStatus = 0x20
	TDS_ROW_IDENTITY     RowFmtStatus = 0x40
	TDS_ROW_PADCHAR      RowFmtStatus = 0x80
)

type DataStatus uint

const (
	TDS_DATA_NONNULL           DataStatus = 0x0
	TDS_DATA_NULL              DataStatus = 0x1
	TDS_DATA_ZEROLENGTHNONNULL DataStatus = 0x2
	TDS_DATA_RESERVED          DataStatus = 0xfc
)

// Interfaces

type FieldFmt interface {
	// Format information as sent to or received from TDS server
	DataType() DataType
	SetName(string)
	Name() string

	// specific to TDS_ROWFMT2
	SetColumnLabel(string)
	ColumnLabel() string
	SetCatalogue(string)
	Catalogue() string
	SetSchema(string)
	Schema() string
	SetTable(string)
	Table() string

	SetStatus(uint)
	Status() uint

	SetUserType(int32)
	UserType() int32
	SetLocaleInfo(string)
	LocaleInfo() string

	// Interface methods for go-ase

	// Returns true if the data type has a fixed length.
	IsFixedLength() bool
	// The return value of LengthBytes depends on IsFixedLength.
	// If the data type has a fixed length LengthBytes returns the
	// total number of bytes of the data portion (not the entire data
	// field - only the actual data).
	// If the data type has a variable length LengthBytes returns the
	// number of bytes to be read from the data stream for the length in
	// bytes of the data portion.
	LengthBytes() int
	// Length returns the maximum length of the column
	// TODO: is this actually required when sending from client?
	MaxLength() int

	ReadFrom(BytesChannel) (int, error)
	WriteTo(BytesChannel) (int, error)

	FormatByteLength() int
}

type FieldData interface {
	// Format information send by TDS server
	Status() DataStatus

	// Interface methods for go-ase
	SetData([]byte)
	Data() []byte
	ReadFrom(BytesChannel) (int, error)
	WriteTo(BytesChannel) (int, error)
}

// Base structs

type fieldFmtBase struct {
	dataType DataType
	name     string

	// specific to TDS_ROWFMT2
	// wide_row controls if the TDS_ROWFMT2 specific members are filled
	// and written. It is set by TDS_ROWFMT2 when creating a field.
	wide_row    bool
	columnLabel string
	catalogue   string
	schema      string
	table       string

	status     fmtStatus
	userType   int32
	localeInfo string

	// isFixedLength defines if the data field belonging to the fmt may
	// have its own length.
	isFixedLength bool
	// lengthBytes defines the byte size of the length field
	lengthBytes int
	// length is the maximum length of the data type
	maxLength int
}

func (field fieldFmtBase) DataType() DataType {
	return field.dataType
}

func (field *fieldFmtBase) SetName(name string) {
	field.name = name
}

func (field fieldFmtBase) Name() string {
	return field.name
}

func (field *fieldFmtBase) SetColumnLabel(columnLabel string) {
	field.columnLabel = columnLabel
}

func (field fieldFmtBase) ColumnLabel() string {
	return field.columnLabel
}

func (field *fieldFmtBase) SetCatalogue(catalogue string) {
	field.catalogue = catalogue
}

func (field fieldFmtBase) Catalogue() string {
	return field.catalogue
}

func (field *fieldFmtBase) SetSchema(schema string) {
	field.schema = schema
}

func (field fieldFmtBase) Schema() string {
	return field.schema
}

func (field *fieldFmtBase) SetTable(table string) {
	field.table = table
}

func (field fieldFmtBase) Table() string {
	return field.table
}

func (field *fieldFmtBase) SetStatus(status uint) {
	field.status = fmtStatus(status)
}

func (field fieldFmtBase) Status() uint {
	return uint(field.status)
}

func (field *fieldFmtBase) SetUserType(userType int32) {
	field.userType = userType
}

func (field fieldFmtBase) UserType() int32 {
	return field.userType
}

func (field *fieldFmtBase) SetLocaleInfo(localeInfo string) {
	field.localeInfo = localeInfo
}

func (field fieldFmtBase) LocaleInfo() string {
	return field.localeInfo
}

func (field fieldFmtBase) IsFixedLength() bool {
	return field.isFixedLength
}

func (field fieldFmtBase) LengthBytes() int {
	return field.lengthBytes
}

func (field fieldFmtBase) MaxLength() int {
	return field.maxLength
}

func (field *fieldFmtBase) readFromBase(ch BytesChannel) (int, error) {
	if field.isFixedLength {
		return 0, nil
	}

	length, err := readLengthBytes(ch, field.lengthBytes)
	if err != nil {
		return 0, err
	}
	field.maxLength = length

	return field.lengthBytes, nil
}

func (field fieldFmtBase) writeToBase(ch BytesChannel) (int, error) {
	if field.isFixedLength {
		return 0, nil
	}

	return field.lengthBytes, writeLengthBytes(ch, field.lengthBytes, field.maxLength)
}

type fieldFmtBasePrecision struct {
	precision uint8
}

func (field fieldFmtBasePrecision) Precision() uint8 {
	return field.precision
}

func (field *fieldFmtBasePrecision) readFromPrecision(ch BytesChannel) (int, error) {
	var err error
	field.precision, err = ch.Uint8()
	if err != nil {
		return 0, fmt.Errorf("failed to read precision: %w", err)
	}
	return 1, nil
}

func (field fieldFmtBasePrecision) writeToPrecision(ch BytesChannel) (int, error) {
	err := ch.WriteUint8(field.precision)
	if err != nil {
		return 0, fmt.Errorf("failed to write precision: %w", err)
	}
	return 1, nil
}

type fieldFmtBaseScale struct {
	scale uint8
}

func (field fieldFmtBaseScale) Scale() uint8 {
	return field.scale
}

func (field *fieldFmtBaseScale) readFromScale(ch BytesChannel) (int, error) {
	var err error
	field.scale, err = ch.Uint8()
	if err != nil {
		return 0, fmt.Errorf("failed to read scale: %w", err)
	}
	return 1, nil
}

func (field fieldFmtBaseScale) writeToScale(ch BytesChannel) (int, error) {
	err := ch.WriteUint8(field.scale)
	if err != nil {
		return 0, fmt.Errorf("failed to write scale: %w", err)
	}
	return 1, nil
}

type fieldDataBase struct {
	fmt    FieldFmt
	status DataStatus
	data   []byte
}

func (field fieldDataBase) Status() DataStatus {
	return field.status
}

func (field *fieldDataBase) SetData(data []byte) {
	field.data = data
}

func (field fieldDataBase) Data() []byte {
	return field.data
}

func (field fieldDataBase) String() string {
	return string(field.data)
}

func (field *fieldDataBase) readFromStatus(ch BytesChannel) (int, error) {
	if fmtStatus(field.fmt.Status())&tdsFmtColumnStatus != tdsFmtColumnStatus {
		return 0, nil
	}

	status, err := ch.Uint8()
	if err != nil {
		return 0, fmt.Errorf("failed to read status: %w", err)
	}
	field.status = DataStatus(status)
	return 1, nil
}

func (field fieldDataBase) writeToStatus(ch BytesChannel) (int, error) {
	if fmtStatus(field.fmt.Status())&tdsFmtColumnStatus != tdsFmtColumnStatus {
		return 0, nil
	}

	// TODO depends on wide
	err := ch.WriteUint8(uint8(field.status))
	if err != nil {
		return 0, fmt.Errorf("failed to write status: %w", err)
	}
	return 1, nil
}

func (field *fieldDataBase) readFrom(ch BytesChannel) (int, error) {
	n, err := field.readFromStatus(ch)
	if err != nil {
		return n, err
	}

	length := field.fmt.LengthBytes()
	if !field.fmt.IsFixedLength() {
		var err error
		length, err = readLengthBytes(ch, field.fmt.LengthBytes())
		if err != nil {
			return n, fmt.Errorf("failed to read %d bytes of length: %w", field.fmt.LengthBytes(), err)
		}
		n += field.fmt.LengthBytes()
	}

	if field.data, err = ch.Bytes(length); err != nil {
		return n, fmt.Errorf("failed to read %d bytes of data: %w", length, err)
	}
	n += length

	return n, nil
}

func (field fieldDataBase) writeTo(ch BytesChannel) (int, error) {
	n, err := field.writeToStatus(ch)
	if err != nil {
		return n, err
	}

	if !field.fmt.IsFixedLength() {
		if err := writeLengthBytes(ch, field.fmt.LengthBytes(), len(field.data)); err != nil {
			return n, fmt.Errorf("failed to write data length: %w", err)
		}
		n += field.fmt.LengthBytes()
	}

	if err := ch.WriteBytes(field.data); err != nil {
		return n, fmt.Errorf("failed to write %d bytes of data: %w", len(field.data), err)
	}
	n += len(field.data)

	return n, nil
}

// Implementations

type fieldFmtLength struct {
	fieldFmtBase
}

// TODO is this being used?
func (field fieldFmtLength) FormatByteLength() int {
	return field.lengthBytes
}

func (field *fieldFmtLength) ReadFrom(ch BytesChannel) (int, error) {
	return field.readFromBase(ch)
}

func (field fieldFmtLength) WriteTo(ch BytesChannel) (int, error) {
	return field.writeToBase(ch)
}

type BitFieldFmt struct{ fieldFmtLength }
type DateTimeFieldFmt struct{ fieldFmtLength }
type DateFieldFmt struct{ fieldFmtLength }
type ShortDateFieldFmt struct{ fieldFmtLength }
type Flt4FieldFmt struct{ fieldFmtLength }
type Flt8FieldFmt struct{ fieldFmtLength }
type Int1FieldFmt struct{ fieldFmtLength }
type Int2FieldFmt struct{ fieldFmtLength }
type Int4FieldFmt struct{ fieldFmtLength }
type Int8FieldFmt struct{ fieldFmtLength }
type IntervalFieldFmt struct{ fieldFmtLength }
type Sint1FieldFmt struct{ fieldFmtLength }
type Uint2FieldFmt struct{ fieldFmtLength }
type Uint4FieldFmt struct{ fieldFmtLength }
type Uint8FieldFmt struct{ fieldFmtLength }
type MoneyFieldFmt struct{ fieldFmtLength }
type ShortMoneyFieldFmt struct{ fieldFmtLength }
type TimeFieldFmt struct{ fieldFmtLength }
type BinaryFieldFmt struct{ fieldFmtLength }
type BoundaryFieldFmt struct{ fieldFmtLength }
type CharFieldFmt struct{ fieldFmtLength }
type DateNFieldFmt struct{ fieldFmtLength }
type DateTimeNFieldFmt struct{ fieldFmtLength }
type FltNFieldFmt struct{ fieldFmtLength }
type IntNFieldFmt struct{ fieldFmtLength }
type UintNFieldFmt struct{ fieldFmtLength }
type LongBinaryFieldFmt struct{ fieldFmtLength }
type LongCharFieldFmt struct{ fieldFmtLength }
type MoneyNFieldFmt struct{ fieldFmtLength }
type SensitivityFieldFmt struct{ fieldFmtLength }
type TimeNFieldFmt struct{ fieldFmtLength }
type VarBinaryFieldFmt struct{ fieldFmtLength }
type VarCharFieldFmt struct{ fieldFmtLength }

type fieldFmtLengthScale struct {
	fieldFmtBase
	fieldFmtBaseScale
}

func (field fieldFmtLengthScale) FormatByteLength() int {
	// 1 byte scale
	// 1 to 4 bytes length
	return 1 + field.lengthBytes
}

func (field *fieldFmtLengthScale) ReadFrom(ch BytesChannel) (int, error) {
	n, err := field.readFromBase(ch)
	if err != nil {
		return n, err
	}

	n2, err := field.readFromScale(ch)
	return n + n2, err
}

func (field fieldFmtLengthScale) WriteTo(ch BytesChannel) (int, error) {
	n, err := field.writeToBase(ch)
	if err != nil {
		return n, err
	}

	n2, err := field.writeToScale(ch)
	return n + n2, err
}

type BigDateTimeNFieldFmt struct{ fieldFmtLengthScale }
type BigTimeNFieldFmt struct{ fieldFmtLengthScale }

type fieldFmtLengthPrecisionScale struct {
	fieldFmtBase
	fieldFmtBasePrecision
	fieldFmtBaseScale
}

func (field fieldFmtLengthPrecisionScale) FormatByteLength() int {
	return 2 + field.lengthBytes
}

func (field *fieldFmtLengthPrecisionScale) ReadFrom(ch BytesChannel) (int, error) {
	n, err := field.readFromBase(ch)
	if err != nil {
		return n, err
	}

	n2, err := field.readFromPrecision(ch)
	if err != nil {
		return n + n2, err
	}

	n3, err := field.readFromScale(ch)
	return n + n2 + n3, err
}

func (field fieldFmtLengthPrecisionScale) WriteTo(ch BytesChannel) (int, error) {
	n, err := field.writeToBase(ch)
	if err != nil {
		return n, err
	}

	n2, err := field.writeToPrecision(ch)
	if err != nil {
		return n + n2, err
	}

	n3, err := field.writeToScale(ch)
	return n + n2 + n3, err
}

type DecNFieldFmt struct{ fieldFmtLengthPrecisionScale }
type NumNFieldFmt struct{ fieldFmtLengthPrecisionScale }

//go:generate stringer -type=BlobType
type BlobType uint8

const (
	TDS_BLOB_FULLCLASSNAME BlobType = 0x01
	TDS_BLOB_DBID_CLASSDEF BlobType = 0x02
	TDS_BLOB_CHAR          BlobType = 0x03
	TDS_BLOB_BINARY        BlobType = 0x04
	TDS_BLOB_UNICHAR       BlobType = 0x05
	TDS_LOBLOC_CHAR        BlobType = 0x06
	TDS_LOBLOC_BINARY      BlobType = 0x07
	TDS_LOBLOC_UNICHAR     BlobType = 0x08
)

//go:generate stringer -type=BlobSerializationType
type BlobSerializationType uint8

const (
	NativeJavaSerialization BlobSerializationType = iota
	NativeCharacterFormat
	BinaryData
	UnicharUTF16
	UnicharUTF8
	UnicharSCSU
)

type fieldFmtBlob struct {
	fieldFmtBase
	blobType BlobType
	classID  string
}

func (field fieldFmtBlob) FormatByteLength() int {
	return 1 + 1 + len(field.classID) + field.lengthBytes
}

func (field *fieldFmtBlob) ReadFrom(ch BytesChannel) (int, error) {
	n, err := field.readFromBase(ch)
	if err != nil {
		return n, err
	}

	blobType, err := ch.Uint8()
	if err != nil {
		return n, fmt.Errorf("failed to read blobtype: %w", err)
	}
	field.blobType = BlobType(blobType)
	n++

	if field.blobType == TDS_BLOB_FULLCLASSNAME || field.blobType == TDS_BLOB_DBID_CLASSDEF {
		classIdLength, err := ch.Uint16()
		if err != nil {
			return n, fmt.Errorf("failed to read ClassID length: %w", err)
		}
		n += 2

		field.classID, err = ch.String(int(classIdLength))
		if err != nil {
			return n, fmt.Errorf("failed to read ClassID: %w", err)
		}
		n += int(classIdLength)
	}

	return n, nil
}

func (field fieldFmtBlob) WriteTo(ch BytesChannel) (int, error) {
	n, err := field.writeToBase(ch)
	if err != nil {
		return n, err
	}

	if err := ch.WriteUint8(uint8(field.blobType)); err != nil {
		return n, fmt.Errorf("failed to write blobtype: %w", err)
	}
	n++

	if field.blobType == TDS_BLOB_FULLCLASSNAME || field.blobType == TDS_BLOB_DBID_CLASSDEF {
		if err := ch.WriteUint16(uint16(len(field.classID))); err != nil {
			return n, fmt.Errorf("failed to write ClassID length: %w", err)
		}
		n += 2

		if len(field.classID) > 0 {
			if err := ch.WriteString(field.classID); err != nil {
				return n, fmt.Errorf("failed to write ClassID: %w", err)
			}
			n += len(field.classID)
		}
	}

	return n, nil
}

type BlobFieldFmt struct{ fieldFmtBlob }

type fieldData struct{ fieldDataBase }

func (field *fieldData) ReadFrom(ch BytesChannel) (int, error) {
	return field.readFrom(ch)
}

func (field fieldData) WriteTo(ch BytesChannel) (int, error) {
	return field.writeTo(ch)
}

type BitFieldData struct{ fieldData }
type DateTimeFieldData struct{ fieldData }
type DateFieldData struct{ fieldData }
type ShortDateFieldData struct{ fieldData }
type Flt4FieldData struct{ fieldData }
type Flt8FieldData struct{ fieldData }
type Int1FieldData struct{ fieldData }
type Int2FieldData struct{ fieldData }
type Int4FieldData struct{ fieldData }
type Int8FieldData struct{ fieldData }
type IntervalFieldData struct{ fieldData }
type Sint1FieldData struct{ fieldData }
type Uint2FieldData struct{ fieldData }
type Uint4FieldData struct{ fieldData }
type Uint8FieldData struct{ fieldData }
type MoneyFieldData struct{ fieldData }
type ShortMoneyFieldData struct{ fieldData }
type TimeFieldData struct{ fieldData }
type BinaryFieldData struct{ fieldData }
type BoundaryFieldData struct{ fieldData }
type CharFieldData struct{ fieldData }
type DateNFieldData struct{ fieldData }
type DateTimeNFieldData struct{ fieldData }
type FltNFieldData struct{ fieldData }
type IntNFieldData struct{ fieldData }
type UintNFieldData struct{ fieldData }
type LongBinaryFieldData struct{ fieldData }
type LongCharFieldData struct{ fieldData }
type MoneyNFieldData struct{ fieldData }
type SensitivityFieldData struct{ fieldData }
type TimeNFieldData struct{ fieldData }
type VarBinaryFieldData struct{ fieldData }
type VarCharFieldData struct{ fieldData }
type BigDateTimeNFieldData struct{ fieldData }
type BigTimeNFieldData struct{ fieldData }
type DecNFieldData struct{ fieldData }
type NumNFieldData struct{ fieldData }

type fieldDataBlob struct {
	fieldData
	serializationType BlobSerializationType
	subClassID        string
	locator           string
}

const fieldDataBlobHighBit uint32 = 0x80000000

func (field *fieldDataBlob) ReadFrom(ch BytesChannel) (int, error) {
	fieldFmt, ok := field.fmt.(*BlobFieldFmt)
	if !ok {
		return 0, fmt.Errorf("field.fmt is not of type BlobFieldfmt")
	}

	n, err := field.readFromStatus(ch)
	if err != nil {
		return n, err
	}

	serialization, err := ch.Uint8()
	if err != nil {
		return n, fmt.Errorf("failed to read serialization type: %w", err)
	}
	n++

	switch serialization {
	case 0:
		switch fieldFmt.blobType {
		case TDS_BLOB_FULLCLASSNAME, TDS_BLOB_DBID_CLASSDEF:
			field.serializationType = NativeJavaSerialization
		case TDS_BLOB_CHAR:
			field.serializationType = NativeCharacterFormat
		case TDS_BLOB_BINARY:
			field.serializationType = BinaryData
		case TDS_BLOB_UNICHAR:
			field.serializationType = UnicharUTF16
		}
	case 1:
		if fieldFmt.blobType != TDS_BLOB_UNICHAR {
			return n, fmt.Errorf("invalid blob (%s) and serialization (%d) type combination",
				fieldFmt.blobType, serialization)
		}
		field.serializationType = UnicharUTF8
	case 2:
		if fieldFmt.blobType != TDS_BLOB_UNICHAR {
			return n, fmt.Errorf("invalid blob (%s) and serialization (%d) type combination",
				fieldFmt.blobType, serialization)
		}
		field.serializationType = UnicharSCSU
	default:
		return n, fmt.Errorf("unhandled serialization type %d", serialization)
	}

	switch fieldFmt.blobType {
	case TDS_BLOB_FULLCLASSNAME, TDS_BLOB_DBID_CLASSDEF:
		subClassIdLength, err := ch.Uint16()
		if err != nil {
			return n, fmt.Errorf("failed to read SubClassID length: %w", err)
		}
		n += 2

		if subClassIdLength > 0 {
			field.subClassID, err = ch.String(int(subClassIdLength))
			if err != nil {
				return n, fmt.Errorf("failed to read SubClassID: %w", err)
			}
			n += int(subClassIdLength)
		}
	case TDS_LOBLOC_CHAR, TDS_LOBLOC_BINARY, TDS_LOBLOC_UNICHAR:
		locatorLength, err := ch.Uint16()
		if err != nil {
			return n, fmt.Errorf("failed to read locator length: %w", err)
		}
		n += 2

		field.locator, err = ch.String(int(locatorLength))
		if err != nil {
			return n, fmt.Errorf("failed to read locator: %w", err)
		}
		n += int(locatorLength)
	}

	for {
		dataLen, err := ch.Uint32()
		if err != nil {
			return n, fmt.Errorf("failed to read data length: %w", err)
		}
		n += 4

		// extract high bit:
		// 0 -> last data set
		// 1 -> another data set follows
		highBitSet := dataLen&fieldDataBlobHighBit == fieldDataBlobHighBit
		dataLen = dataLen &^ fieldDataBlobHighBit

		// if high bit is set and dataLen is zero no data array follows,
		// instead read the next data length immediately
		if highBitSet {
			break
		}

		if dataLen == 0 {
			continue
		}

		data, err := ch.Bytes(int(dataLen))
		if err != nil {
			return n, fmt.Errorf("failed to read data array: %w", err)
		}
		n += int(dataLen)

		// TODO this is inefficient for large datasets - must be
		// replaced by a low-overhead extensible byte storage (so - not
		// a slice)
		field.data = append(field.data, data...)
	}

	return n, nil
}

func (field fieldDataBlob) Writeto(ch BytesChannel) (int, error) {
	fieldFmt, ok := field.fmt.(*BlobFieldFmt)
	if !ok {
		return 0, fmt.Errorf("field.fmt is not of type BlobFieldFmt")
	}

	n, err := field.writeToStatus(ch)
	if err != nil {
		return n, err
	}

	var serialization uint8
	switch field.serializationType {
	case NativeJavaSerialization, NativeCharacterFormat, BinaryData, UnicharUTF16:
		serialization = 0
	case UnicharUTF8:
		serialization = 1
	case UnicharSCSU:
		serialization = 2
	}
	if err := ch.WriteUint8(serialization); err != nil {
		return n, fmt.Errorf("failed to write SerializationType: %w", err)
	}
	n++

	switch fieldFmt.blobType {
	case TDS_BLOB_FULLCLASSNAME, TDS_BLOB_DBID_CLASSDEF:
		if err := ch.WriteUint16(uint16(len(field.subClassID))); err != nil {
			return n, fmt.Errorf("failed to write SubClassID length: %w", err)
		}
		n += 2

		if err := ch.WriteString(field.subClassID); err != nil {
			return n, fmt.Errorf("failed to write SubClassID: %w", err)
		}
		n += len(field.subClassID)
	case TDS_LOBLOC_CHAR, TDS_LOBLOC_BINARY, TDS_LOBLOC_UNICHAR:
		if err := ch.WriteUint16(uint16(len(field.locator))); err != nil {
			return n, fmt.Errorf("failed to write Locator length: %w", err)
		}
		n += 2

		if err := ch.WriteString(field.locator); err != nil {
			return n, fmt.Errorf("failed to write Locator: %w", err)
		}
		n += len(field.locator)
	}

	dataLen := 1024
	if dataLen > len(field.data) {
		dataLen = len(field.data)
	}

	start, end := 0, dataLen
	for {
		passLen := uint32(dataLen)
		if end == len(field.data) {
			passLen |= fieldDataBlobHighBit
		}

		if err := ch.WriteUint32(uint32(passLen)); err != nil {
			return n, fmt.Errorf("failed to write data chunk length: %w", err)
		}
		n += 4

		if err := ch.WriteBytes(field.data[start:end]); err != nil {
			return n, fmt.Errorf("failed to write %d bytes of data: %w", dataLen, err)
		}
		n += end - start

		if end == len(field.data) {
			break
		}

		start = end
		end += dataLen
	}

	return n, nil
}

type BlobFieldData struct{ fieldDataBlob }

type fieldFmtTxtPtr struct {
	fieldFmtBase

	tableName string
}

func (field fieldFmtTxtPtr) FormatByteLength() int {
	return 1 + len(field.tableName) + field.lengthBytes
}

func (field *fieldFmtTxtPtr) ReadFrom(ch BytesChannel) (int, error) {
	n, err := field.readFromBase(ch)
	if err != nil {
		return n, err
	}

	nameLength, err := ch.Uint16()
	if err != nil {
		return n, fmt.Errorf("failed to read name length: %w", err)
	}
	n += 2

	field.name, err = ch.String(int(nameLength))
	if err != nil {
		return n, fmt.Errorf("failed to read name: %w", err)
	}
	n += int(nameLength)

	return n, nil
}

func (field fieldFmtTxtPtr) WriteTo(ch BytesChannel) (int, error) {
	n, err := field.writeToBase(ch)
	if err != nil {
		return n, err
	}

	if err := ch.WriteUint16(uint16(len(field.name))); err != nil {
		return n, fmt.Errorf("failed to write Name length: %w", err)
	}
	n += 2

	if err := ch.WriteString(field.name); err != nil {
		return n, fmt.Errorf("failed to write Name: %w", err)
	}
	n += len(field.name)

	return n, nil
}

type ImageFieldFmt struct{ fieldFmtTxtPtr }
type TextFieldFmt struct{ fieldFmtTxtPtr }
type UniTextFieldFmt struct{ fieldFmtTxtPtr }
type XMLFieldFmt struct{ fieldFmtTxtPtr }

type fieldDataTxtPtr struct {
	fieldData

	txtPtr    []byte
	timeStamp []byte
}

func (field *fieldDataTxtPtr) ReadFrom(ch BytesChannel) (int, error) {
	n, err := field.readFromStatus(ch)
	if err != nil {
		return n, err
	}

	txtPtrLen, err := ch.Uint8()
	if err != nil {
		return n, fmt.Errorf("failed to read TxtPtrLen: %w", err)
	}
	n++

	field.txtPtr, err = ch.Bytes(int(txtPtrLen))
	if err != nil {
		return n, fmt.Errorf("failed to read TxtPtr: %w", err)
	}
	n += int(txtPtrLen)

	field.timeStamp, err = ch.Bytes(8)
	if err != nil {
		return n, fmt.Errorf("failed to read TimeStamp: %w", err)
	}
	n += 8

	dataLen, err := ch.Uint32()
	if err != nil {
		return n, fmt.Errorf("failed to read data length: %w", err)
	}
	n += 4

	field.data, err = ch.Bytes(int(dataLen))
	if err != nil {
		return n, fmt.Errorf("failed to read data: %w", err)
	}
	n += int(dataLen)

	return n, nil
}

func (field fieldDataTxtPtr) WriteTo(ch BytesChannel) (int, error) {
	n, err := field.writeToStatus(ch)
	if err != nil {
		return n, err
	}

	if err := ch.WriteUint8(uint8(len(field.txtPtr))); err != nil {
		return n, fmt.Errorf("failed to write TxtPtr length: %w", err)
	}
	n++

	if err := ch.WriteBytes(field.txtPtr); err != nil {
		return n, fmt.Errorf("failed to write TxtPtr: %w", err)
	}
	n += len(field.txtPtr)

	if err := ch.WriteBytes(field.timeStamp); err != nil {
		return n, fmt.Errorf("failed to write TimeStamp: %w", err)
	}
	n += len(field.timeStamp)

	if err := ch.WriteUint32(uint32(len(field.data))); err != nil {
		return n, fmt.Errorf("failed to write Data length: %w", err)
	}
	n += 4

	if err := ch.WriteBytes(field.data); err != nil {
		return n, fmt.Errorf("failed to write Data: %w", err)
	}
	n += len(field.data)

	return n, nil
}

type ImageFieldData struct{ fieldDataTxtPtr }
type TextFieldData struct{ fieldDataTxtPtr }
type UniTextFieldData struct{ fieldDataTxtPtr }
type XMLFieldData struct{ fieldDataTxtPtr }

// utils

func readLengthBytes(ch BytesChannel, n int) (int, error) {
	var length int
	var err error
	switch n {
	case 4:
		var tmp uint32
		tmp, err = ch.Uint32()
		length = int(tmp)
	case 2:
		var tmp uint16
		tmp, err = ch.Uint16()
		length = int(tmp)
	default:
		var tmp uint8
		tmp, err = ch.Uint8()
		length = int(tmp)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to read length: %w", err)
	}

	return length, nil
}

func writeLengthBytes(ch BytesChannel, byteCount int, n int) error {
	var err error
	switch byteCount {
	case 4:
		err = ch.WriteUint32(uint32(n))
	case 2:
		err = ch.WriteUint16(uint16(n))
	default:
		err = ch.WriteUint8(uint8(n))
	}

	if err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}
	return nil
}

// LookupFieldFmt returns the FieldFmt for a given data type and sets
// required values in it.
func LookupFieldFmt(dataType DataType) (FieldFmt, error) {
	switch dataType {
	case TDS_BIGDATETIMEN:
		v := &BigDateTimeNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_BIGTIMEN:
		v := &BigTimeNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_BIT:
		v := &BitFieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 1
		return v, nil
	case TDS_DATETIME:
		v := &DateTimeFieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 8
		return v, nil
	case TDS_DATE:
		v := &DateFieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 4
		return v, nil
	case TDS_SHORTDATE:
		v := &ShortDateFieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 4
		return v, nil
	case TDS_FLT4:
		v := &Flt4FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 4
		return v, nil
	case TDS_FLT8:
		v := &Flt8FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 8
		return v, nil
	case TDS_INT1:
		v := &Int1FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 1
		return v, nil
	case TDS_INT2:
		v := &Int2FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 2
		return v, nil
	case TDS_INT4:
		v := &Int4FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 4
		return v, nil
	case TDS_INT8:
		v := &Int8FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 8
		return v, nil
	case TDS_INTERVAL:
		v := &IntervalFieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 8
		return v, nil
	case TDS_SINT1:
		v := &Sint1FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 1
		return v, nil
	case TDS_UINT2:
		v := &Uint2FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 2
		return v, nil
	case TDS_UINT4:
		v := &Uint4FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 4
		return v, nil
	case TDS_UINT8:
		v := &Uint8FieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 8
		return v, nil
	case TDS_MONEY:
		v := &MoneyFieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 8
		return v, nil
	case TDS_SHORTMONEY:
		v := &ShortMoneyFieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 4
		return v, nil
	case TDS_TIME:
		v := &TimeFieldFmt{}
		v.dataType = dataType
		v.isFixedLength = true
		v.lengthBytes = 4
		return v, nil
	case TDS_BINARY:
		v := &BinaryFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_BOUNDARY:
		v := &BoundaryFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_CHAR:
		v := &CharFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_DATEN:
		v := &DateNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_DATETIMEN:
		v := &DateTimeNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_FLTN:
		v := &FltNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_INTN:
		v := &IntNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_UINTN:
		v := &UintNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_LONGBINARY:
		v := &LongBinaryFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 4
		v.maxLength = 2147483647
		return v, nil
	case TDS_LONGCHAR:
		v := &LongCharFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 4
		return v, nil
	case TDS_MONEYN:
		v := &MoneyNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_SENSITIVITY:
		v := &SensitivityFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_TIMEN:
		v := &TimeNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_VARBINARY:
		v := &VarBinaryFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_VARCHAR:
		v := &VarCharFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		v.maxLength = 255
		return v, nil
	case TDS_DECN:
		v := &DecNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_NUMN:
		v := &NumNFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_BLOB:
		v := &BlobFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_IMAGE:
		v := &ImageFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 4
		return v, nil
	case TDS_TEXT:
		v := &TextFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 4
		return v, nil
	case TDS_UNITEXT:
		v := &UniTextFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 4
		return v, nil
	case TDS_XML:
		v := &XMLFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 4
		return v, nil
	default:
		return nil, fmt.Errorf("unhandled datatype '%s'", dataType)
	}
}

/// LookupFieldData returns the FieldData for a given field format.
func LookupFieldData(fieldFmt FieldFmt) (FieldData, error) {
	switch fieldFmt.DataType() {
	case TDS_BIGDATETIMEN:
		v := &BigDateTimeNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_BIGTIMEN:
		v := &BigTimeNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_BIT:
		v := &BitFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_DATETIME:
		v := &DateTimeFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_DATE:
		v := &DateFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_SHORTDATE:
		v := &ShortDateFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_FLT4:
		v := &Flt4FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_FLT8:
		v := &Flt8FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_INT1:
		v := &Int1FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_INT2:
		v := &Int2FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_INT4:
		v := &Int4FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_INT8:
		v := &Int8FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_INTERVAL:
		v := &IntervalFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_SINT1:
		v := &Sint1FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_UINT2:
		v := &Uint2FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_UINT4:
		v := &Uint4FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_UINT8:
		v := &Uint8FieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_MONEY:
		v := &MoneyFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_SHORTMONEY:
		v := &ShortMoneyFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_TIME:
		v := &TimeFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_BINARY:
		v := &BinaryFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_BOUNDARY:
		v := &BoundaryFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_CHAR:
		v := &CharFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_DATEN:
		v := &DateNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_DATETIMEN:
		v := &DateTimeNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_FLTN:
		v := &FltNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_INTN:
		v := &IntNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_UINTN:
		v := &UintNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_LONGBINARY:
		v := &LongBinaryFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_LONGCHAR:
		v := &LongCharFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_MONEYN:
		v := &MoneyNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_SENSITIVITY:
		v := &SensitivityFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_TIMEN:
		v := &TimeNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_VARBINARY:
		v := &VarBinaryFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_VARCHAR:
		v := &VarCharFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_DECN:
		v := &DecNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_NUMN:
		v := &NumNFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_BLOB:
		v := &BlobFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_IMAGE:
		v := &ImageFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_TEXT:
		v := &TextFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_UNITEXT:
		v := &UniTextFieldData{}
		v.fmt = fieldFmt
		return v, nil
	case TDS_XML:
		v := &XMLFieldData{}
		v.fmt = fieldFmt
		return v, nil
	default:
		return nil, fmt.Errorf("unhandled datatype: '%s'", fieldFmt.DataType())
	}
}

// LookupFieldFmtData returns both Fieldfmt and FieldData for a given
// data type.
func LookupFieldFmtData(dataType DataType) (FieldFmt, FieldData, error) {
	fieldFmt, err := LookupFieldFmt(dataType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find field format: %w", err)
	}

	data, err := LookupFieldData(fieldFmt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find field data: %w", err)
	}

	return fieldFmt, data, nil
}
