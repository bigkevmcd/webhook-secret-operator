package webhooksecret

func containsString(c []string, s string) bool {
	for _, item := range c {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(c []string, s string) []string {
	result := []string{}
	for _, item := range c {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return result
}
