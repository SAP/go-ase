package tds

import (
	"fmt"
)

// Interfaces
type FieldData interface {
	// Base information from ParamFmt
	Type() DataType

	HasUserType() bool
	SetUserType(uint32)
	UserType() uint32

	MaxLength() int
	Length() int

	HasStatus() bool
	SetStatus(DataFieldStatus)
	Status() DataFieldStatus

	SetName(string)
	Name() string

	SetLocaleInfo(string)
	LocaleInfo() string

	// Information from ParamFmt and Params
	// convenience from .Status()
	IsNullable() bool

	IsNull() bool

	// Data from Params - each top-level slice is one row containing the
	// byte slice send by the TDS server
	FieldStati() []DataFieldStatus
	Data() [][]byte

	readFormat(ch *channel) error
	readData(ch *channel) error
}

// Base structs

type fieldDataBase struct {
	dataType DataType
	// TODO maybe extra field for hasusertype, depends on the
	// definition of usertypes
	userType uint32

	lengthBytes int
	maxLength   int
	length      int

	columnStatus DataFieldStatus
	fieldStati   []DataFieldStatus
	data         [][]byte
}

func (field fieldDataBase) copyDataTo(other *fieldDataBase) error {
	other.data = make([]byte, len(field.data))
	copy(other.data, field.data)
	return nil
}

func (field fieldDataBase) Type() DataType {
	return field.dataType
}

func (field fieldDataBase) HasUserType() bool {
	return field.userType != 0
}

func (field fieldDataBase) UserType() uint32 {
	return field.userType
}

func (field fieldDataBase) MaxLength() int {
	return field.maxLength
}

func (field *fieldDataBase) readFormatLength(ch *channel) error {
	var err error
	switch field.lengthBytes {
	case 0:
		// if lengthBytes is zero the length is fixed and already set
		return nil
	case 4:
		var tmp uint32
		tmp, err = ch.Uint32()
		field.maxLength = int(tmp)
	case 2:
		var tmp uint16
		tmp, err = ch.Uint16()
		field.maxLength = int(tmp)
	default:
		var tmp uint8
		tmp, err = ch.Uint8()
		field.maxLength = int(tmp)
	}

	if err != nil {
		return fmt.Errorf("failed to read format length: %w", err)
	}

	return nil
}

func (field fieldDataBase) Length() int {
	return field.length
}

func (field *fieldDataBase) readDataLength(ch *channel) error {
	var err error
	switch field.lengthBytes {
	case 0:
		// if lengthBytes is zero the length is fixed and already set
		return nil
	case 4:
		var tmp uint32
		tmp, err = ch.Uint32()
		field.length = int(tmp)
	case 2:
		var tmp uint16
		tmp, err = ch.Uint16()
		field.length = int(tmp)
	default:
		var tmp uint8
		tmp, err = ch.Uint8()
		field.length = int(tmp)
	}
	if err != nil {
		return fmt.Errorf("failed to read data length: %w", err)
	}

	if field.length > field.maxLength {
		return fmt.Errorf("read field length %d is greater than max length %d",
			field.length, field.maxLength)
	}
	return nil
}

func (field fieldDataBase) HasStatus() bool {
	return field.columnStatus == noStatus
}

func (field fieldDataBase) Status() DataFieldStatus {
	return field.columnStatus
}

func (field *fieldDataBase) readDataStatus(ch *channel) error {
	if field.status&TDS_PARAM_COLUMNSTATUS != TDS_PARAM_COLUMNSTATUS {
		return nil
	}

	status, err := ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read status: %w", err)
	}
	field.status = DataFieldStatus(status)
	return nil
}

func (field fieldDataBase) IsNullable() bool {
	return field.status&TDS_PARAM_NULLALLOWED == TDS_PARAM_NULLALLOWED
}

func (field fieldDataBase) IsNull() bool {
	if field.data == nil {
		return true
	}

	if len(field.data) == 0 && field.IsNullable() {
		return true
	}

	return false
}

func (field fieldDataBase) Data() []byte {
	return field.data
}

