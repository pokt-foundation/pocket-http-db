package router

import (
	"fmt"
	"net/url"
	"strconv"
)

type queryParamDetail struct {
	paramType paramType
	required  bool
}

type paramType string

const (
	paramTypeString paramType = "string"
	paramTypeBool   paramType = "bool"
	paramTypeInt    paramType = "int"
)

func (p paramType) isValid() bool {
	switch p {
	case paramTypeString, paramTypeBool, paramTypeInt:
		return true
	default:
		return false
	}
}

type chainQueryParams struct {
	IncludeInactive string
}

var chainParams = chainQueryParams{
	IncludeInactive: "include_inactive",
}

// paramProcessor validates and processes request query params based on provided specifications.
// It returns a map of processed parameters and any encountered error. Type conversion for the returned
// type should be handled in the route handler method as param types may be string, bool or int.
func paramProcessor(params url.Values, expectedParams map[string]queryParamDetail) (map[string]string, error) {
	result := make(map[string]string)

	for param, detail := range expectedParams {
		if !detail.paramType.isValid() {
			return nil, fmt.Errorf("invalid param type %s", detail.paramType)
		}

		val, ok := params[param]
		if !ok {
			if !detail.required {
				continue
			} else {
				return nil, fmt.Errorf("required parameter %s not found", param)
			}
		}

		result[param] = val[0]
	}

	return result, nil
}

func parseStringParam(paramMap map[string]string, key string) (string, error) {
	val, ok := paramMap[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in paramMap", key)
	}

	return val, nil
}

func parseBoolParam(paramMap map[string]string, key string) (bool, error) {
	val, ok := paramMap[key]
	if !ok {
		return false, fmt.Errorf("key %s not found in paramMap", key)
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("parseBool failed for key %s: %w", key, err)
	}

	return boolVal, nil
}

func parseIntParam(paramMap map[string]string, key string) (int, error) {
	val, ok := paramMap[key]
	if !ok {
		return 0, fmt.Errorf("key %s not found in paramMap", key)
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("parseInt failed for key %s: %w", key, err)
	}

	return intVal, nil
}
