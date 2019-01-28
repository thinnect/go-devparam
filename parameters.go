package deviceparameters

import "fmt"
import "errors"
import "encoding/hex"
import "encoding/binary"
import "bytes"
import "strconv"

type DeviceParameterType uint8

const (
	DP_TYPE_RAW    DeviceParameterType = 0x00
	DP_TYPE_UINT8  DeviceParameterType = 0x01
	DP_TYPE_UINT16 DeviceParameterType = 0x02
	DP_TYPE_UINT32 DeviceParameterType = 0x04
	DP_TYPE_UINT64 DeviceParameterType = 0x08

	DP_TYPE_STRING DeviceParameterType = 0x80
	DP_TYPE_INT8   DeviceParameterType = 0x81
	DP_TYPE_INT16  DeviceParameterType = 0x82
	DP_TYPE_INT32  DeviceParameterType = 0x84
	DP_TYPE_INT64  DeviceParameterType = 0x88

	DP_TYPE_NIL DeviceParameterType = 0xFF
)

var DeviceParameterTypeToString = map[DeviceParameterType]string{
	DP_TYPE_RAW:    "raw",
	DP_TYPE_UINT8:  "u8",
	DP_TYPE_UINT16: "u16",
	DP_TYPE_UINT32: "u32",
	DP_TYPE_UINT64: "u64",

	DP_TYPE_STRING: "str",
	DP_TYPE_INT8:   "i8",
	DP_TYPE_INT16:  "i16",
	DP_TYPE_INT32:  "i32",
	DP_TYPE_INT64:  "i64",

	DP_TYPE_NIL: "nil",
}

var DeviceParameterStringToType = map[string]DeviceParameterType{}

func init() {
	for k, v := range DeviceParameterTypeToString {
		DeviceParameterStringToType[v] = k
	}
}

func (dpt DeviceParameterType) String() string {
	return DeviceParameterTypeToString[dpt]
}

func ParseDeviceParameterType(name string) (DeviceParameterType, error) {
	v, ok := DeviceParameterStringToType[name]
	if ok {
		return v, nil
	}
	return DP_TYPE_NIL, errors.New(fmt.Sprintf("%s is not a valid parameter type!", name))
}

func ParseParameterValue(tpt DeviceParameterType, tval string) ([]byte, error) {
	var value []byte
	var err error
	var t interface{}

	switch tpt {
	case DP_TYPE_RAW:
		return hex.DecodeString(tval)
	case DP_TYPE_STRING:
		return []byte(tval), nil
	case DP_TYPE_NIL:
		return nil, nil
	case DP_TYPE_UINT8:
		if v, err := strconv.ParseUint(tval, 10, 8); err == nil {
			t = uint8(v)
		} else {
			return nil, err
		}
	case DP_TYPE_UINT16:
		if v, err := strconv.ParseUint(tval, 10, 16); err == nil {
			t = uint16(v)
		} else {
			return nil, err
		}
	case DP_TYPE_UINT32:
		if v, err := strconv.ParseUint(tval, 10, 32); err == nil {
			t = uint32(v)
		} else {
			return nil, err
		}
	case DP_TYPE_UINT64:
		if v, err := strconv.ParseUint(tval, 10, 64); err == nil {
			t = uint64(v)
		} else {
			return nil, err
		}
	case DP_TYPE_INT8:
		if v, err := strconv.ParseInt(tval, 10, 8); err == nil {
			t = int8(v)
		} else {
			return nil, err
		}
	case DP_TYPE_INT16:
		if v, err := strconv.ParseInt(tval, 10, 16); err == nil {
			t = int16(v)
		} else {
			return nil, err
		}
	case DP_TYPE_INT32:
		if v, err := strconv.ParseInt(tval, 10, 32); err == nil {
			t = int32(v)
		} else {
			return nil, err
		}
	case DP_TYPE_INT64:
		if v, err := strconv.ParseInt(tval, 10, 64); err == nil {
			t = int64(v)
		} else {
			return nil, err
		}
	}

	if t != nil {
		switch t := t.(type) {
		case uint8, uint16, uint32, uint64, int8, int16, int32, int64:
			buf := new(bytes.Buffer)
			if err := binary.Write(buf, binary.BigEndian, t); err != nil {
				return nil, err
			}
			value = buf.Bytes()
		}
	}

	return value, err
}

func ParameterValueString(dpt DeviceParameterType, value []byte) (string, error) {
	if dpt == DP_TYPE_NIL {
		return fmt.Sprintf("%X", value), nil
	} else if dpt == DP_TYPE_RAW {
		return fmt.Sprintf("%X", value), nil
	} else if dpt == DP_TYPE_STRING {
		return string(value), nil
	} else {
		s := fmt.Sprintf("%v", value)
		buf := bytes.NewBuffer(value)
		if dpt == DP_TYPE_UINT8 {
			var v uint8
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if dpt == DP_TYPE_UINT16 {
			var v uint16
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if dpt == DP_TYPE_UINT32 {
			var v uint32
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if dpt == DP_TYPE_UINT64 {
			var v uint64
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if dpt == DP_TYPE_INT8 {
			var v int8
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if dpt == DP_TYPE_INT16 {
			var v int16
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if dpt == DP_TYPE_INT32 {
			var v int32
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if dpt == DP_TYPE_INT64 {
			var v int64
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else {
			return s, errors.New("Unrecognized parameter type!")
		}
		return s, nil
	}
}
