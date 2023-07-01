package utils

import "encoding/binary"

func IntSliceToByteSlice(intSlice []int) []byte {
	byteSlice := make([]byte, len(intSlice))
	for i, val := range intSlice {
		byteSlice[i] = byte(val)
	}
	return byteSlice
}

func Uint16SliceToByteSlice(intSlice []uint16) []byte {
	byteSlice := make([]byte, len(intSlice))
	for i, val := range intSlice {
		byteSlice[i] = byte(val)
	}
	return byteSlice
}

func ByteSliceToIntSlice(byteSlice []byte) []int {
	intSlice := make([]int, len(byteSlice))
	for i, val := range byteSlice {
		intSlice[i] = int(val)
	}
	return intSlice
}

func ByteSliceToUint16Slice(byteSlice []byte) []uint16 {
	intSlice := make([]uint16, len(byteSlice))
	for i, val := range byteSlice {
		intSlice[i] = uint16(val)
	}
	return intSlice
}

func IntToBytes(n int) []byte {
	bytes := make([]byte, 4) // Assuming int size of 4 bytes (32 bits)
	binary.BigEndian.PutUint32(bytes, uint32(n))
	return bytes
}

func Uint16ToBytes(value uint16) []byte {
	bytes := make([]byte, 2)
	bytes[0] = byte(value >> 8)   // Most significant byte
	bytes[1] = byte(value & 0xFF) // Least significant byte
	return bytes
}