func (field *fieldDataBase) readDataData(ch *channel) error {
	if field.length == 0 {
		return nil
	}

	var err error
	field.data, err = ch.Bytes(field.length)
	if err != nil {
		return fmt.Errorf("failed to read %d bytes of data: %w", field.length, err)
	}
	return nil
}

type fieldDataPrecision struct {
	precision uint8
}

func (field fieldDataPrecision) copyPrecision() fieldDataPrecision {
	return field
}

func (field *fieldDataPrecision) readFormatPrecision(ch *channel) error {
	var err error
	field.precision, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read precision: %w", err)
	}
	return nil
}

type fieldDataScale struct {
	scale uint8
}

func (field fieldDataScale) copyScale() fieldDataScale {
	return field
}

func (field *fieldDataScale) readFormatScale(ch *channel) error {
	var err error
	field.scale, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read scale: %w", err)
	}
	return nil
}

// Implementations

type fieldDataStatusData struct {
	fieldDataBase
}

func (field *fieldDataStatusData) readFormat(ch *channel) error {
	// Nothing to read
	return nil
}

func (field *fieldDataStatusData) readData(ch *channel) error {
	err := field.readDataStatus(ch)
	if err != nil {
		return err
	}
	return field.readDataData(ch)
}

type BitField struct{ fieldDataStatusData }
type DateTimeField struct{ fieldDataStatusData }
type DateField struct{ fieldDataStatusData }
type ShortDateField struct{ fieldDataStatusData }
type Flt4Field struct{ fieldDataStatusData }
type Flt8Field struct{ fieldDataStatusData }
type Int1Field struct{ fieldDataStatusData }
type Int2Field struct{ fieldDataStatusData }
type Int4Field struct{ fieldDataStatusData }
type Int8Field struct{ fieldDataStatusData }
type IntervalField struct{ fieldDataStatusData }
type Sint1Field struct{ fieldDataStatusData }
type Uint2Field struct{ fieldDataStatusData }
type Uint4Field struct{ fieldDataStatusData }
type Uint8Field struct{ fieldDataStatusData }
type MoneyField struct{ fieldDataStatusData }
type ShortMoneyField struct{ fieldDataStatusData }
type TimeField struct{ fieldDataStatusData }

type fieldDataLength struct {
	fieldDataBase
}

func (field *fieldDataLength) readFormat(ch *channel) error {
	return field.readFormatLength(ch)
}

func (field *fieldDataLength) readData(ch *channel) error {
	err := field.readDataStatus(ch)
	if err != nil {
		return err
	}

	err = field.readDataLength(ch)
	if err != nil {
		return err
	}

	return field.readDataData(ch)
}

type BinaryField struct{ fieldDataLength }
type BoundaryField struct{ fieldDataLength }
type CharField struct{ fieldDataLength }
type DateNField struct{ fieldDataLength }
type DateTimeNField struct{ fieldDataLength }
type FltNField struct{ fieldDataLength }
type IntNField struct{ fieldDataLength }
type UintNField struct{ fieldDataLength }
type LongBinaryField struct{ fieldDataLength }
type LongCharField struct{ fieldDataLength }
type MoneyNField struct{ fieldDataLength }
type SensitivityField struct{ fieldDataLength }
type TimeNField struct{ fieldDataLength }
type VarBinaryField struct{ fieldDataLength }
type VarCharField struct{ fieldDataLength }

type fieldDataLengthScale struct {
	fieldDataBase
	fieldDataScale
}

func (field *fieldDataLengthScale) readFormat(ch *channel) error {
	err := field.readFormatLength(ch)
	if err != nil {
		return err
	}
	return field.readFormatScale(ch)
}

func (field *fieldDataLengthScale) readData(ch *channel) error {
	err := field.readDataStatus(ch)
	if err != nil {
		return err
	}

	err = field.readDataLength(ch)
	if err != nil {
		return err
	}

	return field.readDataData(ch)
}

type BigDateTimeNField struct{ fieldDataLengthScale }
type BigTimeNField struct{ fieldDataLengthScale }

type fieldDataLengthPrecisionScale struct {
	fieldDataBase
	fieldDataPrecision
	fieldDataScale
}

func (field *fieldDataLengthPrecisionScale) readFormat(ch *channel) error {
	err := field.readFormatLength(ch)
	if err != nil {
		return err
	}

	err = field.readFormatPrecision(ch)
	if err != nil {
		return err
	}

	return field.readFormatScale(ch)
}

