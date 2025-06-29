package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hh-resume-parser/internal/config"
	"hh-resume-parser/internal/domain/entities"
	"hh-resume-parser/internal/domain/repositories"
	"hh-resume-parser/internal/infrastructure/logger"
)

// hhRepository - реализация репозитория для работы с API hh.ru
type hhRepository struct {
	client    *http.Client     // HTTP клиент для запросов
	config    *config.Config   // Конфигурация приложения
	logger    logger.Logger    // Логгер
	rateLimit time.Duration    // Ограничение скорости запросов
	lastCall  time.Time        // Время последнего вызова API
}

// NewHHRepository - создание нового репозитория для hh.ru
func NewHHRepository(cfg *config.Config, logger logger.Logger) repositories.ResumeRepository {
	return &hhRepository{
		client: &http.Client{
			Timeout: cfg.API.Timeout,
		},
		config:    cfg,
		logger:    logger,
		rateLimit: cfg.API.RateLimit,
		lastCall:  time.Time{},
	}
}

// SearchResumes - поиск резюме по критериям через API hh.ru
func (r *hhRepository) SearchResumes(ctx context.Context, criteria repositories.SearchCriteria) ([]entities.Resume, error) {
	// Применение ограничения скорости
	r.applyRateLimit()

	// Построение URL для поиска
	searchURL := r.buildSearchURL(criteria)
	
	r.logger.Info("Выполняем запрос к API hh.ru", map[string]interface{}{
		"url":  searchURL,
		"page": criteria.Page,
	})

	// Выполнение HTTP запроса
	resp, err := r.makeAPIRequest(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к API: %w", err)
	}
	defer resp.Body.Close()

	// Парсинг ответа
	var apiResponse HHAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа API: %w", err)
	}

	// Конвертация в доменные сущности
	resumes := make([]entities.Resume, 0, len(apiResponse.Items))
	for _, item := range apiResponse.Items {
		resume := r.convertToResume(item)
		resumes = append(resumes, resume)
	}

	r.logger.Info("Получены резюме из API", map[string]interface{}{
		"count":       len(resumes),
		"total_found": apiResponse.Found,
		"page":        apiResponse.Page,
		"total_pages": apiResponse.Pages,
	})

	return resumes, nil
}

// GetResumeByID - получение детального резюме по ID
func (r *hhRepository) GetResumeByID(ctx context.Context, id string) (*entities.Resume, error) {
	// Применение ограничения скорости
	r.applyRateLimit()

	detailURL := fmt.Sprintf("https://api.hh.ru/resumes/%s", id)
	
	r.logger.Debug("Получаем детальную информацию о резюме", map[string]interface{}{
		"resume_id": id,
		"url":       detailURL,
	})

	// Выполнение HTTP запроса
	resp, err := r.makeAPIRequest(ctx, detailURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения резюме %s: %w", id, err)
	}
	defer resp.Body.Close()

	// Парсинг ответа
	var apiItem HHResumeItem
	if err := json.NewDecoder(resp.Body).Decode(&apiItem); err != nil {
		return nil, fmt.Errorf("ошибка парсинга резюме %s: %w", id, err)
	}

	// Конвертация в доменную сущность
	resume := r.convertToResume(apiItem)
	
	return &resume, nil
}

// applyRateLimit - применение ограничения скорости запросов
func (r *hhRepository) applyRateLimit() {
	if !r.lastCall.IsZero() {
		elapsed := time.Since(r.lastCall)
		if elapsed < r.rateLimit {
			sleepTime := r.rateLimit - elapsed
			r.logger.Debug("Применяем ограничение скорости", map[string]interface{}{
				"sleep_duration": sleepTime.String(),
			})
			time.Sleep(sleepTime)
		}
	}
	r.lastCall = time.Now()
}

