package array

func ArraySearchString(slice []string, value string) int {
	for p, v := range slice {
		if (v == value) {
			return p
		}
	}
	return -1
}
