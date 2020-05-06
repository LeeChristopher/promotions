package tools

func InString(str string, data []string) bool {
	if len(data) == 0 {
		return false
	}

	for i := range data {
		if data[i] == str {
			return true
		}
	}

	return false
}

func InUint64(subject uint64, data []uint64) bool {
	if len(data) == 0 {
		return false
	}

	for i := range data {
		if data[i] == subject {
			return true
		}
	}

	return false
}

func InUint32(subject uint32, data []uint32) bool {
	if len(data) == 0 {
		return false
	}

	for i := range data {
		if data[i] == subject {
			return true
		}
	}

	return false
}
