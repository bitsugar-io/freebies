package rules

import (
	"strings"
)

// Evaluate evaluates a rule condition against generic event data
func Evaluate(
	condition map[string]interface{},
	eventData map[string]interface{},
) bool {
	if condition == nil {
		return false
	}

	for key, value := range condition {
		// Get the actual value from event data
		var actualValue interface{}
		var exists bool
		actualValue, exists = eventData[key]
		if !exists {
			// Key not in event data, condition fails
			return false
		}

		// Check if it's a comparison operator
		var compOp map[string]interface{}
		var ok bool
		compOp, ok = value.(map[string]interface{})
		if ok {
			if !evaluateComparison(actualValue, compOp) {
				return false
			}
		} else {
			// Direct value comparison
			if !directCompare(actualValue, value) {
				return false
			}
		}
	}

	return true
}

// evaluateComparison evaluates comparison operators (gte, gt, lte, lt, eq)
func evaluateComparison(
	actual interface{},
	compOp map[string]interface{},
) bool {
	for op, expectedVal := range compOp {
		switch op {
		case "gte":
			if !compareGTE(actual, expectedVal) {
				return false
			}
		case "gt":
			if !compareGT(actual, expectedVal) {
				return false
			}
		case "lte":
			if !compareLTE(actual, expectedVal) {
				return false
			}
		case "lt":
			if !compareLT(actual, expectedVal) {
				return false
			}
		case "eq":
			if !compareEQ(actual, expectedVal) {
				return false
			}
		case "contains":
			if !compareContains(actual, expectedVal) {
				return false
			}
		case "in":
			if !compareIn(actual, expectedVal) {
				return false
			}
		default:
			// Unknown operator, fail safe
			return false
		}
	}
	return true
}

// directCompare compares values directly (for booleans, strings, etc.)
func directCompare(
	actual interface{},
	expected interface{},
) bool {
	// Type-specific comparisons
	switch actualTyped := actual.(type) {
	case bool:
		var expectedBool bool
		var ok bool
		expectedBool, ok = expected.(bool)
		if ok {
			return actualTyped == expectedBool
		}
	case string:
		var expectedStr string
		var ok bool
		expectedStr, ok = expected.(string)
		if ok {
			return actualTyped == expectedStr
		}
	case int, int64, float64:
		return compareEQ(actual, expected)
	}
	return false
}

// Numeric comparison helpers
func compareGTE(
	actual interface{},
	expected interface{},
) bool {
	var actualNum float64
	var ok1 bool
	actualNum, ok1 = toFloat64(actual)
	var expectedNum float64
	var ok2 bool
	expectedNum, ok2 = toFloat64(expected)
	if !ok1 || !ok2 {
		return false
	}
	return actualNum >= expectedNum
}

func compareGT(
	actual interface{},
	expected interface{},
) bool {
	var actualNum float64
	var ok1 bool
	actualNum, ok1 = toFloat64(actual)
	var expectedNum float64
	var ok2 bool
	expectedNum, ok2 = toFloat64(expected)
	if !ok1 || !ok2 {
		return false
	}
	return actualNum > expectedNum
}

func compareLTE(
	actual interface{},
	expected interface{},
) bool {
	var actualNum float64
	var ok1 bool
	actualNum, ok1 = toFloat64(actual)
	var expectedNum float64
	var ok2 bool
	expectedNum, ok2 = toFloat64(expected)
	if !ok1 || !ok2 {
		return false
	}
	return actualNum <= expectedNum
}

func compareLT(
	actual interface{},
	expected interface{},
) bool {
	var actualNum float64
	var ok1 bool
	actualNum, ok1 = toFloat64(actual)
	var expectedNum float64
	var ok2 bool
	expectedNum, ok2 = toFloat64(expected)
	if !ok1 || !ok2 {
		return false
	}
	return actualNum < expectedNum
}

func compareEQ(
	actual interface{},
	expected interface{},
) bool {
	// Try numeric comparison first
	var actualNum float64
	var ok1 bool
	actualNum, ok1 = toFloat64(actual)
	var expectedNum float64
	var ok2 bool
	expectedNum, ok2 = toFloat64(expected)
	if ok1 && ok2 {
		return actualNum == expectedNum
	}

	// Fall back to direct equality
	return actual == expected
}

// compareContains checks if a string contains a substring
func compareContains(
	actual interface{},
	expected interface{},
) bool {
	var actualStr string
	var ok1 bool
	actualStr, ok1 = actual.(string)
	var expectedStr string
	var ok2 bool
	expectedStr, ok2 = expected.(string)
	if !ok1 || !ok2 {
		return false
	}
	return strings.Contains(strings.ToLower(actualStr), strings.ToLower(expectedStr))
}

// compareIn checks if actual value is in the expected array
func compareIn(
	actual interface{},
	expected interface{},
) bool {
	var expectedSlice []interface{}
	var ok bool
	expectedSlice, ok = expected.([]interface{})
	if !ok {
		return false
	}

	for _, item := range expectedSlice {
		if directCompare(actual, item) {
			return true
		}
	}
	return false
}

// toFloat64 converts interface{} to float64 for numeric comparisons
func toFloat64(
	val interface{},
) (float64, bool) {
	switch v := val.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// EvaluateNested evaluates nested conditions (for complex event data)
// Example: {"players": [{"name": "Ohtani", "hr": {"gte": 1}}]}
func EvaluateNested(
	condition map[string]interface{},
	eventData map[string]interface{},
) bool {
	for key, condValue := range condition {
		var actualValue interface{}
		var exists bool
		actualValue, exists = eventData[key]
		if !exists {
			return false
		}

		// Handle array matching (e.g., players array)
		var condMap map[string]interface{}
		var ok bool
		condMap, ok = condValue.(map[string]interface{})
		if ok {
			var actualSlice []interface{}
			actualSlice, ok = actualValue.([]interface{})
			if ok {
				// Check if any item in the slice matches the condition
				matched := false
				for _, item := range actualSlice {
					var itemMap map[string]interface{}
					itemMap, ok = item.(map[string]interface{})
					if ok {
						if Evaluate(condMap, itemMap) {
							matched = true
							break
						}
					}
				}
				if !matched {
					return false
				}
			} else {
				var actualMap map[string]interface{}
				actualMap, ok = actualValue.(map[string]interface{})
				if ok {
					// Nested object comparison
					if !Evaluate(condMap, actualMap) {
						return false
					}
				} else {
					// Single value with comparison ops
					if !evaluateComparison(actualValue, condMap) {
						return false
					}
				}
			}
		} else {
			// Direct comparison
			if !directCompare(actualValue, condValue) {
				return false
			}
		}
	}

	return true
}
