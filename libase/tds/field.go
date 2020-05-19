package tds

import "fmt"

// Interfaces

type FieldFmt interface {
	// Format information send by TDS server
	// The setters are required to prevent type assertions
	DataType() DataType
	SetName(string)
	Name() string
	SetStatus(DataFieldStatus)
	Status() DataFieldStatus
	SetUserType(int32)
	UserType() int32
	SetLocaleInfo(string)
	LocaleInfo() string

	LengthBytes() int
	SetLength(int)
	Length() int

	// Interface methods for go-ase
	ReadFrom(*channel) error
	WriteTo(*channel) error

	FormatByteLength() int
}

type FieldData interface {
	// Format information send by TDS server
	Status() DataFieldStatus

	// Interface methods for go-ase
	SetData([]byte)
	Data() []byte
	ReadFrom(*channel) error
	WriteTo(*channel) error
}

// Base structs

type fieldFmtBase struct {
	dataType   DataType
	name       string
	status     DataFieldStatus
	userType   int32
	localeInfo string

	lengthBytes int
	length      int
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

func (field *fieldFmtBase) SetStatus(status DataFieldStatus) {
	field.status = status
}

func (field fieldFmtBase) Status() DataFieldStatus {
	return field.status
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

func (field fieldFmtBase) LengthBytes() int {
	return field.lengthBytes
}

func (field *fieldFmtBase) SetLength(length int) {
	field.length = length
}

func (field fieldFmtBase) Length() int {
	return field.length
}

func (field *fieldFmtBase) readFromBase(ch *channel) error {
	if field.lengthBytes == 0 {
		return nil
	}

	length, err := readLengthBytes(ch, field.lengthBytes)
	if err != nil {
		return err
	}
	field.length = length

	return nil
}

func (field fieldFmtBase) writeToBase(ch *channel) error {
	if field.lengthBytes == 0 {
		return nil
	}

	return writeLengthBytes(ch, field.lengthBytes, field.length)
}

type fieldFmtBasePrecision struct {
	precision uint8
}

func (field fieldFmtBasePrecision) Precision() uint8 {
	return field.precision
}

func (field *fieldFmtBasePrecision) readFromPrecision(ch *channel) error {
	var err error
	field.precision, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read precision: %w", err)
	}
	return nil
}

func (field fieldFmtBasePrecision) writeToPrecision(ch *channel) error {
	err := ch.WriteUint8(field.precision)
	if err != nil {
		return fmt.Errorf("failed to write precision: %w", err)
	}
	return nil
}

type fieldFmtBaseScale struct {
	scale uint8
}

func (field fieldFmtBaseScale) Scale() uint8 {
	return field.scale
}

func (field *fieldFmtBaseScale) readFromScale(ch *channel) error {
	var err error
	field.scale, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read scale: %w", err)
	}
	return nil
}

func (field fieldFmtBaseScale) writeToScale(ch *channel) error {
	err := ch.WriteUint8(field.scale)
	if err != nil {
		return fmt.Errorf("failed to write scale: %w", err)
	}
	return nil
}

type fieldDataBase struct {
	fmt    FieldFmt
	status DataFieldStatus
	data   []byte
}

func (field fieldDataBase) Status() DataFieldStatus {
	return field.status
}

func (field *fieldDataBase) SetData(data []byte) {
	field.data = data
}

func (field fieldDataBase) Data() []byte {
	return field.data
}

func (field *fieldDataBase) readFromStatus(ch *channel) error {
	if field.fmt.Status()&TDS_PARAM_COLUMNSTATUS != TDS_PARAM_COLUMNSTATUS {
		return nil
	}

	status, err := ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read status: %w", err)
	}
	field.status = DataFieldStatus(status)
	return nil
}

func (field fieldDataBase) writeToStatus(ch *channel) error {
	if field.fmt.Status()&TDS_PARAM_COLUMNSTATUS != TDS_PARAM_COLUMNSTATUS {
		return nil
	}

	err := ch.WriteUint8(uint8(field.status))
	if err != nil {
		return fmt.Errorf("failed to write status: %w", err)
	}
	return nil
}

