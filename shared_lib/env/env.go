package config

import (
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	logLevel   = "ENVIRONMENT"
	debugBuild = "LOCAL-DEV"
)

func WithDefaultString(key, defaultValue string) string {
	log.Trace("Env WithDefaultString")

	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func WithDefaultInt(key string, defaultValue string) int {
	log.Trace("Env WithDefaultInt")

	if value, exists := os.LookupEnv(key); exists {
		if result, err := strconv.Atoi(value); err == nil {
			return result
		} else {
			log.Errorf("could not load environment variable, expected integer, got type: %T\n", value)
			panic(err)
		}
	}
	defaultIntValue, err := strconv.Atoi(defaultValue)
	if err != nil {
		log.Errorf("could not convert default environment variable, expected integer, got type: %T\n", defaultIntValue)
	}
	return defaultIntValue
}

func WithDefaultInt64(key string, defaultValue string) int64 {
	log.Trace("Env WithDefaultInt64")

	if value, exists := os.LookupEnv(key); exists {
		if i64, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i64
		} else {
			log.Errorf("could not load environment variable, expected int64, got type: %T\n", value)
			panic(err)
		}
	}
	defaultInt64Value, err := strconv.ParseInt(defaultValue, 10, 64)
	if err != nil {
		log.Errorf("could not convert default environment variable, expected 64 bit integer, got type: %T\n", defaultInt64Value)
	}
	return defaultInt64Value
}

func IsLocalDev() bool {
	buildType := WithDefaultString(logLevel, debugBuild)
	return strings.ToUpper(buildType) == debugBuild
}
