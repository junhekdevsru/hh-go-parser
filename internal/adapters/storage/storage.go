package storage

import (
	"context"
	"encoding/json"
	"os"

	"hh-resume-parser/internal/domain/entities"
	"hh-resume-parser/internal/domain/repositories"
	"hh-resume-parser/internal/infrastructure/logger"
)

// FileStorage реализует интерфейс StorageRepository для работы с файлами
type FileStorage struct {
	format string
	file   string
	logger logger.Logger
}

// NewFileStorage создает новый экземпляр файлового хранилища
func NewFileStorage(format, file string, logger logger.Logger) repositories.StorageRepository {
	return &FileStorage{
		format: format,
		file:   file,
		logger: logger,
	}
}

// SaveResumes сохраняет резюме в файл в указанном формате
func (s *FileStorage) SaveResumes(ctx context.Context, resumes []entities.Resume) error {
	s.logger.Info("Сохранение резюме в файл", map[string]interface{}{
		"format": s.format,
		"file":   s.file,
		"count":  len(resumes),
	})

	file, err := os.Create(s.file)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resumes)
}

// GetSavedResumeIDs возвращает список ID сохраненных резюме
func (s *FileStorage) GetSavedResumeIDs(ctx context.Context) ([]string, error) {
	file, err := os.Open(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var resumes []entities.Resume
	if err := json.NewDecoder(file).Decode(&resumes); err != nil {
		return nil, err
	}

	ids := make([]string, len(resumes))
	for i, resume := range resumes {
		ids[i] = resume.ID
	}

	return ids, nil
}
