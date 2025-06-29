package config

import "time"

// Config - основная конфигурация приложения
// Содержит все настройки для работы парсера резюме
type Config struct {
	API      APIConfig      `json:"api"`      // Настройки API
	Search   SearchConfig   `json:"search"`   // Параметры поиска
	Output   OutputConfig   `json:"output"`   // Настройки вывода
	Database DatabaseConfig `json:"database"` // Настройки базы данных
	LogFile  string         `json:"log_file"` // Файл логов
}

// APIConfig - конфигурация для работы с API hh.ru
type APIConfig struct {
	Token     string        `json:"token"`      // API токен hh.ru
	RateLimit time.Duration `json:"rate_limit"` // Ограничение скорости запросов
	UserAgent string        `json:"user_agent"` // User-Agent для запросов
	Timeout   time.Duration `json:"timeout"`    // Таймаут HTTP запросов
}

// SearchConfig - параметры поиска резюме
type SearchConfig struct {
	Keywords   []string `json:"keywords"`    // Ключевые слова для поиска
	City       string   `json:"city"`        // Город поиска
	Experience string   `json:"experience"`  // Требуемый опыт работы
	UpdateDays int      `json:"update_days"` // Количество дней с последнего обновления
}

// OutputConfig - настройки форматов вывода
type OutputConfig struct {
	Format string `json:"format"` // Формат вывода (csv, json, sql)
	File   string `json:"file"`   // Файл для сохранения результатов
}

// DatabaseConfig - настройки подключения к PostgreSQL
type DatabaseConfig struct {
	Host     string `json:"host"`     // Хост базы данных
	Port     int    `json:"port"`     // Порт базы данных
	User     string `json:"user"`     // Имя пользователя
	Password string `json:"password"` // Пароль
	DBName   string `json:"db_name"`  // Имя базы данных
}

// GetDefaultConfig - возвращает конфигурацию по умолчанию
func GetDefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			RateLimit: time.Second,
			UserAgent: "HH Resume Parser v2.0",
			Timeout:   30 * time.Second,
		},
		Search: SearchConfig{
			City:       "Moscow",
			UpdateDays: 7,
		},
		Output: OutputConfig{
			Format: "json",
			File:   "resumes.json",
		},
		Database: DatabaseConfig{
			Host:   "localhost",
			Port:   5432,
			User:   "postgres",
			DBName: "resumes",
		},
		LogFile: "parser.log",
	}
}
