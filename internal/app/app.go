package app

import (
	"context"
	"fmt"
	"time"

	"hh-resume-parser/internal/adapters/storage"
	"hh-resume-parser/internal/config"
	"hh-resume-parser/internal/domain/repositories"
	"hh-resume-parser/internal/domain/usecases"
	"hh-resume-parser/internal/infrastructure/logger"
	hhrepo "hh-resume-parser/internal/infrastructure/repositories"
)

// Application представляет основное приложение
type Application struct {
	config     *config.Config
	logger     logger.Logger
	useCase    *usecases.ResumeUseCase
	repository repositories.ResumeRepository
	storage    repositories.StorageRepository
}

// New создает новый экземпляр приложения
func New(cfg *config.Config, logger logger.Logger) *Application {
	// Создаем репозиторий для работы с API
	repository := hhrepo.NewHHRepository(cfg, logger)

	// Выбираем подходящий адаптер хранилища на основе конфигурации
	var fileStorage repositories.StorageRepository
	switch cfg.Output.Format {
	case "csv":
		fileStorage = storage.NewCSVStorage(cfg.Output.File, logger)
	case "sql":
		fileStorage = storage.NewSQLStorage(cfg.Output.File, logger)
	default: // json по умолчанию
		fileStorage = storage.NewFileStorage(cfg.Output.Format, cfg.Output.File, logger)
	}

	// Создаем основной use case
	useCase := usecases.NewResumeUseCase(repository, fileStorage, nil, logger)

	return &Application{
		config:     cfg,
		logger:     logger,
		useCase:    useCase,
		repository: repository,
		storage:    fileStorage,
	}
}

// Run запускает основной процесс парсинга резюме
func (a *Application) Run() error {
	ctx := context.Background()

	// Создаем критерии поиска из конфигурации
	criteria := repositories.SearchCriteria{
		Keywords:   a.config.Search.Keywords,
		City:       a.config.Search.City,
		Experience: a.config.Search.Experience,
		UpdateDays: a.config.Search.UpdateDays,
		PerPage:    20,
	}

	a.logger.Info("Запуск парсинга резюме", map[string]interface{}{
		"keywords":    criteria.Keywords,
		"city":        criteria.City,
		"experience":  criteria.Experience,
		"update_days": criteria.UpdateDays,
		"format":      a.config.Output.Format,
		"output":      a.config.Output.File,
	})

	// Запускаем процесс парсинга
	startTime := time.Now()

	result, err := a.useCase.ParseResumesByCriteria(ctx, criteria)
	if err != nil {
		return fmt.Errorf("ошибка парсинга: %w", err)
	}

	// Логируем результаты
	a.logger.Info("Парсинг завершен", map[string]interface{}{
		"total_found":  result.TotalFound,
		"saved":        result.SavedCount,
		"skipped":      result.SkippedCount,
		"errors":       len(result.Errors),
		"elapsed_time": time.Since(startTime).String(),
	})

	return nil
}
