package util

import (
	"errors"
	"os"
)

func ValueGetter(vars ...string) (map[string]string, error) {
	result := make(map[string]string)
	for _, v := range vars {
		env := os.Getenv(v)
		if env == "" {
			return nil, errors.New("environment variable " + v + " is not set or empty")
		}
		result[v] = env
	}
	return result, nil
}
