package protonostale

func bytesChunkEvery(b []byte, n int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(b); i += n {
		end := i + n
		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(b) {
			end = len(b)
		}
		chunks = append(chunks, b[i:end])
	}
	return chunks
}
