package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hh-resume-parser/internal/config"
	"hh-resume-parser/internal/infrastructure/logger"
)

// Client представляет HTTP клиент для работы с API
type Client struct {
	httpClient *http.Client
	config     *config.Config
	logger     logger.Logger
	lastCall   time.Time
}

// NewClient создает новый HTTP клиент
func NewClient(cfg *config.Config, logger logger.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.API.Timeout,
		},
		config:   cfg,
		logger:   logger,
		lastCall: time.Time{},
	}
}

// Do выполняет HTTP запрос с учетом ограничения скорости
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Применяем ограничение скорости
	if !c.lastCall.IsZero() {
		elapsed := time.Since(c.lastCall)
		if elapsed < c.config.API.RateLimit {
			sleepTime := c.config.API.RateLimit - elapsed
			c.logger.Debug("Применяем ограничение скорости", map[string]interface{}{
				"sleep_duration": sleepTime.String(),
			})
			time.Sleep(sleepTime)
		}
	}

	// Добавляем заголовки авторизации
	req.Header.Set("Authorization", "Bearer "+c.config.API.Token)
	req.Header.Set("User-Agent", c.config.API.UserAgent)
	req.Header.Set("Accept", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	c.lastCall = time.Now()

	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Errors []string `json:"errors"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("API вернул статус %d: %v", resp.StatusCode, errResp.Errors)
		}
		return nil, fmt.Errorf("API вернул статус %d", resp.StatusCode)
	}

	return resp, nil
}
