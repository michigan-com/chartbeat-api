package main

func mapStringsToTrue(list []string) map[string]bool {
	result := make(map[string]bool, len(list))
	for _, item := range list {
		result[item] = true
	}
	return result
}
