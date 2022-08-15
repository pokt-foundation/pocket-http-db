package environment

import (
	"os"
	"strconv"
	"strings"

	// autoload env vars
	_ "github.com/joho/godotenv/autoload"
)

// GetString gets the environment var as a string
func GetString(varName string, defaultValue string) string {
	val, _ := os.LookupEnv(varName)
	if val == "" {
		return defaultValue
	}

	return val
}

// GetInt64 gets the env var as an int
func GetInt64(varName string, defaultValue int64) int64 {
	val, ok := os.LookupEnv(varName)
	if !ok {
		return defaultValue
	}

	iVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultValue
	}

	return iVal
}

// GetStringMap gets the environment var as a string map
func GetStringMap(varName, defaultValue, separator string) map[string]bool {
	rawString := GetString(varName, defaultValue)
	stringSlice := strings.Split(rawString, separator)

	stringMap := make(map[string]bool, len(stringSlice))

	for _, singleString := range stringSlice {
		stringMap[singleString] = true
	}

	return stringMap
}
