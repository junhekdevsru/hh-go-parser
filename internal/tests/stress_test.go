package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"hh-resume-parser/internal/app"
	"hh-resume-parser/internal/config"
	"hh-resume-parser/internal/infrastructure/logger"
)

// TestConfig содержит конфигурацию для тестов
type TestConfig struct {
	NumGoroutines int           // Количество параллельных горутин
	RequestsPerGo int           // Количество запросов на горутину
	RateLimit     time.Duration // Ограничение скорости запросов
	Keywords      []string      // Ключевые слова для поиска
	OutputFormats []string      // Форматы вывода для тестирования
}

func TestStressParsingWithDifferentFormats(t *testing.T) {
	testCfg := TestConfig{
		NumGoroutines: 5,
		RequestsPerGo: 10,
		RateLimit:     time.Second,
		Keywords: []string{
			"Go developer",
			"Golang engineer",
			"Backend developer",
		},
		OutputFormats: []string{"json", "csv", "sql"},
	}

	// Создаем временную директорию для тестовых файлов
	tempDir := t.TempDir()

	// Инициализируем логгер для тестов
	testLogger := logger.NewConsoleWithLevel(logger.DEBUG)

	var wg sync.WaitGroup
	errorChan := make(chan error, testCfg.NumGoroutines*testCfg.RequestsPerGo)

	// Запускаем тесты для каждого формата
	for _, format := range testCfg.OutputFormats {
		t.Run(fmt.Sprintf("Format_%s", format), func(t *testing.T) {
			wg.Add(testCfg.NumGoroutines)

			for i := 0; i < testCfg.NumGoroutines; i++ {
				go func(goroutineID int) {
					defer wg.Done()

					for j := 0; j < testCfg.RequestsPerGo; j++ {
						// Создаем конфигурацию для каждого запроса
						cfg := config.GetDefaultConfig()
						cfg.API.RateLimit = testCfg.RateLimit
						cfg.Search.Keywords = testCfg.Keywords
						cfg.Output.Format = format
						cfg.Output.File = fmt.Sprintf("%s/test_output_%d_%d.%s",
							tempDir, goroutineID, j, format)
						cfg.LogFile = fmt.Sprintf("%s/test_log_%d_%d.log",
							tempDir, goroutineID, j)

						// Создаем и запускаем приложение
						application := app.New(cfg, testLogger)
						if err := application.Run(); err != nil {
							errorChan <- fmt.Errorf("goroutine %d, request %d: %v",
								goroutineID, j, err)
						}

						// Соблюдаем rate limit
						time.Sleep(testCfg.RateLimit)
					}
				}(i)
			}

			// Ожидаем завершения всех горутин
			wg.Wait()
			close(errorChan)

			// Проверяем наличие ошибок
			var errors []error
			for err := range errorChan {
				errors = append(errors, err)
			}

			if len(errors) > 0 {
				for _, err := range errors {
					t.Errorf("Ошибка при выполнении теста: %v", err)
				}
			}
		})
	}
}

func TestRateLimitCompliance(t *testing.T) {
	rateLimits := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	for _, rateLimit := range rateLimits {
		t.Run(fmt.Sprintf("RateLimit_%s", rateLimit), func(t *testing.T) {
			cfg := config.GetDefaultConfig()
			cfg.API.RateLimit = rateLimit
			cfg.Search.Keywords = []string{"Go"}
			cfg.Output.Format = "json"
			cfg.Output.File = fmt.Sprintf("test_rate_%s.json", rateLimit)

			testLogger := logger.NewConsoleWithLevel(logger.DEBUG)
			application := app.New(cfg, testLogger)

			start := time.Now()
			err := application.Run()
			elapsed := time.Since(start)

			if err != nil {
				t.Errorf("Ошибка при выполнении теста с rate limit %s: %v", rateLimit, err)
			}

			// Проверяем, что время выполнения соответствует rate limit
			expectedMinTime := rateLimit * time.Duration(5) // Минимум 5 запросов
			if elapsed < expectedMinTime {
				t.Errorf("Rate limiting работает некорректно. Ожидаемое минимальное время: %s, фактическое: %s",
					expectedMinTime, elapsed)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	testCases := []struct {
		name          string
		invalidConfig func() *config.Config
		expectedErr   bool
	}{
		{
			name: "InvalidToken",
			invalidConfig: func() *config.Config {
				cfg := config.GetDefaultConfig()
				cfg.API.Token = "invalid_token"
				return cfg
			},
			expectedErr: true,
		},
		{
			name: "EmptyKeywords",
			invalidConfig: func() *config.Config {
				cfg := config.GetDefaultConfig()
				cfg.Search.Keywords = []string{}
				return cfg
			},
			expectedErr: true,
		},
		{
			name: "InvalidOutputFormat",
			invalidConfig: func() *config.Config {
				cfg := config.GetDefaultConfig()
				cfg.Output.Format = "invalid_format"
				return cfg
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.invalidConfig()
			testLogger := logger.NewConsoleWithLevel(logger.DEBUG)
			application := app.New(cfg, testLogger)

			err := application.Run()

			if tc.expectedErr && err == nil {
				t.Error("Ожидалась ошибка, но её не было")
			}
			if !tc.expectedErr && err != nil {
				t.Errorf("Неожиданная ошибка: %v", err)
			}
		})
	}
}

func TestConcurrentSearchRequests(t *testing.T) {
	cfg := config.GetDefaultConfig()
	cfg.API.RateLimit = 500 * time.Millisecond

	testLogger := logger.NewConsoleWithLevel(logger.DEBUG)

	keywords := []string{
		"Go developer",
		"Backend developer",
		"API developer",
		"Golang engineer",
		"Software engineer",
	}

	var wg sync.WaitGroup
	results := make(chan int, len(keywords))
	errors := make(chan error, len(keywords))

	for _, keyword := range keywords {
		wg.Add(1)
		go func(kw string) {
			defer wg.Done()

			// Создаем конфигурацию для каждого запроса
			cfg := config.GetDefaultConfig()
			cfg.Search.Keywords = []string{kw}
			cfg.Search.City = "Moscow"

			// Создаем и запускаем приложение для каждого поиска
			app := app.New(cfg, testLogger)
			err := app.Run()
			if err != nil {
				errors <- fmt.Errorf("ошибка поиска для '%s': %v", kw, err)
				return
			}

			// Note: В реальном приложении здесь нужно добавить подсчет найденных резюме
			results <- 1 // Временное решение
		}(keyword)

		// Соблюдаем rate limit между запусками горутин
		time.Sleep(cfg.API.RateLimit)
	}

	// Ожидаем завершения всех горутин
	wg.Wait()
	close(results)
	close(errors)

	// Проверяем результаты и ошибки
	totalResults := 0
	for count := range results {
		totalResults += count
	}

	for err := range errors {
		t.Errorf("Ошибка при параллельном поиске: %v", err)
	}

	t.Logf("Всего выполнено поисков: %d", totalResults)
}
