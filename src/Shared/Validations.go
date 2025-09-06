package Shared

import (
	"os"
	"strings"
)

func ValidateEnvironmentVariable(name string) string {
	result := os.Getenv(strings.ToUpper(name))
	if result != "" {
		return result
	} else {
		panic("Environment Variable " + name + " not available")

	}

}