func (field *fieldDataBase) readFrom(ch *channel) error {
	if err := field.readFromStatus(ch); err != nil {
		return err
	}

	if field.fmt.Length() == 0 {
		return nil
	}

	length := field.fmt.Length()
	if field.fmt.LengthBytes() > 0 {
		var err error
		length, err = readLengthBytes(ch, field.fmt.LengthBytes())
		if err != nil {
			return err
		}
	}

	var err error
	if field.data, err = ch.Bytes(length); err != nil {
		return fmt.Errorf("failed to read %d bytes of data: %w", length, err)
	}
	return nil
}

func (field fieldDataBase) writeTo(ch *channel) error {
	if err := field.writeToStatus(ch); err != nil {
		return err
	}

	if field.fmt.Length() == 0 {
		return nil
	}

	if field.fmt.LengthBytes() > 0 {
		err := writeLengthBytes(ch, field.fmt.LengthBytes(), len(field.data))
		if err != nil {
			return err
		}
	}

	err := ch.WriteBytes(field.data)
	if err != nil {
		return fmt.Errorf("failed to write %d bytes of data: %w", len(field.data), err)
	}
	return nil
}

// Implementations

type fieldFmtLength struct {
	fieldFmtBase
}

func (field fieldFmtLength) FormatByteLength() int {
	if field.lengthBytes > 0 {
		return field.lengthBytes
	}
	return 0
}

func (field *fieldFmtLength) ReadFrom(ch *channel) error {
	return field.readFromBase(ch)
}

func (field fieldFmtLength) WriteTo(ch *channel) error {
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
	if field.lengthBytes > 0 {
		return 1 + field.lengthBytes
	}
	return 1
}

func (field *fieldFmtLengthScale) ReadFrom(ch *channel) error {
	if err := field.readFromBase(ch); err != nil {
		return err
	}
	return field.readFromScale(ch)
}

func (field fieldFmtLengthScale) WriteTo(ch *channel) error {
	if err := field.writeToBase(ch); err != nil {
		return err
	}
	return field.writeToScale(ch)
}

type BigDateTimeNFieldFmt struct{ fieldFmtLengthScale }
type BigTimeNFieldFmt struct{ fieldFmtLengthScale }

type fieldFmtLengthPrecisionScale struct {
	fieldFmtBase
	fieldFmtBasePrecision
	fieldFmtBaseScale
}

func (field fieldFmtLengthPrecisionScale) FormatByteLength() int {
	if field.lengthBytes > 0 {
		return 2 + field.lengthBytes
	}
	return 2
}

func (field *fieldFmtLengthPrecisionScale) ReadFrom(ch *channel) error {
	if err := field.readFromBase(ch); err != nil {
		return err
	}
	if err := field.readFromPrecision(ch); err != nil {
		return err
	}
	return field.readFromScale(ch)
}

func (field fieldFmtLengthPrecisionScale) WriteTo(ch *channel) error {
	if err := field.writeToBase(ch); err != nil {
		return err
	}
	if err := field.writeToPrecision(ch); err != nil {
		return err
	}
	return field.writeToScale(ch)
}

type DecNFieldFmt struct{ fieldFmtLengthPrecisionScale }
type NumNFieldFmt struct{ fieldFmtLengthPrecisionScale }

//go:generate stringer -type=BlobType
type BlobType uint8

const (
	TDS_BLOB_FULLCLASSNAME BlobType = 0x01
	TDS_BLOB_DBID_CLASSDEF          = 0x02
	TDS_BLOB_CHAR                   = 0x03
	TDS_BLOB_BINARY                 = 0x04
	TDS_BLOB_UNICHAR                = 0x05
	TDS_LOBLOC_CHAR                 = 0x06
	TDS_LOBLOC_BINARY               = 0x07
	TDS_LOBLOC_UNICHAR              = 0x08
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
	base := 1 + 1 + len(field.classID)
	if field.lengthBytes > 0 {
		return base + field.lengthBytes
	}
	return base
}

