package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
type Config struct {
	ServerPort         string   `mapstructure:"PORT"`
	MongoURI           string   `mapstructure:"MONGO_URI"`
	DBName             string   `mapstructure:"DB_NAME"`
	JWTSecretKey       string   `mapstructure:"JWT_SECRET_KEY"`
	JWTExpirationHours int      `mapstructure:"JWT_EXPIRATION_HOURS"`
	EnableCache        bool     `mapstructure:"ENABLE_CACHE"`
	RedisAddr          string   `mapstructure:"REDIS_ADDR"`
	RedisPassword      string   `mapstructure:"REDIS_PASSWORD"`
	LogLevel           string   `mapstructure:"LOG_LEVEL"`
	LogFormat          string   `mapstructure:"LOG_FORMAT"`
	CookieDomains      []string `mapstructure:"COOKIE_DOMAINS"`
	SecureCookie       bool     `mapstructure:"SECURE_COOKIE"`
	AllowedOrigins     []string `mapstructure:"ALLOWED_ORIGINS"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// Bind expected keys so values from environment variables are available
	// consistently when loading into the Config struct.
	keys := []string{
		"PORT",
		"MONGO_URI",
		"DB_NAME",
		"JWT_SECRET_KEY",
		"JWT_EXPIRATION_HOURS",
		"ENABLE_CACHE",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"COOKIE_DOMAINS",
		"SECURE_COOKIE",
		"ALLOWED_ORIGINS",
	}
	for _, key := range keys {
		if bindErr := viper.BindEnv(key); bindErr != nil {
			return config, bindErr
		}
	}

	// Set default values
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_NAME", "much_todo_db")
	viper.SetDefault("ENABLE_CACHE", false)
	viper.SetDefault("JWT_EXPIRATION_HOURS", 72)
	viper.SetDefault("COOKIE_DOMAINS", []string{"localhost"})
	viper.SetDefault("SECURE_COOKIE", false)
	viper.SetDefault("ALLOWED_ORIGINS", []string{"http://localhost:5173"})

	err = viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundErr) && !strings.Contains(err.Error(), "Config File \".env\" Not Found") {
			return
		}
		err = nil
	}

	config = Config{
		ServerPort:         viper.GetString("PORT"),
		MongoURI:           viper.GetString("MONGO_URI"),
		DBName:             viper.GetString("DB_NAME"),
		JWTSecretKey:       viper.GetString("JWT_SECRET_KEY"),
		JWTExpirationHours: viper.GetInt("JWT_EXPIRATION_HOURS"),
		EnableCache:        viper.GetBool("ENABLE_CACHE"),
		RedisAddr:          viper.GetString("REDIS_ADDR"),
		RedisPassword:      viper.GetString("REDIS_PASSWORD"),
		LogLevel:           viper.GetString("LOG_LEVEL"),
		LogFormat:          viper.GetString("LOG_FORMAT"),
		SecureCookie:       viper.GetBool("SECURE_COOKIE"),
		AllowedOrigins:     viper.GetStringSlice("ALLOWED_ORIGINS"),
		CookieDomains:      viper.GetStringSlice("COOKIE_DOMAINS"),
	}

	// Manually handle comma-separated strings for slices if viper didn't split them
	if allowedOrigins := viper.GetString("ALLOWED_ORIGINS"); allowedOrigins != "" {
		parts := strings.Split(allowedOrigins, ",")
		var cleaned []string
		for _, p := range parts {
			// Trim spaces and quotes
			trimmed := strings.TrimSpace(p)
			trimmed = strings.Trim(trimmed, "\"'")
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		config.AllowedOrigins = cleaned
	}

	if cookieDomains := viper.GetString("COOKIE_DOMAINS"); cookieDomains != "" {
		parts := strings.Split(cookieDomains, ",")
		var cleaned []string
		for _, p := range parts {
			// Trim spaces and quotes
			trimmed := strings.TrimSpace(p)
			trimmed = strings.Trim(trimmed, "\"'")
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		config.CookieDomains = cleaned
	}

	return
}
