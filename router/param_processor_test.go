package router

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_paramProcessor(t *testing.T) {
	expectedParams := map[string]queryParamDetail{
		"include_inactive": {
			paramType: paramTypeBool,
			required:  false,
		},
		"age": {
			paramType: paramTypeInt,
			required:  true,
		},
		"name": {
			paramType: paramTypeString,
			required:  false,
		},
	}

	tests := []struct {
		name          string
		params        url.Values
		expected      map[string]string
		expectedError string
	}{
		{
			name: "Should process valid params correctly",
			params: url.Values{
				"include_inactive": []string{"true"},
				"age":              []string{"22"},
				"name":             []string{"John"},
			},
			expected: map[string]string{
				"include_inactive": "true",
				"age":              "22",
				"name":             "John",
			},
		},
		{
			name: "Should return error for missing required int param",
			params: url.Values{
				"include_inactive": []string{"true"},
				"name":             []string{"John"},
			},
			expectedError: "required parameter age not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := paramProcessor(test.params, expectedParams)

			if test.expectedError != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), test.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func Test_parseStringParam(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]string
		key            string
		expected       string
		expectingError bool
	}{
		{
			name: "Parses valid string",
			params: map[string]string{
				"name": "John",
			},
			key:      "name",
			expected: "John",
		},
		{
			name:           "Returns error when key is not found",
			params:         map[string]string{},
			key:            "name",
			expectingError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseStringParam(test.params, test.key)
			if test.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func Test_parseBoolParam(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]string
		key            string
		expected       bool
		expectingError bool
	}{
		{
			name: "Parses valid bool",
			params: map[string]string{
				"active": "true",
			},
			key:      "active",
			expected: true,
		},
		{
			name:           "Returns error when key is not found",
			params:         map[string]string{},
			key:            "active",
			expectingError: true,
		},
		{
			name: "Returns error for invalid bool",
			params: map[string]string{
				"active": "invalid_bool",
			},
			key:            "active",
			expectingError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseBoolParam(test.params, test.key)
			if test.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func Test_parseIntParam(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]string
		key            string
		expected       int
		expectingError bool
	}{
		{
			name: "Parses valid int",
			params: map[string]string{
				"age": "22",
			},
			key:      "age",
			expected: 22,
		},
		{
			name:           "Returns error when key is not found",
			params:         map[string]string{},
			key:            "age",
			expectingError: true,
		},
		{
			name: "Returns error for invalid int",
			params: map[string]string{
				"age": "invalid_int",
			},
			key:            "age",
			expectingError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseIntParam(test.params, test.key)
			if test.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func Test_paramProcessorAndParse(t *testing.T) {
	expectedParams := map[string]queryParamDetail{
		"include_inactive": {
			paramType: paramTypeBool,
			required:  false,
		},
		"age": {
			paramType: paramTypeInt,
			required:  true,
		},
		"name": {
			paramType: paramTypeString,
			required:  false,
		},
	}

	tests := []struct {
		name          string
		params        url.Values
		expected      map[string]interface{} // changed to interface{} to accommodate different types
		expectedError string
	}{
		{
			name: "Should process and parse valid bool, int, and string params correctly",
			params: url.Values{
				"include_inactive": []string{"true"},
				"age":              []string{"22"},
				"name":             []string{"John"},
			},
			expected: map[string]interface{}{
				"include_inactive": true,   // parsed as bool
				"age":              22,     // parsed as int
				"name":             "John", // parsed as string
			},
		},
		{
			name: "Should return error for missing required int param",
			params: url.Values{
				"include_inactive": []string{"true"},
				"name":             []string{"John"},
			},
			expectedError: "required parameter age not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := paramProcessor(test.params, expectedParams)
			if err != nil {
				assert.EqualError(t, err, test.expectedError)
				return
			}

			// If no processing error, we check the parsing
			for key, expectedValue := range test.expected {
				switch expectedParams[key].paramType {
				case paramTypeBool:
					actual, err := parseBoolParam(result, key)
					assert.NoError(t, err)
					assert.Equal(t, expectedValue, actual)
				case paramTypeInt:
					actual, err := parseIntParam(result, key)
					assert.NoError(t, err)
					assert.Equal(t, expectedValue, actual)
				case paramTypeString:
					actual, err := parseStringParam(result, key)
					assert.NoError(t, err)
					assert.Equal(t, expectedValue, actual)
				}
			}
		})
	}
}
