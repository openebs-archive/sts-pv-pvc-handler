package utils

import (
	"fmt"
	"os"
	"strings"
)

func EnvVarSlice(envVarName string) []string {
	provisionersEnvVar, exists := os.LookupEnv(envVarName)
	if !exists {
		panic(fmt.Sprintf("Environment Variable %v not found", envVarName))
	}
	provisioners := strings.Split(provisionersEnvVar, ",")
	return provisioners
}
