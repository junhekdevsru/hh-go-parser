package storage

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"hh-resume-parser/internal/domain/entities"
	"hh-resume-parser/internal/domain/repositories"
	"hh-resume-parser/internal/infrastructure/logger"
)

// SQLStorage реализует интерфейс StorageRepository для SQL формата
type SQLStorage struct {
	file   string
	logger logger.Logger
}

// NewSQLStorage создает новый экземпляр SQL хранилища
func NewSQLStorage(file string, logger logger.Logger) repositories.StorageRepository {
	return &SQLStorage{
		file:   file,
		logger: logger,
	}
}

// SaveResumes сохраняет резюме в SQL скрипт
func (s *SQLStorage) SaveResumes(ctx context.Context, resumes []entities.Resume) error {
	file, err := os.Create(s.file)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	// Записываем схему таблиц
	if err := s.writeSchema(file); err != nil {
		return err
	}

	// Записываем данные
	for _, resume := range resumes {
		if err := s.writeResume(file, resume); err != nil {
			return err
		}
	}

	s.logger.Info("Резюме сохранены в SQL", map[string]interface{}{
		"file":  s.file,
		"count": len(resumes),
	})

	return nil
}

// GetSavedResumeIDs возвращает список ID сохраненных резюме
func (s *SQLStorage) GetSavedResumeIDs(ctx context.Context) ([]string, error) {
	// В случае SQL скрипта мы не можем получить список сохраненных ID
	// без реального подключения к базе данных
	return []string{}, nil
}

// writeSchema записывает SQL схему таблиц
func (s *SQLStorage) writeSchema(file *os.File) error {
	schema := `
-- Схема базы данных для резюме
CREATE TABLE IF NOT EXISTS resumes (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500),
    title VARCHAR(500),
    skills TEXT,
    last_update TIMESTAMP,
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),
    url VARCHAR(500),
    location VARCHAR(255),
    age INT,
    gender VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS experience (
    id SERIAL PRIMARY KEY,
    resume_id VARCHAR(255) REFERENCES resumes(id),
    company VARCHAR(500),
    position VARCHAR(500),
    start_date VARCHAR(50),
    end_date VARCHAR(50),
    description TEXT,
    industry VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS education (
    id SERIAL PRIMARY KEY,
    resume_id VARCHAR(255) REFERENCES resumes(id),
    institution VARCHAR(500),
    faculty VARCHAR(500),
    specialty VARCHAR(500),
    year VARCHAR(50),
    level VARCHAR(255)
);

-- Индексы для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_resumes_skills ON resumes USING gin (to_tsvector('russian', skills));
CREATE INDEX IF NOT EXISTS idx_resumes_location ON resumes(location);
CREATE INDEX IF NOT EXISTS idx_experience_company ON experience(company);
CREATE INDEX IF NOT EXISTS idx_education_institution ON education(institution);

`
	_, err := file.WriteString(schema)
	return err
}

// writeResume записывает одно резюме в SQL формате
func (s *SQLStorage) writeResume(file *os.File, resume entities.Resume) error {
	// Основная информация о резюме
	mainSQL := fmt.Sprintf(`
INSERT INTO resumes (
    id, name, title, skills, last_update,
    contact_phone, contact_email, url,
    location, age, gender
) VALUES (
    '%s', '%s', '%s', '%s', '%s',
    '%s', '%s', '%s',
    '%s', %d, '%s'
) ON CONFLICT (id) DO UPDATE SET
    last_update = EXCLUDED.last_update,
    skills = EXCLUDED.skills;
`,
		escape(resume.ID),
		escape(resume.Name),
		escape(resume.Title),
		escape(strings.Join(resume.Skills, "; ")),
		resume.LastUpdate.Format(time.RFC3339),
		escape(resume.Contact.Phone),
		escape(resume.Contact.Email),
		escape(resume.URL),
		escape(resume.Location),
		resume.Age,
		escape(resume.Gender),
	)

	if _, err := file.WriteString(mainSQL); err != nil {
		return err
	}

	// Опыт работы
	for _, job := range resume.Experience {
		expSQL := fmt.Sprintf(`
INSERT INTO experience (
    resume_id, company, position,
    start_date, end_date, description, industry
) VALUES (
    '%s', '%s', '%s',
    '%s', '%s', '%s', '%s'
);
`,
			escape(resume.ID),
			escape(job.Company),
			escape(job.Position),
			escape(job.StartDate),
			escape(job.EndDate),
			escape(job.Description),
			escape(job.Industry),
		)

		if _, err := file.WriteString(expSQL); err != nil {
			return err
		}
	}

	// Образование
	for _, edu := range resume.Education {
		eduSQL := fmt.Sprintf(`
INSERT INTO education (
    resume_id, institution, faculty,
    specialty, year, level
) VALUES (
    '%s', '%s', '%s',
    '%s', '%s', '%s'
);
`,
			escape(resume.ID),
			escape(edu.Institution),
			escape(edu.Faculty),
			escape(edu.Specialty),
			escape(edu.Year),
			escape(edu.Level),
		)

		if _, err := file.WriteString(eduSQL); err != nil {
			return err
		}
	}

	return nil
}

// escape экранирует специальные символы в SQL строках
func escape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