func (field *fieldFmtBlob) ReadFrom(ch *channel) error {
	blobType, err := ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read blobtype: %w", err)
	}
	field.blobType = BlobType(blobType)

	if field.blobType == TDS_BLOB_FULLCLASSNAME || field.blobType == TDS_BLOB_DBID_CLASSDEF {
		classIdLength, err := ch.Uint16()
		if err != nil {
			return fmt.Errorf("failed to read ClassID length: %w", err)
		}

		if classIdLength > 0 {
			field.classID, err = ch.String(int(classIdLength))
			if err != nil {
				return fmt.Errorf("failed to read ClassID: %w", err)
			}
		}
	}

	return nil
}

func (field fieldFmtBlob) WriteTo(ch *channel) error {
	if err := ch.WriteUint8(uint8(field.blobType)); err != nil {
		return fmt.Errorf("failed to write blobtype: %w", err)
	}

	if field.blobType == TDS_BLOB_FULLCLASSNAME || field.blobType == TDS_BLOB_DBID_CLASSDEF {
		if err := ch.WriteUint16(uint16(len(field.classID))); err != nil {
			return fmt.Errorf("failed to write ClassID length: %w", err)
		}

		if len(field.classID) > 0 {
			if err := ch.WriteString(field.classID); err != nil {
				return fmt.Errorf("failed to write ClassID: %w", err)
			}
		}
	}

	return nil
}

type BlobFieldFmt struct{ fieldFmtBlob }

type fieldData struct{ fieldDataBase }

func (field *fieldData) ReadFrom(ch *channel) error {
	return field.readFrom(ch)
}

func (field fieldData) WriteTo(ch *channel) error {
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

func (field *fieldDataBlob) ReadFrom(ch *channel) error {
	fieldFmt := field.fmt.(*BlobFieldFmt)

	if err := field.readFromStatus(ch); err != nil {
		return err
	}

	serialization, err := ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read serialization type: %w", err)
	}

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
			return fmt.Errorf("invalid blob (%s) and serialization (%d) type combination",
				fieldFmt.blobType, serialization)
		}
		field.serializationType = UnicharUTF8
	case 2:
		if fieldFmt.blobType != TDS_BLOB_UNICHAR {
			return fmt.Errorf("invalid blob (%s) and serialization (%d) type combination",
				fieldFmt.blobType, serialization)
		}
		field.serializationType = UnicharSCSU
	default:
		return fmt.Errorf("unhandled serialization type %d", serialization)
	}

	switch fieldFmt.blobType {
	case TDS_BLOB_FULLCLASSNAME, TDS_BLOB_DBID_CLASSDEF:
		subClassIdLength, err := ch.Uint16()
		if err != nil {
			return fmt.Errorf("failed to read SubClassID length: %w", err)
		}

		if subClassIdLength > 0 {
			field.subClassID, err = ch.String(int(subClassIdLength))
			if err != nil {
				return fmt.Errorf("failed to read SubClassID: %w", err)
			}
		}
	case TDS_LOBLOC_CHAR, TDS_LOBLOC_BINARY, TDS_LOBLOC_UNICHAR:
		locatorLength, err := ch.Uint16()
		if err != nil {
			return fmt.Errorf("failed to read locator length: %w", err)
		}

		field.locator, err = ch.String(int(locatorLength))
		if err != nil {
			return fmt.Errorf("failed to read locator: %w", err)
		}
	default:
	}

	for {
		dataLen, err := ch.Uint32()
		if err != nil {
			return fmt.Errorf("failed to read data length: %w", err)
		}

		// extract high bit:
		// 0 -> last data set
		// 1 -> another data set follows
		highBitSet := dataLen&fieldDataBlobHighBit == fieldDataBlobHighBit
		dataLen = dataLen &^ fieldDataBlobHighBit

		// if high bit is set and dataLen is zero no data array follows,
		// instead read the next data length immediately
		if highBitSet && dataLen == 0 {
			break
		}

		data, err := ch.Bytes(int(dataLen))
		if err != nil {
			return fmt.Errorf("failed to read data array: %w", err)
		}

		// TODO this is inefficient for large datasets - must be
		// replaced by a low-overhead extensible byte storage (so - not
		// a slice)
		field.data = append(field.data, data...)
	}

	return nil
}

