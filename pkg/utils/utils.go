package utils

import (
	"fmt"
	"os"
	"strings"
)

func EnvVarSlice(envVarName string) []string {
	envVar, exists := os.LookupEnv(envVarName)
	if !exists {
		panic(fmt.Sprintf("Environment Variable %v not found", envVarName))
	}
	slice := strings.Split(envVar, ",")
	return slice
}
