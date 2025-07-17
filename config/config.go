package config

import (
	"errors"
	"fmt"
	"os"
)

func GetEnvVariable(key string) (string, error) {
	value, ok := os.LookupEnv(key)

	if !ok {
		fmt.Printf("Variável de ambiente %s não encontrada.\n", key)
		return "", errors.New("environment variable not found")
	}

	return value, nil
}
