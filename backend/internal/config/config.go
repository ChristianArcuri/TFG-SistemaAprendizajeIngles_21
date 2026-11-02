// Package config carga y valida la configuración de la aplicación
// desde variables de entorno. Es el único lugar del backend que debe
// leer os.Getenv: el resto del código recibe un *Config ya validado.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config agrupa toda la configuración del backend.
type Config struct {
	App    AppConfig
	HTTP   HTTPConfig
	DB     DBConfig
	JWT    JWTConfig
	AI     AIConfig
	OpenAI OpenAIConfig
}

type AppConfig struct {
	Env  string // development | staging | production
	Name string
}

type HTTPConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	CORSOrigins     string // comma-separated; vacio = refleja Origin en dev
}

type DBConfig struct {
	URL            string
	MaxOpenConns   int32
	MaxIdleTime    time.Duration
	ConnectTimeout time.Duration
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

// AIConfig controla qué proveedor de IA usar (mock por default) y los
// límites operativos asociados (tamaño máximo del audio aceptado).
type AIConfig struct {
	Provider      string // "mock" | "openai"
	MaxAudioBytes int64
}

// OpenAIConfig agrupa credenciales y parametros del cliente OpenAI.
type OpenAIConfig struct {
	APIKey               string
	TranscriptionModel   string
	FeedbackModel        string
	TranscriptionTimeout time.Duration
	FeedbackTimeout      time.Duration
}

// Load lee las variables de entorno y devuelve un Config validado.
// Falla rápido si falta algún valor obligatorio.
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Name: getEnv("APP_NAME", "tfg-sai"),
		},
		HTTP: HTTPConfig{
			Port:            getEnv("HTTP_PORT", "8080"),
			ReadTimeout:     getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			CORSOrigins: getEnv("CORS_ALLOWED_ORIGINS",
				"http://localhost:5173,http://127.0.0.1:5173"),
		},
		DB: DBConfig{
			URL:            getEnv("DATABASE_URL", ""),
			MaxOpenConns:   int32(getEnvInt("DB_MAX_OPEN_CONNS", 20)),
			MaxIdleTime:    getEnvDuration("DB_MAX_IDLE_TIME", 5*time.Minute),
			ConnectTimeout: getEnvDuration("DB_CONNECT_TIMEOUT", 5*time.Second),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", ""),
			ExpiresIn: getEnvDuration("JWT_EXPIRES_IN", 24*time.Hour),
		},
		AI: AIConfig{
			Provider:      strings.ToLower(getEnv("AI_PROVIDER", "mock")),
			MaxAudioBytes: int64(getEnvInt("MAX_AUDIO_BYTES", 15*1024*1024)),
		},
		OpenAI: OpenAIConfig{
			APIKey:               getEnv("OPENAI_API_KEY", ""),
			TranscriptionModel:   getEnv("OPENAI_TRANSCRIPTION_MODEL", "whisper-1"),
			FeedbackModel:        getEnv("OPENAI_FEEDBACK_MODEL", "gpt-4o-mini"),
			TranscriptionTimeout: getEnvDuration("OPENAI_TRANSCRIPTION_TIMEOUT", 60*time.Second),
			FeedbackTimeout:      getEnvDuration("OPENAI_FEEDBACK_TIMEOUT", 30*time.Second),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	var errs []string

	if strings.TrimSpace(c.DB.URL) == "" {
		errs = append(errs, "DATABASE_URL is required")
	}
	if strings.TrimSpace(c.JWT.Secret) == "" {
		errs = append(errs, "JWT_SECRET is required")
	}
	if c.IsProduction() && len(c.JWT.Secret) < 32 {
		errs = append(errs, "JWT_SECRET must be at least 32 characters in production")
	}

	switch c.AI.Provider {
	case "mock", "openai":
	default:
		errs = append(errs, "AI_PROVIDER must be 'mock' or 'openai'")
	}
	if c.AI.Provider == "openai" && strings.TrimSpace(c.OpenAI.APIKey) == "" {
		errs = append(errs, "OPENAI_API_KEY is required when AI_PROVIDER=openai")
	}
	if c.AI.MaxAudioBytes <= 0 {
		errs = append(errs, "MAX_AUDIO_BYTES must be greater than zero")
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid configuration: %s", strings.Join(errs, "; "))
	}
	return nil
}

// IsProduction indica si el backend corre en modo producción.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "production")
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

// ErrMissingDatabaseURL se exporta por si algún paquete necesita
// distinguir este caso (no se usa todavía, pero queda preparado).
var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required")
