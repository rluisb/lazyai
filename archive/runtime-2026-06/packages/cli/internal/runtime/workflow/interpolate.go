package workflow

import (
	"fmt"
	"regexp"
	"strings"
)

// Interpolate replaces {VAR} placeholders with values from context
func Interpolate(template string, ctx map[string]string) string {
	result := template
	for key, value := range ctx {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// EvaluateTernary parses and evaluates ${COND ? 'then' : 'else'}
func EvaluateTernary(expr string, ctx map[string]string) (string, error) {
	// Pattern: ${VAR == 'value' ? 'then' : 'else'}
	re := regexp.MustCompile(`\$\{(\w+)\s*==\s*'([^']*)'\s*\?\s*'([^']*)'\s*:\s*'([^']*)'\}`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		// Not a ternary expression, return as-is
		return expr, nil
	}

	varName := matches[1]
	expectedValue := matches[2]
	thenValue := matches[3]
	elseValue := matches[4]

	actualValue, ok := ctx[varName]
	if !ok {
		actualValue = ""
	}

	if actualValue == expectedValue {
		return thenValue, nil
	}
	return elseValue, nil
}

// InterpolateWithTernary replaces {VAR} and evaluates ternary expressions
func InterpolateWithTernary(template string, ctx map[string]string) (string, error) {
	// First evaluate ternary expressions
	result, err := EvaluateTernary(template, ctx)
	if err != nil {
		return "", fmt.Errorf("evaluate ternary: %w", err)
	}

	// Then replace simple placeholders
	result = Interpolate(result, ctx)

	return result, nil
}
