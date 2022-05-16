package convertions

import "encoding/binary"

func Bytes2Uint16(buf []byte) []uint16 {
	data := make([]uint16, 0, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		data = append(data, binary.BigEndian.Uint16(buf[i*2:]))
	}
	return data
}

func Uint162Bytes(value ...uint16) []byte {
	data := make([]byte, 2*len(value))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	return data
}
