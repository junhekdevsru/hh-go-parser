package usecases

import (
	"context"
	"fmt"
	"hh-resume-parser/internal/domain/entities"
	"hh-resume-parser/internal/domain/repositories"
	"hh-resume-parser/internal/infrastructure/logger"
)

// ResumeUseCase - основной сценарий использования для работы с резюме
// Содержит бизнес-логику парсинга и обработки резюме
type ResumeUseCase struct {
	resumeRepo  repositories.ResumeRepository  // Репозиторий для получения резюме
	storageRepo repositories.StorageRepository // Репозиторий для сохранения данных
	cacheRepo   repositories.CacheRepository   // Репозиторий для кэширования
	logger      logger.Logger                  // Логгер для записи событий
	processed   map[string]bool                // Кэш обработанных резюме в памяти
}

// NewResumeUseCase - создание нового экземпляра use case
func NewResumeUseCase(
	resumeRepo repositories.ResumeRepository,
	storageRepo repositories.StorageRepository,
	cacheRepo repositories.CacheRepository,
	logger logger.Logger,
) *ResumeUseCase {
	return &ResumeUseCase{
		resumeRepo:  resumeRepo,
		storageRepo: storageRepo,
		cacheRepo:   cacheRepo,
		logger:      logger,
		processed:   make(map[string]bool),
	}
}

// ParseResumesByCriteria - основной метод парсинга резюме по критериям
// Выполняет поиск, фильтрацию и сохранение резюме
func (uc *ResumeUseCase) ParseResumesByCriteria(ctx context.Context, criteria repositories.SearchCriteria) (*ParseResult, error) {
	uc.logger.Info("Начинаем парсинг резюме", map[string]interface{}{
		"keywords":    criteria.Keywords,
		"city":        criteria.City,
		"experience":  criteria.Experience,
		"update_days": criteria.UpdateDays,
	})

	// Инициализация результата
	result := &ParseResult{
		ProcessedCount: 0,
		SavedCount:     0,
		SkippedCount:   0,
		Errors:         make([]error, 0),
	}

	// Загрузка списка уже сохраненных резюме для предотвращения дублирования
	if err := uc.loadProcessedResumes(ctx); err != nil {
		uc.logger.Error("Ошибка загрузки списка обработанных резюме", err)
		// Продолжаем работу без предварительной загрузки
	}

	// Поиск резюме по критериям
	var allResumes []entities.Resume
	page := 0

	for {
		criteria.Page = page
		resumes, err := uc.resumeRepo.SearchResumes(ctx, criteria)
		if err != nil {
			uc.logger.Error("Ошибка поиска резюме", err)
			result.Errors = append(result.Errors, fmt.Errorf("ошибка поиска на странице %d: %w", page, err))
			break
		}

		if len(resumes) == 0 {
			uc.logger.Info("Достигнут конец результатов поиска", map[string]interface{}{"page": page})
			break
		}

		// Обработка найденных резюме
		for _, resume := range resumes {
			result.ProcessedCount++

			// Проверка на дублирование
			if uc.isAlreadyProcessed(resume.ID) {
				result.SkippedCount++
				uc.logger.Debug("Резюме уже обработано, пропускаем", map[string]interface{}{"resume_id": resume.ID})
				continue
			}

			// Валидация резюме
			if !uc.validateResume(&resume) {
				result.SkippedCount++
				uc.logger.Debug("Резюме не прошло валидацию", map[string]interface{}{"resume_id": resume.ID})
				continue
			}

			// Обогащение данных резюме
			if err := uc.enrichResumeData(ctx, &resume); err != nil {
				uc.logger.Error("Ошибка обогащения данных резюме", err)
				// Продолжаем с основными данными
			}

			allResumes = append(allResumes, resume)
			uc.markAsProcessed(resume.ID)
			result.SavedCount++
		}

		uc.logger.Info("Обработана страница результатов", map[string]interface{}{
			"page":            page,
			"resumes_on_page": len(resumes),
			"total_saved":     result.SavedCount,
		})

		page++
	}

	// Сохранение результатов
	if len(allResumes) > 0 {
		if err := uc.storageRepo.SaveResumes(ctx, allResumes); err != nil {
			uc.logger.Error("Ошибка сохранения резюме", err)
			return result, fmt.Errorf("ошибка сохранения резюме: %w", err)
		}

		uc.logger.Info("Резюме успешно сохранены", map[string]interface{}{
			"count": len(allResumes),
		})
	}

	result.TotalFound = result.ProcessedCount
	return result, nil
}

