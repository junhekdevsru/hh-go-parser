package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"encoding/json"
	"hh-resume-parser/internal/app"
	"hh-resume-parser/internal/config"
	"hh-resume-parser/internal/infrastructure/logger"
)

// main - точка входа в приложение
// Инициализирует конфигурацию, зависимости и запускает парсер резюме
func main() {
	// Инициализация конфигурации
	cfg := parseFlags()

	// Валидация конфигурации
	if err := validateConfig(cfg); err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	// Создание логгера
	appLogger, err := logger.New(cfg.LogFile)
	if err != nil {
		log.Fatalf("Не удалось создать логгер: %v", err)
	}
	defer appLogger.Close()

	// Инициализация и запуск приложения
	application := app.New(cfg, appLogger)
	if err := application.Run(); err != nil {
		appLogger.Error("Ошибка выполнения парсинга", err)
		os.Exit(1)
	}

	fmt.Println("✅ Парсинг резюме завершен успешно!")
}

// parseFlags - парсинг аргументов командной строки
func parseFlags() *config.Config {
	cfg := config.GetDefaultConfig()

	flag.StringVar(&cfg.API.Token, "token", os.Getenv("HH_API_TOKEN"), "API токен hh.ru")
	flag.StringVar(&cfg.Search.City, "city", cfg.Search.City, "Город для поиска")
	flag.StringVar(&cfg.Search.Experience, "experience", cfg.Search.Experience, "Уровень опыта")
	flag.IntVar(&cfg.Search.UpdateDays, "update-days", cfg.Search.UpdateDays, "Дни обновления")
	flag.StringVar(&cfg.Output.Format, "format", cfg.Output.Format, "Формат вывода (json, csv, sql)")
	flag.StringVar(&cfg.Output.File, "output", cfg.Output.File, "Файл вывода")
	flag.StringVar(&cfg.LogFile, "log", cfg.LogFile, "Файл логов")

	// Парсинг ключевых слов
	var keywords string
	flag.StringVar(&keywords, "keywords", "", "Ключевые слова (через запятую)")

	var keywordsFile string
	flag.StringVar(&keywordsFile, "keywords-file", "", "Файл с ключевыми словами")

	flag.Parse()

	// Загрузка ключевых слов
	if keywordsFile != "" {
		if keywords, err := loadKeywordsFromFile(keywordsFile); err == nil {
			cfg.Search.Keywords = keywords
		}
	} else if keywords != "" {
		cfg.Search.Keywords = strings.Split(keywords, ",")
		for i, kw := range cfg.Search.Keywords {
			cfg.Search.Keywords[i] = strings.TrimSpace(kw)
		}
	}

	return cfg
}

// validateConfig - валидация конфигурации
func validateConfig(cfg *config.Config) error {
	if cfg.API.Token == "" {
		return fmt.Errorf("не указан API токен")
	}

	if len(cfg.Search.Keywords) == 0 {
		return fmt.Errorf("не указаны ключевые слова для поиска")
	}

	return nil
}

// loadKeywordsFromFile - загрузка ключевых слов из файла
func loadKeywordsFromFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var keywords []string
	if err := parseJSONKeywords(content, &keywords); err != nil {
		// Если не JSON, читаем как текст построчно
		return strings.Split(strings.TrimSpace(string(content)), "\n"), nil
	}

	return keywords, nil
}

// parseJSONKeywords - парсинг ключевых слов из JSON
func parseJSONKeywords(content []byte, keywords *[]string) error {
	return json.Unmarshal(content, keywords)
}