func (field fieldDataBlob) Writeto(ch *channel) error {
	fieldFmt := field.fmt.(*BlobFieldFmt)

	if err := field.writeToStatus(ch); err != nil {
		return err
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
		return fmt.Errorf("failed to write SerializationType: %w", err)
	}

	switch fieldFmt.blobType {
	case TDS_BLOB_FULLCLASSNAME, TDS_BLOB_DBID_CLASSDEF:
		if err := ch.WriteUint16(uint16(len(field.subClassID))); err != nil {
			return fmt.Errorf("failed to write SubClassID length: %w", err)
		}

		if len(field.subClassID) > 0 {
			if err := ch.WriteString(field.subClassID); err != nil {
				return fmt.Errorf("failed to write SubClassID: %w", err)
			}
		}
	case TDS_LOBLOC_CHAR, TDS_LOBLOC_BINARY, TDS_LOBLOC_UNICHAR:
		if err := ch.WriteUint16(uint16(len(field.locator))); err != nil {
			return fmt.Errorf("failed to write Locator length: %w", err)
		}

		if len(field.locator) > 0 {
			if err := ch.WriteString(field.locator); err != nil {
				return fmt.Errorf("failed to write Locator: %w", err)
			}
		}
	default:
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
			return fmt.Errorf("failed to write data chunk length: %w", err)
		}

		if err := ch.WriteBytes(field.data[start:end]); err != nil {
			return fmt.Errorf("failed to write %d bytes of data: %w", dataLen, err)
		}

		if end == len(field.data) {
			break
		}

		start = end
		end += dataLen
	}

	return nil
}

type BlobFieldData struct{ fieldDataBlob }

type fieldFmtTxtPtr struct {
	fieldFmtBase

	tableName string
}

func (field fieldFmtTxtPtr) FormatByteLength() int {
	base := 1 + len(field.tableName)
	if field.lengthBytes > 0 {
		return base + field.lengthBytes
	}
	return base
}

func (field *fieldFmtTxtPtr) ReadFrom(ch *channel) error {
	if err := field.readFromBase(ch); err != nil {
		return err
	}

	nameLength, err := ch.Uint16()
	if err != nil {
		return fmt.Errorf("failed to read name length: %w", err)
	}

	field.name, err = ch.String(int(nameLength))
	if err != nil {
		return fmt.Errorf("failed to read name: %w", err)
	}

	return nil
}

func (field fieldFmtTxtPtr) WriteTo(ch *channel) error {
	if err := field.writeToBase(ch); err != nil {
		return err
	}

	if err := ch.WriteUint16(uint16(len(field.name))); err != nil {
		return fmt.Errorf("failed to write Name length: %w", err)
	}

	if len(field.name) > 0 {
		if err := ch.WriteString(field.name); err != nil {
			return fmt.Errorf("failed to write Name: %w", err)
		}
	}

	return nil
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

func (field *fieldDataTxtPtr) ReadFrom(ch *channel) error {
	err := field.readFromStatus(ch)
	if err != nil {
		return err
	}

	txtPtrLen, err := ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read TxtPtrLen: %w", err)
	}

	field.txtPtr, err = ch.Bytes(int(txtPtrLen))
	if err != nil {
		return fmt.Errorf("failed to read TxtPtr: %w", err)
	}

	field.timeStamp, err = ch.Bytes(8)
	if err != nil {
		return fmt.Errorf("failed to read TimeStamp: %w", err)
	}

	dataLen, err := ch.Uint32()
	if err != nil {
		return fmt.Errorf("failed to read data length: %w", err)
	}

	field.data, err = ch.Bytes(int(dataLen))
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	return nil
}