// makeAPIRequest - выполнение HTTP запроса к API
func (r *hhRepository) makeAPIRequest(ctx context.Context, requestURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Установка заголовков
	req.Header.Set("Authorization", "Bearer "+r.config.API.Token)
	req.Header.Set("User-Agent", r.config.API.UserAgent)
	req.Header.Set("Accept", "application/json")

	// Выполнение запроса
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		r.logger.Error("API вернул ошибку", fmt.Errorf("статус %d", resp.StatusCode))
		
		return nil, fmt.Errorf("API вернул статус %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// buildSearchURL - построение URL для поиска резюме
func (r *hhRepository) buildSearchURL(criteria repositories.SearchCriteria) string {
	baseURL := "https://api.hh.ru/resumes"
	params := url.Values{}

	// Добавление номера страницы
	params.Add("page", strconv.Itoa(criteria.Page))

	// Добавление ключевых слов
	if len(criteria.Keywords) > 0 {
		text := strings.Join(criteria.Keywords, " ")
		params.Add("text", text)
	}

	// Добавление города
	if criteria.City != "" {
		areaID := r.getCityAreaID(criteria.City)
		if areaID != "" {
			params.Add("area", areaID)
		}
	}

	// Добавление фильтра по опыту
	if criteria.Experience != "" {
		params.Add("experience", criteria.Experience)
	}

	// Добавление фильтра по дате обновления
	if criteria.UpdateDays > 0 {
		params.Add("period", strconv.Itoa(criteria.UpdateDays))
	}

	// Добавление количества элементов на странице
	if criteria.PerPage > 0 {
		params.Add("per_page", strconv.Itoa(criteria.PerPage))
	} else {
		params.Add("per_page", "20") // Значение по умолчанию
	}

	return baseURL + "?" + params.Encode()
}

// getCityAreaID - получение ID региона по названию города
func (r *hhRepository) getCityAreaID(cityName string) string {
	// Маппинг популярных городов к их ID в API hh.ru
	cityMap := map[string]string{
		"москва":        "1",
		"moscow":        "1",
		"санкт-петербург": "2",
		"спб":           "2",
		"saint petersburg": "2",
		"екатеринбург":   "3",
		"ekaterinburg":  "3",
		"новосибирск":   "4",
		"novosibirsk":   "4",
		"нижний новгород": "5",
		"nizhny novgorod": "5",
		"казань":        "88",
		"kazan":         "88",
		"челябинск":     "91",
		"chelyabinsk":   "91",
		"омск":          "95",
		"omsk":          "95",
		"самара":        "78",
		"samara":        "78",
		"ростов-на-дону": "76",
		"rostov-on-don": "76",
		"уфа":           "99",
		"ufa":           "99",
		"красноярск":    "26",
		"krasnoyarsk":   "26",
		"воронеж":       "193",
		"voronezh":      "193",
		"пермь":         "75",
		"perm":          "75",
		"волгоград":     "24",
		"volgograd":     "24",
		"краснодар":     "53",
		"krasnodar":     "53",
	}

	// Нормализация названия города
	normalizedCity := strings.ToLower(strings.TrimSpace(cityName))
	
	if areaID, exists := cityMap[normalizedCity]; exists {
		return areaID
	}

	r.logger.Warn("Неизвестный город, используем поиск без фильтра по региону", map[string]interface{}{
		"city": cityName,
	})
	
	return ""
}

// convertToResume - конвертация данных API в доменную сущность
func (r *hhRepository) convertToResume(apiItem HHResumeItem) entities.Resume {
	resume := entities.Resume{
		ID:    apiItem.ID,
		Title: apiItem.Title,
		URL:   apiItem.URL,
	}

	// Формирование полного имени
	if apiItem.FirstName != "" || apiItem.LastName != "" {
		resume.Name = strings.TrimSpace(apiItem.FirstName + " " + apiItem.LastName)
	}

	// Парсинг даты последнего обновления
	if apiItem.UpdatedAt != "" {
		if updatedTime, err := time.Parse(time.RFC3339, apiItem.UpdatedAt); err == nil {
			resume.LastUpdate = updatedTime
		}
	}

	// Конвертация навыков
	resume.Skills = make([]string, 0, len(apiItem.Skills))
	for _, skill := range apiItem.Skills {
		if skill.Name != "" {
			resume.Skills = append(resume.Skills, skill.Name)
		}
	}

	// Конвертация опыта работы
	resume.Experience = make([]entities.Job, 0, len(apiItem.Experience))
	for _, exp := range apiItem.Experience {
		job := entities.Job{
			Company:     exp.Company.Name,
			Position:    exp.Position,
			Description: exp.Description,
		}
		
		// Парсинг дат работы
		if exp.Start != "" {
			job.StartDate = exp.Start
		}
		if exp.End != "" {
			job.EndDate = exp.End
		}
		
		resume.Experience = append(resume.Experience, job)
	}

	// Конвертация образования
	resume.Education = make([]entities.Edu, 0, len(apiItem.Education))
	for _, edu := range apiItem.Education {
		education := entities.Edu{
			Institution: edu.Name,
			Faculty:     edu.Organization,
			Specialty:   edu.Result,
			Year:        strconv.Itoa(edu.Year),
		}
		resume.Education = append(resume.Education, education)
	}

	// Конвертация контактной информации
	if apiItem.Contact.Phone.Formatted != "" || apiItem.Contact.Email.Email != "" {
		resume.Contact = entities.Contact{
			Phone: apiItem.Contact.Phone.Formatted,
			Email: apiItem.Contact.Email.Email,
		}
	}

	// Дополнительная информация
	if apiItem.Age != 0 {
		resume.Age = apiItem.Age
	}
	
	if apiItem.Gender.Name != "" {
		resume.Gender = apiItem.Gender.Name
	}

	// Информация о зарплате
	if apiItem.Salary.Amount != 0 {
		resume.Salary = &entities.Salary{
			Amount:   apiItem.Salary.Amount,
			Currency: apiItem.Salary.Currency,
			Gross:    apiItem.Salary.Gross,
		}
	}

	return resume
}

// HHAPIResponse - структура ответа API поиска резюме
type HHAPIResponse struct {
	Items []HHResumeItem `json:"items"` // Список резюме
	Found int            `json:"found"` // Общее количество найденных
	Pages int            `json:"pages"` // Количество страниц
	Page  int            `json:"page"`  // Текущая страница
}

// HHResumeItem - структура элемента резюме из API hh.ru
type HHResumeItem struct {
	ID        string `json:"id"`         // Идентификатор резюме
	Title     string `json:"title"`      // Заголовок резюме  
	FirstName string `json:"first_name"` // Имя
	LastName  string `json:"last_name"`  // Фамилия
	UpdatedAt string `json:"updated_at"` // Дата обновления
	URL       string `json:"url"`        // Ссылка на резюме
	Age       int    `json:"age"`        // Возраст

	// Навыки
	Skills []struct {
		Name string `json:"name"` // Название навыка
	} `json:"skills"`

	// Опыт работы
	Experience []struct {
		Company struct {
			Name string `json:"name"` // Название компании
		} `json:"company"`
		Position    string `json:"position"`    // Должность
		Start       string `json:"start"`       // Дата начала
		End         string `json:"end"`         // Дата окончания  
		Description string `json:"description"` // Описание
	} `json:"experience"`

	// Образование
	Education []struct {
		Name         string `json:"name"`         // Название учебного заведения
		Organization string `json:"organization"` // Факультет/организация
		Result       string `json:"result"`       // Специальность/результат
		Year         int    `json:"year"`         // Год окончания
	} `json:"education"`

	// Контакты
	Contact struct {
		Phone struct {
			Formatted string `json:"formatted"` // Отформатированный номер
		} `json:"phone"`
		Email struct {
			Email string `json:"email"` // Email адрес
		} `json:"email"`
	} `json:"contact"`

	// Пол
	Gender struct {
		Name string `json:"name"` // Пол (мужской/женский)
	} `json:"gender"`

	// Зарплата
	Salary struct {
		Amount   int    `json:"amount"`   // Сумма
		Currency string `json:"currency"` // Валюта
		Gross    bool   `json:"gross"`    // До налогов
	} `json:"salary"`
}