// GetResumeDetails - получение детальной информации о резюме
func (uc *ResumeUseCase) GetResumeDetails(ctx context.Context, resumeID string) (*entities.Resume, error) {
	uc.logger.Info("Получение детальной информации о резюме", map[string]interface{}{
		"resume_id": resumeID,
	})

	// Попытка получить из кэша
	if uc.cacheRepo != nil {
		if cached := uc.getCachedResume(ctx, resumeID); cached != nil {
			uc.logger.Debug("Резюме получено из кэша", map[string]interface{}{"resume_id": resumeID})
			return cached, nil
		}
	}

	// Получение из основного источника
	resume, err := uc.resumeRepo.GetResumeByID(ctx, resumeID)
	if err != nil {
		uc.logger.Error("Ошибка получения резюме", err)
		return nil, fmt.Errorf("ошибка получения резюме %s: %w", resumeID, err)
	}

	// Кэширование результата
	if uc.cacheRepo != nil {
		uc.cacheResume(ctx, resume)
	}

	return resume, nil
}

// validateResume - валидация резюме перед сохранением
func (uc *ResumeUseCase) validateResume(resume *entities.Resume) bool {
	// Базовая валидация
	if !resume.IsValid() {
		return false
	}

	// Дополнительные проверки
	if len(resume.Skills) == 0 && len(resume.Experience) == 0 {
		uc.logger.Debug("Резюме без навыков и опыта работы", map[string]interface{}{
			"resume_id": resume.ID,
		})
		return false
	}

	return true
}

// enrichResumeData - обогащение данных резюме дополнительной информацией
func (uc *ResumeUseCase) enrichResumeData(ctx context.Context, resume *entities.Resume) error {
	// Здесь можно добавить логику обогащения данных:
	// - Получение дополнительной информации из API
	// - Нормализация данных
	// - Извлечение дополнительных навыков из описаний
	
	// Пример: нормализация навыков
	resume.Skills = uc.normalizeSkills(resume.Skills)
	
	return nil
}

// normalizeSkills - нормализация списка навыков
func (uc *ResumeUseCase) normalizeSkills(skills []string) []string {
	// Удаление дублей и нормализация
	seen := make(map[string]bool)
	var normalized []string
	
	for _, skill := range skills {
		if skill != "" && !seen[skill] {
			seen[skill] = true
			normalized = append(normalized, skill)
		}
	}
	
	return normalized
}

// loadProcessedResumes - загрузка списка уже обработанных резюме
func (uc *ResumeUseCase) loadProcessedResumes(ctx context.Context) error {
	savedIDs, err := uc.storageRepo.GetSavedResumeIDs(ctx)
	if err != nil {
		return err
	}

	for _, id := range savedIDs {
		uc.processed[id] = true
	}

	uc.logger.Info("Загружен список обработанных резюме", map[string]interface{}{
		"count": len(savedIDs),
	})

	return nil
}

// isAlreadyProcessed - проверка, было ли резюме уже обработано
func (uc *ResumeUseCase) isAlreadyProcessed(resumeID string) bool {
	return uc.processed[resumeID]
}

// markAsProcessed - отметка резюме как обработанного
func (uc *ResumeUseCase) markAsProcessed(resumeID string) {
	uc.processed[resumeID] = true
}

// getCachedResume - получение резюме из кэша
func (uc *ResumeUseCase) getCachedResume(ctx context.Context, resumeID string) *entities.Resume {
	// В реальном приложении здесь будет десериализация из кэша
	return nil
}

// cacheResume - сохранение резюме в кэш
func (uc *ResumeUseCase) cacheResume(ctx context.Context, resume *entities.Resume) {
	// В реальном приложении здесь будет сериализация в кэш
}

// ParseResult - результат выполнения парсинга
type ParseResult struct {
	TotalFound     int     // Общее количество найденных резюме
	ProcessedCount int     // Количество обработанных резюме
	SavedCount     int     // Количество сохраненных резюме
	SkippedCount   int     // Количество пропущенных резюме
	Errors         []error // Список ошибок, возникших в процессе
}
