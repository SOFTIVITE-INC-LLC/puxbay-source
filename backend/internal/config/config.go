package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Stripe   StripeConfig
	Google   GoogleConfig
	Xero     XeroConfig
	CORS     CORSConfig
	SMTP     SMTPConfig
	SMS      SMSConfig
	Fernet   FernetConfig
	Paystack PaystackConfig
}

type AppConfig struct {
	Env        string `mapstructure:"APP_ENV"` // e.g. "development", "staging", "production"
	Port       string `mapstructure:"APP_PORT"`
	Name       string `mapstructure:"APP_NAME"`
	RootDomain string `mapstructure:"ROOT_DOMAIN"`
	UseSecrets bool   `mapstructure:"USE_AWS_SECRETS"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"DB_HOST"`
	Port            string        `mapstructure:"DB_PORT"`
	User            string        `mapstructure:"DB_USER"`
	Password        string        `mapstructure:"DB_PASSWORD"`
	DBName          string        `mapstructure:"DB_NAME"`
	SSLMode         string        `mapstructure:"DB_SSLMODE"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

type RedisConfig struct {
	URL string `mapstructure:"REDIS_URL"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"JWT_SECRET"`
	AccessExpiry  time.Duration `mapstructure:"JWT_ACCESS_EXPIRY"`
	RefreshExpiry time.Duration `mapstructure:"JWT_REFRESH_EXPIRY"`
}

type StripeConfig struct {
	SecretKey      string `mapstructure:"STRIPE_SECRET_KEY"`
	WebhookSecret  string `mapstructure:"STRIPE_WEBHOOK_SECRET"`
	PublishableKey string `mapstructure:"STRIPE_PUBLISHABLE_KEY"`
}

type PaystackConfig struct {
	SecretKey   string `mapstructure:"PAYSTACK_SECRET_KEY"`
	CallbackURL string `mapstructure:"PAYSTACK_CALLBACK_URL"`
}

type GoogleConfig struct {
	ClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	ClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GeminiAPIKey string `mapstructure:"GEMINI_API_KEY"`
}

type XeroConfig struct {
	ClientID     string `mapstructure:"XERO_CLIENT_ID"`
	ClientSecret string `mapstructure:"XERO_CLIENT_SECRET"`
	RedirectURL  string `mapstructure:"XERO_REDIRECT_URL"`
}

type CORSConfig struct {
	AllowedOrigins []string
}

type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     string `mapstructure:"SMTP_PORT"`
	User     string `mapstructure:"SMTP_USER"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	From     string `mapstructure:"SMTP_FROM"`
}

type SMSConfig struct {
	APIKey   string `mapstructure:"SMS_API_KEY"`
	SenderID string `mapstructure:"SMS_SENDER_ID"`
}

type FernetConfig struct {
	Key string `mapstructure:"FERNET_KEY"`
}

// Load reads configuration from .env file and environment variables.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "5000")
	viper.SetDefault("APP_NAME", "puxbay")
	viper.SetDefault("ROOT_DOMAIN", "localhost:5000")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "puxbay_go")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379/0")
	viper.SetDefault("JWT_SECRET", "change-me-in-production")
	viper.SetDefault("JWT_ACCESS_EXPIRY", "400m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h") // 7 days
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:4200")
	viper.SetDefault("PAYSTACK_CALLBACK_URL", "http://localhost:4200/billing")
	viper.SetDefault("USE_AWS_SECRETS", false)

	// Read .env file (ignore error if file doesn't exist)
	_ = viper.ReadInConfig()

	if viper.GetBool("USE_AWS_SECRETS") {
		loadSecretsFromVault()
	}

	cfg := &Config{}

	// App
	cfg.App = AppConfig{
		Env:        viper.GetString("APP_ENV"),
		Port:       viper.GetString("APP_PORT"),
		Name:       viper.GetString("APP_NAME"),
		RootDomain: viper.GetString("ROOT_DOMAIN"),
	}

	// Database
	cfg.Database = DatabaseConfig{
		Host:            viper.GetString("DB_HOST"),
		Port:            viper.GetString("DB_PORT"),
		User:            viper.GetString("DB_USER"),
		Password:        viper.GetString("DB_PASSWORD"),
		DBName:          viper.GetString("DB_NAME"),
		SSLMode:         viper.GetString("DB_SSLMODE"),
		MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
	}

	// Redis
	cfg.Redis = RedisConfig{
		URL: viper.GetString("REDIS_URL"),
	}

	// JWT
	cfg.JWT = JWTConfig{
		Secret:        viper.GetString("JWT_SECRET"),
		AccessExpiry:  viper.GetDuration("JWT_ACCESS_EXPIRY"),
		RefreshExpiry: viper.GetDuration("JWT_REFRESH_EXPIRY"),
	}

	// Stripe
	cfg.Stripe = StripeConfig{
		SecretKey:      viper.GetString("STRIPE_SECRET_KEY"),
		WebhookSecret:  viper.GetString("STRIPE_WEBHOOK_SECRET"),
		PublishableKey: viper.GetString("STRIPE_PUBLISHABLE_KEY"),
	}

	// Paystack
	cfg.Paystack = PaystackConfig{
		SecretKey:   viper.GetString("PAYSTACK_SECRET_KEY"),
		CallbackURL: viper.GetString("PAYSTACK_CALLBACK_URL"),
	}

	// Google
	cfg.Google = GoogleConfig{
		ClientID:     viper.GetString("GOOGLE_CLIENT_ID"),
		ClientSecret: viper.GetString("GOOGLE_CLIENT_SECRET"),
		GeminiAPIKey: viper.GetString("GEMINI_API_KEY"),
	}

	// Xero
	cfg.Xero = XeroConfig{
		ClientID:     viper.GetString("XERO_CLIENT_ID"),
		ClientSecret: viper.GetString("XERO_CLIENT_SECRET"),
		RedirectURL:  viper.GetString("XERO_REDIRECT_URL"),
	}

	// CORS
	origins := viper.GetString("CORS_ALLOWED_ORIGINS")
	cfg.CORS = CORSConfig{
		AllowedOrigins: strings.Split(origins, ","),
	}

	// SMTP
	cfg.SMTP = SMTPConfig{
		Host:     viper.GetString("SMTP_HOST"),
		Port:     viper.GetString("SMTP_PORT"),
		User:     viper.GetString("SMTP_USER"),
		Password: viper.GetString("SMTP_PASSWORD"),
		From:     viper.GetString("SMTP_FROM"),
	}

	// SMS
	cfg.SMS = SMSConfig{
		APIKey:   viper.GetString("SMS_API_KEY"),
		SenderID: viper.GetString("SMS_SENDER_ID"),
	}

	// Fernet
	cfg.Fernet = FernetConfig{
		Key: viper.GetString("FERNET_KEY"),
	}

	return cfg, nil
}

// loadSecretsFromVault is a stub representing fetching secrets from AWS Secrets Manager or HashiCorp Vault.
// In a real implementation, this would use the AWS SDK to fetch a JSON string and parse it into viper configs.
func loadSecretsFromVault() {
	log.Println("Loading secrets from AWS Secrets Manager...")

	// Example mock secret JSON returned from AWS
	mockSecretJSON := `{
		"JWT_SECRET": "production-super-secret-key-loaded-from-aws",
		"DB_PASSWORD": "secure-production-db-password"
	}`

	var secrets map[string]interface{}
	if err := json.Unmarshal([]byte(mockSecretJSON), &secrets); err == nil {
		for key, val := range secrets {
			viper.Set(key, val)
		}
	}
}
