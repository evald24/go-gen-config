package helpers

import "fmt"

func Contains(arr []string, value string) bool {
	for index := range arr {
		if arr[index] == value {
			return true
		}
	}
	return false
}

func GetString(m map[string]interface{}, k string) string {
	if _, ok := m[k]; ok {
		return fmt.Sprintf("%v", m[k])
	}
	return ""
}

func GetEnum(m map[string]interface{}) []string {
	var result []string
	if _, ok := m["enum"]; ok {
		for _, v := range m["enum"].([]interface{}) {
			result = append(result, fmt.Sprintf("%v", v))
		}
	}
	return result
}
