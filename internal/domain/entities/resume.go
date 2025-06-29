package entities

import "time"

// Resume - основная сущность резюме
// Представляет резюме соискателя с полной информацией
type Resume struct {
	ID          string    `json:"id"`                    // Уникальный идентификатор резюме
	Name        string    `json:"name,omitempty"`        // ФИО соискателя
	Skills      []string  `json:"skills,omitempty"`      // Навыки и технологии
	Experience  []Job     `json:"experience,omitempty"`  // Опыт работы
	Education   []Edu     `json:"education,omitempty"`   // Образование
	LastUpdate  time.Time `json:"last_update"`           // Дата последнего обновления
	Contact     Contact   `json:"contact,omitempty"`     // Контактная информация
	URL         string    `json:"url,omitempty"`         // Ссылка на резюме
	Title       string    `json:"title,omitempty"`       // Заголовок резюме
	Salary      *Salary   `json:"salary,omitempty"`      // Желаемая зарплата
	Location    string    `json:"location,omitempty"`    // Местоположение соискателя
	Age         int       `json:"age,omitempty"`         // Возраст соискателя
	Gender      string    `json:"gender,omitempty"`      // Пол соискателя
}

// Job - опыт работы
// Представляет один период работы в компании
type Job struct {
	Company     string `json:"company,omitempty"`     // Название компании
	Position    string `json:"position,omitempty"`    // Должность
	StartDate   string `json:"start_date,omitempty"`  // Дата начала работы
	EndDate     string `json:"end_date,omitempty"`    // Дата окончания работы
	Description string `json:"description,omitempty"` // Описание обязанностей
	Industry    string `json:"industry,omitempty"`    // Отрасль компании
}

// Edu - образование
// Представляет одно учебное заведение или курс
type Edu struct {
	Institution string `json:"institution,omitempty"` // Название учебного заведения
	Faculty     string `json:"faculty,omitempty"`     // Факультет
	Specialty   string `json:"specialty,omitempty"`   // Специальность
	Year        string `json:"year,omitempty"`        // Год окончания
	Level       string `json:"level,omitempty"`       // Уровень образования
}

// Contact - контактная информация
// Содержит способы связи с соискателем
type Contact struct {
	Phone    string   `json:"phone,omitempty"`    // Номер телефона
	Email    string   `json:"email,omitempty"`    // Электронная почта
	Telegram string   `json:"telegram,omitempty"` // Telegram
	Skype    string   `json:"skype,omitempty"`    // Skype
	Social   []Social `json:"social,omitempty"`   // Социальные сети
}

// Social - профиль в социальной сети
type Social struct {
	Type string `json:"type"` // Тип социальной сети (LinkedIn, GitHub, etc.)
	URL  string `json:"url"`  // Ссылка на профиль
}

// Salary - информация о желаемой зарплате
type Salary struct {
	Amount   int    `json:"amount"`   // Сумма
	Currency string `json:"currency"` // Валюта (RUR, USD, EUR)
	Gross    bool   `json:"gross"`    // До налогов (true) или после (false)
}

// IsValid - проверяет валидность резюме
// Возвращает true, если резюме содержит минимально необходимую информацию
func (r *Resume) IsValid() bool {
	return r.ID != "" && (r.Name != "" || r.Title != "")
}

// GetExperienceYears - возвращает общий стаж работы в годах
// Подсчитывает примерное количество лет опыта на основе дат
func (r *Resume) GetExperienceYears() int {
	if len(r.Experience) == 0 {
		return 0
	}

	// Простая оценка: считаем количество записей опыта
	// В реальном приложении можно более точно парсить даты
	return len(r.Experience)
}

// HasSkill - проверяет наличие навыка
// Возвращает true, если в списке навыков есть указанный навык
func (r *Resume) HasSkill(skill string) bool {
	for _, s := range r.Skills {
		if s == skill {
			return true
		}
	}
	return false
}

// GetLatestJob - возвращает последнее место работы
// Возвращает nil, если опыт работы отсутствует
func (r *Resume) GetLatestJob() *Job {
	if len(r.Experience) == 0 {
		return nil
	}
	
	// Возвращаем первую запись (предполагается, что они отсортированы по дате)
	return &r.Experience[0]
}
