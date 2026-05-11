package utils

// RemoveDuplicates 去重
func RemoveDuplicates(input []string) []string {
	uniqueMap := make(map[string]struct{})
	var result []string

	for _, str := range input {
		if _, exists := uniqueMap[str]; !exists {
			uniqueMap[str] = struct{}{}
			result = append(result, str)
		}
	}

	return result
}

func InSlice(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
