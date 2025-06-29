package storage

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"hh-resume-parser/internal/domain/entities"
	"hh-resume-parser/internal/domain/repositories"
	"hh-resume-parser/internal/infrastructure/logger"
)

// CSVStorage реализует интерфейс StorageRepository для CSV формата
type CSVStorage struct {
	file   string
	logger logger.Logger
}

// NewCSVStorage создает новый экземпляр CSV хранилища
func NewCSVStorage(file string, logger logger.Logger) repositories.StorageRepository {
	return &CSVStorage{
		file:   file,
		logger: logger,
	}
}

// SaveResumes сохраняет резюме в CSV формате
func (s *CSVStorage) SaveResumes(ctx context.Context, resumes []entities.Resume) error {
	file, err := os.Create(s.file)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записываем заголовки
	headers := []string{
		"ID", "Name", "Title", "Skills", "Experience",
		"Education", "Last Update", "Phone", "Email",
		"URL", "Location", "Age", "Gender",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("ошибка записи заголовков: %w", err)
	}

	// Записываем данные
	for _, resume := range resumes {
		record := []string{
			resume.ID,
			resume.Name,
			resume.Title,
			s.formatSkills(resume.Skills),
			s.formatExperience(resume.Experience),
			s.formatEducation(resume.Education),
			resume.LastUpdate.Format("2006-01-02 15:04:05"),
			resume.Contact.Phone,
			resume.Contact.Email,
			resume.URL,
			resume.Location,
			fmt.Sprintf("%d", resume.Age),
			resume.Gender,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("ошибка записи резюме %s: %w", resume.ID, err)
		}
	}

	s.logger.Info("Резюме сохранены в CSV", map[string]interface{}{
		"file":  s.file,
		"count": len(resumes),
	})

	return nil
}

// GetSavedResumeIDs возвращает список ID сохраненных резюме
func (s *CSVStorage) GetSavedResumeIDs(ctx context.Context) ([]string, error) {
	file, err := os.Open(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Пропускаем заголовки
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var ids []string
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if len(record) > 0 {
			ids = append(ids, record[0]) // ID в первой колонке
		}
	}

	return ids, nil
}

// Вспомогательные методы для форматирования данных

func (s *CSVStorage) formatSkills(skills []string) string {
	return joinStrings(skills, "; ")
}

func (s *CSVStorage) formatExperience(experience []entities.Job) string {
	var parts []string
	for _, job := range experience {
		part := fmt.Sprintf("%s at %s (%s - %s)",
			job.Position, job.Company, job.StartDate, job.EndDate)
		parts = append(parts, part)
	}
	return joinStrings(parts, " | ")
}

func (s *CSVStorage) formatEducation(education []entities.Edu) string {
	var parts []string
	for _, edu := range education {
		part := fmt.Sprintf("%s, %s (%s)",
			edu.Institution, edu.Specialty, edu.Year)
		parts = append(parts, part)
	}
	return joinStrings(parts, " | ")
}

func joinStrings(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for _, item := range items[1:] {
		result += sep + item
	}
	return result
}
