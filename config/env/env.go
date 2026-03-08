package env
package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func String(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}

func Int(key string, fallback int) (int, error) {
	value := String(key, "")
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %q", key, value)
	}

	return parsed, nil
}

func Duration(key string, fallback time.Duration) (time.Duration, error) {
	value := String(key, "")
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %q", key, value)
	}

	return parsed, nil
}

func Bool(key string, fallback bool) (bool, error) {
	value := String(key, "")
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid bool for %s: %q", key, value)
	}

	return parsed, nil
}
