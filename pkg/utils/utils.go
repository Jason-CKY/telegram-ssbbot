package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func FloatPtr(num float64) *float64 {
	return &num
}

func IsUsernameAllowed(username string) bool {
	for _, allowed_username := range WhitelistedUsernames {
		if strings.Compare(username, allowed_username) == 0 {
			return true
		}
	}
	return false
}

func LookupEnvStringArray(key string) []string {
	envVariable, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Errorf("%v environment variable not set", key))
	}
	return strings.Split(envVariable, ",")
}

func LookupEnvString(key string) string {
	envVariable, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Errorf("%v environment variable not set", key))
	}
	return envVariable
}

func LookupEnvInt(key string) int {
	envVariable, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Errorf("%v environment variable not set", key))
	}
	num, err := strconv.Atoi(envVariable)
	if err != nil {
		panic(err.Error())
	}
	return num
}
