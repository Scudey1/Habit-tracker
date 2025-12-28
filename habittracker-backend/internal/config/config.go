package config

import "os"

type Config struct {
    JWTSecret string
}

func Load() *Config {
    return &Config{
        JWTSecret: GetEnv("JWT_SECRET", "habittracker-secret-key-change-this"),
    }
}


func GetEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}