func (field *fieldDataLengthPrecisionScale) readData(ch *channel) error {
	err := field.readDataStatus(ch)
	if err != nil {
		return err
	}

	err = field.readDataLength(ch)
	if err != nil {
		return err
	}

	return field.readDataData(ch)
}

type DecNField struct{ fieldDataLengthPrecisionScale }
type NumNField struct{ fieldDataLengthPrecisionScale }

type fieldDataBlobType struct {
	fieldDataBase
	blobType          BlobType
	serializationType BlobSerializationType
	classID           string
	subClassID        string
	locator           string
	date              []byte
}

func (field fieldDataBlobType) copyDateTo(other *fieldDataBlobType) error {
	other.date = make([]byte, len(field.date))
	copy(other.date, field.date)
	return nil
}

func (field *fieldDataBlobType) readFormat(ch *channel) error {
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

func (field *fieldDataBlobType) readData(ch *channel) error {
	err := field.readDataStatus(ch)
	if err != nil {
		return err
	}

	serialization, err := ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read serialization type: %w", err)
	}

	switch serialization {
	case 0:
		switch field.blobType {
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
		if field.blobType != TDS_BLOB_UNICHAR {
			return fmt.Errorf("invalid blob (%s) and serialization (%d) type combination",
				field.blobType, serialization)
		}
		field.serializationType = UnicharUTF8
	case 2:
		if field.blobType != TDS_BLOB_UNICHAR {
			return fmt.Errorf("invalid blob (%s) and serialization (%d) type combination",
				field.blobType, serialization)
		}
		field.serializationType = UnicharSCSU
	default:
		return fmt.Errorf("unhandled serialization type %d", serialization)
	}

	switch field.blobType {
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

		var highBit uint32 = 0x80000000

		// extract high bit:
		// 0 -> last data set
		// 1 -> another data set follows
		highBitSet := dataLen&highBit == highBit
		dataLen = dataLen &^ highBit

		// if high bit is set and dataLen is zero no data array follows,
		// instead read the next data length immediately
		if highBitSet && dataLen == 0 {
			break
		}

		data, err := ch.Bytes(int(dataLen))
		if err != nil {
			return fmt.Errorf("failed to read data array: %w", err)
		}

		// TODO this could be inefficient
		field.data = append(field.data, data...)
	}

	return nil
}

type BlobField struct{ fieldDataBlobType }

type fieldDataTxtPtr struct {
	fieldDataBase

	name      string
	txtPtr    []byte
	timeStamp []byte
}

func (field *fieldDataTxtPtr) readFormat(ch *channel) error {
	length, err := ch.Uint32()
	if err != nil {
		return fmt.Errorf("failed to read length: %w", err)
	}
	field.length = int(length)

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

func (field *fieldDataTxtPtr) readData(ch *channel) error {
	err := field.readDataStatus(ch)
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

type ImageField struct{ fieldDataTxtPtr }
type TextField struct{ fieldDataTxtPtr }
type UniTextField struct{ fieldDataTxtPtr }
type XMLField struct{ fieldDataTxtPtr }

// Lookup

func LookupFieldData(dataType DataType) (FieldData, error) {
	switch dataType {
	case TDS_BIGDATETIMEN:
		v := &BigDateTimeNField{}
		v.dataType = dataType
		return v, nil
	case TDS_BIGTIMEN:
		v := &BigTimeNField{}
		v.dataType = dataType
		return v, nil
	case TDS_BIT:
		v := &BitField{}
		v.dataType = dataType
		v.length = 1
		return v, nil
	case TDS_DATETIME:
		v := &DateTimeField{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_DATE:
		v := &DateField{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_SHORTDATE:
		v := &ShortDateField{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_FLT4:
		v := &Flt4Field{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_FLT8:
		v := &Flt8Field{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_INT1:
		v := &Int1Field{}
		v.dataType = dataType
		v.length = 1
		return v, nil
	case TDS_INT2:
		v := &Int2Field{}
		v.dataType = dataType
		v.length = 2
		return v, nil
	case TDS_INT4:
		v := &Int4Field{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_INT8:
		v := &Int8Field{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_INTERVAL:
		v := &IntervalField{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_SINT1:
		v := &Sint1Field{}
		v.dataType = dataType
		v.length = 1
		return v, nil
	case TDS_UINT2:
		v := &Uint2Field{}
		v.dataType = dataType
		v.length = 2
		return v, nil
	case TDS_UINT4:
		v := &Uint4Field{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_UINT8:
		v := &Uint8Field{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_MONEY:
		v := &MoneyField{}
		v.dataType = dataType
		v.length = 8
		return v, nil
	case TDS_SHORTMONEY:
		v := &ShortMoneyField{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_TIME:
		v := &TimeField{}
		v.dataType = dataType
		v.length = 4
		return v, nil
	case TDS_BINARY:
		v := &BinaryField{}
		v.dataType = dataType
		return v, nil
	case TDS_BOUNDARY:
		v := &BoundaryField{}
		v.dataType = dataType
		return v, nil
	case TDS_CHAR:
		v := &CharField{}
		v.dataType = dataType
		return v, nil
	case TDS_DATEN:
		v := &DateNField{}
		v.dataType = dataType
		return v, nil
	case TDS_DATETIMEN:
		v := &DateTimeNField{}
		v.dataType = dataType
		return v, nil
	case TDS_FLTN:
		v := &FltNField{}
		v.dataType = dataType
		return v, nil
	case TDS_INTN:
		v := &IntNField{}
		v.dataType = dataType
		return v, nil
	case TDS_UINTN:
		v := &UintNField{}
		v.dataType = dataType
		return v, nil
	case TDS_LONGBINARY:
		v := &LongBinaryField{}
		v.dataType = dataType
		v.lengthBytes = 4
		return v, nil
	case TDS_LONGCHAR:
		v := &LongCharField{}
		v.dataType = dataType
		return v, nil
	case TDS_MONEYN:
		v := &MoneyNField{}
		v.dataType = dataType
		return v, nil
	case TDS_SENSITIVITY:
		v := &SensitivityField{}
		v.dataType = dataType
		return v, nil
	case TDS_TIMEN:
		v := &TimeNField{}
		v.dataType = dataType
		return v, nil
	case TDS_VARBINARY:
		v := &VarBinaryField{}
		v.dataType = dataType
		return v, nil
	case TDS_VARCHAR:
		v := &VarCharField{}
		v.dataType = dataType
		return v, nil
	case TDS_DECN:
		v := &DecNField{}
		v.dataType = dataType
		return v, nil
	case TDS_NUMN:
		v := &NumNField{}
		v.dataType = dataType
		return v, nil
	case TDS_BLOB:
		v := &BlobField{}
		v.dataType = dataType
		return v, nil
	case TDS_IMAGE:
		v := &ImageField{}
		v.dataType = dataType
		return v, nil
	case TDS_TEXT:
		v := &TextField{}
		v.dataType = dataType
		return v, nil
	case TDS_UNITEXT:
		v := &UniTextField{}
		v.dataType = dataType
		return v, nil
	case TDS_XML:
		v := &XMLField{}
		v.dataType = dataType
		return v, nil
	default:
		return nil, fmt.Errorf("unhandled datatype '%s'", dataType)
	}

	return nil, nil
}

func CopyFieldData(fieldData FieldData) (FieldData, error) {
	ret := *fieldData

	if target, ok := ret.(fieldDataBase); ok {
		src := fieldData.(fieldDataBase)
		if err := src.copyDataTo(target); err != nil {
			return fmt.Errorf("failed to copy data to target: %w", err)
		}
	}

	if target, ok := ret.(fieldDataBlobType); ok {
		src := fieldData.(fieldDataBlobType)
		if err := src.copyDateTo(target); err != nil {
			return fmt.Errorrf("failed to copy date to target: %w", err)
		}
	}

	if target, ok := ret.(fieldDataTxtPtr); ok {
		src := fieldData.(fieldDataTxtPtr)
		if err != src.copyTxtPtrTo(target); err != nil {
			return fmt.Errorf("failed to copy txt ptr to target: %w", err)
		}
	}

	return ret, nil
}
