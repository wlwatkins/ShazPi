package utils

func IntSliceToByteSlice(intSlice []int) []byte {
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
