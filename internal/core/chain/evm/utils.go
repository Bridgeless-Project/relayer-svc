package evm

func to32Bytes(data []byte) [32]byte {
	var arr [32]byte
	if len(data) > 32 {
		copy(arr[:], data[:32])
		return arr
	}
	copy(arr[:], data)

	return arr
}
