package repositories

import (
	"context"
	"hh-resume-parser/internal/domain/entities"
)

// ResumeRepository - интерфейс репозитория для работы с резюме
// Определяет контракт для получения данных о резюме из внешних источников
type ResumeRepository interface {
	// SearchResumes - поиск резюме по заданным критериям
	// Возвращает список резюме и ошибку, если поиск не удался
	SearchResumes(ctx context.Context, criteria SearchCriteria) ([]entities.Resume, error)
	
	// GetResumeByID - получение резюме по идентификатору
	// Возвращает детальную информацию о резюме
	GetResumeByID(ctx context.Context, id string) (*entities.Resume, error)
}

// StorageRepository - интерфейс репозитория для сохранения данных
// Определяет контракт для сохранения резюме в различных форматах
type StorageRepository interface {
	// SaveResumes - сохранение списка резюме
	// Поддерживает различные форматы вывода
	SaveResumes(ctx context.Context, resumes []entities.Resume) error
	
	// GetSavedResumeIDs - получение идентификаторов уже сохраненных резюме
	// Используется для предотвращения дублирования
	GetSavedResumeIDs(ctx context.Context) ([]string, error)
}

// SearchCriteria - критерии поиска резюме
// Содержит параметры для фильтрации результатов поиска
type SearchCriteria struct {
	Keywords   []string // Ключевые слова для поиска
	City       string   // Город поиска
	Experience string   // Требуемый опыт работы
	UpdateDays int      // Количество дней с последнего обновления
	Page       int      // Номер страницы результатов
	PerPage    int      // Количество результатов на странице
}

// SearchResult - результат поиска резюме
// Содержит найденные резюме и метаинформацию о поиске
type SearchResult struct {
	Resumes    []entities.Resume // Найденные резюме
	TotalFound int               // Общее количество найденных резюме
	TotalPages int               // Общее количество страниц
	CurrentPage int              // Текущая страница
}

// CacheRepository - интерфейс для кэширования данных
// Используется для оптимизации повторных запросов
type CacheRepository interface {
	// Get - получение данных из кэша
	Get(ctx context.Context, key string) ([]byte, error)
	
	// Set - сохранение данных в кэш
	Set(ctx context.Context, key string, value []byte, ttl int) error
	
	// Delete - удаление данных из кэша
	Delete(ctx context.Context, key string) error
	
	// Exists - проверка существования ключа в кэше
	Exists(ctx context.Context, key string) bool
}