func (field fieldDataTxtPtr) WriteTo(ch *channel) error {
	if err := field.writeToStatus(ch); err != nil {
		return err
	}

	if err := ch.WriteUint8(uint8(len(field.txtPtr))); err != nil {
		return fmt.Errorf("failed to write TxtPtr length: %w", err)
	}

	if err := ch.WriteBytes(field.txtPtr); err != nil {
		return fmt.Errorf("failed to write TxtPtr: %w", err)
	}

	if err := ch.WriteBytes(field.timeStamp); err != nil {
		return fmt.Errorf("failed to write TimeStamp: %w", err)
	}

	if err := ch.WriteUint32(uint32(len(field.data))); err != nil {
		return fmt.Errorf("failed to write Data length: %w", err)
	}

	if err := ch.WriteBytes(field.data); err != nil {
		return fmt.Errorf("failed to write Data: %w", err)
	}

	return nil
}

type ImageFieldData struct{ fieldDataTxtPtr }
type TextFieldData struct{ fieldDataTxtPtr }
type UniTextFieldData struct{ fieldDataTxtPtr }
type XMLFieldData struct{ fieldDataTxtPtr }

// utils

func readLengthBytes(ch *channel, n int) (int, error) {
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

func writeLengthBytes(ch *channel, byteCount int, n int) error {
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

func LookupFieldFmt(dataType DataType) (FieldFmt, error) {
	switch dataType {
	case TDS_BIGDATETIMEN:
		v := &BigDateTimeNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_BIGTIMEN:
		v := &BigTimeNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_BIT:
		v := &BitFieldFmt{}
		v.dataType = dataType
		v.length = 1
		return v, nil
	case TDS_DATETIME:
		v := &DateTimeFieldFmt{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_DATE:
		v := &DateFieldFmt{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_SHORTDATE:
		v := &ShortDateFieldFmt{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_FLT4:
		v := &Flt4FieldFmt{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_FLT8:
		v := &Flt8FieldFmt{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_INT1:
		v := &Int1FieldFmt{}
		v.dataType = dataType
		v.length = 1
		return v, nil
	case TDS_INT2:
		v := &Int2FieldFmt{}
		v.dataType = dataType
		v.length = 2
		return v, nil
	case TDS_INT4:
		v := &Int4FieldFmt{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_INT8:
		v := &Int8FieldFmt{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_INTERVAL:
		v := &IntervalFieldFmt{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_SINT1:
		v := &Sint1FieldFmt{}
		v.dataType = dataType
		v.length = 1
		return v, nil
	case TDS_UINT2:
		v := &Uint2FieldFmt{}
		v.dataType = dataType
		v.length = 2
		return v, nil
	case TDS_UINT4:
		v := &Uint4FieldFmt{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_UINT8:
		v := &Uint8FieldFmt{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_MONEY:
		v := &MoneyFieldFmt{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_SHORTMONEY:
		v := &ShortMoneyFieldFmt{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_TIME:
		v := &TimeFieldFmt{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_BINARY:
		v := &BinaryFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_BOUNDARY:
		v := &BoundaryFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_CHAR:
		v := &CharFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_DATEN:
		v := &DateNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_DATETIMEN:
		v := &DateTimeNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_FLTN:
		v := &FltNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_INTN:
		v := &IntNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_UINTN:
		v := &UintNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_LONGBINARY:
		v := &LongBinaryFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 4
		return v, nil
	case TDS_LONGCHAR:
		v := &LongCharFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_MONEYN:
		v := &MoneyNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_SENSITIVITY:
		v := &SensitivityFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_TIMEN:
		v := &TimeNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_VARBINARY:
		v := &VarBinaryFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_VARCHAR:
		v := &VarCharFieldFmt{}
		v.dataType = dataType
		v.lengthBytes = 1
		return v, nil
	case TDS_DECN:
		v := &DecNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_NUMN:
		v := &NumNFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_BLOB:
		v := &BlobFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_IMAGE:
		v := &ImageFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_TEXT:
		v := &TextFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_UNITEXT:
		v := &UniTextFieldFmt{}
		v.dataType = dataType
		return v, nil
	case TDS_XML:
		v := &XMLFieldFmt{}
		v.dataType = dataType
		return v, nil
	default:
		return nil, fmt.Errorf("unhandled datatype '%s'", dataType)
	}
	return nil, fmt.Errorf("unhandled datatype '%s'", dataType)
}

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
	}
	return nil, fmt.Errorf("unhandled datatype: '%s'", fieldFmt.DataType())
}
