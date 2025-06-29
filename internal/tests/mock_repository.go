package tests

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"hh-resume-parser/internal/domain/entities"
	"hh-resume-parser/internal/domain/repositories"
)

// MockRepository implements ResumeRepository interface for testing
type MockRepository struct {
	delay    time.Duration // Симуляция задержки сети
	failRate float64       // Вероятность ошибки (0-1)
}

// NewMockRepository creates a new mock repository
func NewMockRepository(delay time.Duration, failRate float64) repositories.ResumeRepository {
	return &MockRepository{
		delay:    delay,
		failRate: failRate,
	}
}

// SearchResumes simulates resume search with configurable delay and errors
func (m *MockRepository) SearchResumes(ctx context.Context, criteria repositories.SearchCriteria) ([]entities.Resume, error) {
	// Симуляция задержки сети
	time.Sleep(m.delay)

	// Симуляция случайных ошибок
	if rand.Float64() < m.failRate {
		return nil, fmt.Errorf("simulated network error")
	}

	// Генерируем тестовые резюме
	resumes := make([]entities.Resume, 0, 20)
	for i := 0; i < 20; i++ {
		resume := m.generateMockResume(criteria)
		resumes = append(resumes, resume)
	}

	return resumes, nil
}

// GetResumeByID simulates getting a resume by ID
func (m *MockRepository) GetResumeByID(ctx context.Context, id string) (*entities.Resume, error) {
	time.Sleep(m.delay)

	if rand.Float64() < m.failRate {
		return nil, fmt.Errorf("simulated error getting resume %s", id)
	}

	resume := m.generateMockResume(repositories.SearchCriteria{})
	resume.ID = id
	return &resume, nil
}

// generateMockResume creates a mock resume with random data
func (m *MockRepository) generateMockResume(criteria repositories.SearchCriteria) entities.Resume {
	skills := []string{"Go", "Docker", "Kubernetes", "PostgreSQL", "REST API"}
	companies := []string{"Tech Corp", "Innovation Labs", "Software House", "Digital Solutions"}
	positions := []string{"Senior Go Developer", "Backend Engineer", "Software Architect", "Tech Lead"}
	universities := []string{"Moscow State University", "ITMO University", "Bauman Moscow State Technical University"}

	return entities.Resume{
		ID:         fmt.Sprintf("resume_%d", rand.Int31()),
		Name:       fmt.Sprintf("Test User %d", rand.Int31()),
		Skills:     skills[rand.Intn(len(skills)):],
		LastUpdate: time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
		Experience: []entities.Job{
			{
				Company:   companies[rand.Intn(len(companies))],
				Position:  positions[rand.Intn(len(positions))],
				StartDate: "2020-01",
				EndDate:   "2023-12",
			},
		},
		Education: []entities.Edu{
			{
				Institution: universities[rand.Intn(len(universities))],
				Faculty:     "Computer Science",
				Specialty:   "Software Engineering",
				Year:        "2020",
			},
		},
		Contact: entities.Contact{
			Phone: fmt.Sprintf("+7-9%d", rand.Int31n(999999999)),
			Email: fmt.Sprintf("user%d@example.com", rand.Int31()),
		},
		URL: fmt.Sprintf("https://hh.ru/resume/%d", rand.Int31()),
	}
}
