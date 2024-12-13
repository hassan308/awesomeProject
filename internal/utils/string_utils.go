package utils

// GetStringValue konverterar ett interface{} till en sträng
func GetStringValue(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// GetStringValueWithDefault konverterar ett interface{} till en sträng med ett default-värde
func GetStringValueWithDefault(value interface{}, defaultValue string) string {
	if str := GetStringValue(value); str != "" {
		return str
	}
	return defaultValue
}

// GetStringValueFromMap hämtar ett strängvärde från en map med ett default-värde
func GetStringValueFromMap(m map[string]interface{}, key string, defaultValue string) string {
	if value, exists := m[key]; exists {
		return GetStringValueWithDefault(value, defaultValue)
	}
	return defaultValue
} 