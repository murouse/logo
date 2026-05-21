package logo

import (
	"fmt"
)

// Config объединяет все параметры кастомизации логгера.
type Config struct {
	Level       Level  // Уровень логирования (строка: debug, info, warn, error)
	Format      Format // Формат вывода логов (строка: json, console)
	WithCaller  bool   // Флаг добавления места вызова в лог (file.go:line)
	CallerSkip  int
	ServiceName *string // Имя сервиса, сквозным образом добавляемое во все записи
}

// Default возвращает базовый неизменяемый пресет настроек для локальной разработки.
func Default() *Config {
	return &Config{
		Level:       LevelDebug,
		Format:      FormatJSON,
		WithCaller:  true,
		CallerSkip:  1,
		ServiceName: nil,
	}
}

// Apply последовательно накатывает функциональные опции на текущую структуру конфигурации.
func (c *Config) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return fmt.Errorf("apply option: %w", err)
		}
	}

	return nil
}

// DefaultWith создает дефолтный конфиг и сразу модифицирует его переданными опциями.
func DefaultWith(opts ...Option) (*Config, error) {
	cfg := Default()
	if err := cfg.Apply(opts...); err != nil {
		return nil, fmt.Errorf("apply options: %w", err)
	}
	return cfg, nil
}

// Option инкапсулирует замыкание для безопасной настройки полей Config с валидацией "на лету".
type Option func(*Config) error

// WithLevel проверяет уровень по белому списку и прошивает его в конфиг.
func WithLevel(level Level) Option {
	return func(config *Config) error {
		_, ok := levelMap[level]
		if !ok {
			return fmt.Errorf("invalid level: %s", level)
		}
		config.Level = level
		return nil
	}
}

// WithFormat проверяет формат (json/console) и прошивает его в конфиг.
func WithFormat(format Format) Option {
	return func(config *Config) error {
		_, ok := formatMap[format]
		if !ok {
			return fmt.Errorf("invalid format: %s", format)
		}
		config.Format = format
		return nil
	}
}

// WithServiceName задает глобальный идентификатор сервиса.
func WithServiceName(serviceName string) Option {
	return func(config *Config) error {
		config.ServiceName = &serviceName
		return nil
	}
}

// WithCaller управляет отображением метаданных исходного кода в лог-линии.
func WithCaller(enabled bool) Option {
	return func(config *Config) error {
		config.WithCaller = enabled
		return nil
	}
}

func WithCallerSkip(skip int) Option {
	return func(cfg *Config) error {
		cfg.CallerSkip = skip
		return nil
	}
}
