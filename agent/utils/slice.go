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
