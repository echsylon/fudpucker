package utils

import "encoding/binary"

func BoolToBytes(value bool) []byte {
	if value {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

func ByteToBytes(value byte) []byte {
	return []byte{value}
}

func Int64ToBytes(value int64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(value))
	return data
}

func StringToBytes(value string) []byte {
	return []byte(value)
}

func BytesToBool(data []byte) bool {
	if len(data) == 0 { // handles nil too
		return false
	} else {
		return binary.BigEndian.Uint64(data) != uint64(0)
	}
}

func BytesToByte(data []byte) byte {
	if len(data) == 0 { // handles nil too
		return byte(0)
	} else {
		return byte(data[len(data)-1])
	}
}

func BytesToInt64(data []byte) int64 {
	if len(data) == 0 { // handles nil too
		return int64(0)
	} else {
		value := binary.BigEndian.Uint64(data)
		return int64(value)
	}
}

func BytesToString(data []byte) string {
	if len(data) == 0 { // handles nil too
		return ""
	} else {
		return string(data)
	}
}
