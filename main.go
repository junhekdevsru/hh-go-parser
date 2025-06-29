package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

)

// Configuration struct for application settings
type Config struct {
	APIToken     string
	Keywords     []string
	City         string
	Experience   string
	UpdateDays   int
	OutputFormat string
	OutputFile   string
	DBConfig     DatabaseConfig
	RateLimit    time.Duration
	LogFile      string
}

// DatabaseConfig holds PostgreSQL connection details
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// Resume represents a parsed resume from hh.ru API
type Resume struct {
	ID          string    `json:"id"`
	Name        string    `json:"name,omitempty"`
	Skills      []string  `json:"skills,omitempty"`
	Experience  []Job     `json:"experience,omitempty"`
	Education   []Edu     `json:"education,omitempty"`
	LastUpdate  time.Time `json:"last_update"`
	Contact     Contact   `json:"contact,omitempty"`
	URL         string    `json:"url,omitempty"`
}

// Job represents work experience
type Job struct {
	Company   string `json:"company,omitempty"`
	Position  string `json:"position,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// Edu represents education
type Edu struct {
	Institution string `json:"institution,omitempty"`
	Faculty     string `json:"faculty,omitempty"`
	Specialty   string `json:"specialty,omitempty"`
	Year        string `json:"year,omitempty"`
}

// Contact represents contact information
type Contact struct {
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

// APIResponse represents the API response structure
type APIResponse struct {
	Items []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		UpdatedAt   string `json:"updated_at"`
		URL         string `json:"url"`
		Experience  []Job  `json:"experience"`
		Education   []Edu  `json:"education"`
		Skills      []struct {
			Name string `json:"name"`
		} `json:"skills"`
		Contact struct {
			Phone struct {
				Formatted string `json:"formatted"`
			} `json:"phone"`
			Email struct {
				Email string `json:"email"`
			} `json:"email"`
		} `json:"contact"`
	} `json:"items"`
	Found int `json:"found"`
	Pages int `json:"pages"`
	Page  int `json:"page"`
}

// Logger handles application logging
type Logger struct {
	file   *os.File
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	
	logger := log.New(file, "", log.LstdFlags)
	return &Logger{file: file, logger: logger}, nil
}

// Log writes a log entry
func (l *Logger) Log(level, message string) {
	l.logger.Printf("[%s] %s", level, message)
}

// Close closes the log file
func (l *Logger) Close() error {
	return l.file.Close()
}

// ResumeParser handles the main parsing logic
type ResumeParser struct {
	config    Config
	logger    *Logger
	client    *http.Client
	processed map[string]bool // Cache for avoiding duplicates
}

// NewResumeParser creates a new parser instance
func NewResumeParser(config Config) (*ResumeParser, error) {
	logger, err := NewLogger(config.LogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	parser := &ResumeParser{
		config:    config,
		logger:    logger,
		client:    &http.Client{Timeout: 30 * time.Second},
		processed: make(map[string]bool),
	}


	return parser, nil
}


// makeRequest makes an HTTP request to hh.ru API with rate limiting
func (rp *ResumeParser) makeRequest(url string) (*APIResponse, error) {
	time.Sleep(rp.config.RateLimit) // Rate limiting

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+rp.config.APIToken)
	req.Header.Set("User-Agent", "Resume Parser v1.0")

	resp, err := rp.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, body)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

// buildSearchURL constructs the search URL for hh.ru API
func (rp *ResumeParser) buildSearchURL(page int) string {
	baseURL := "https://api.hh.ru/resumes"
	params := []string{fmt.Sprintf("page=%d", page)}

	if len(rp.config.Keywords) > 0 {
		params = append(params, "text="+strings.Join(rp.config.Keywords, "+"))
	}

	if rp.config.City != "" {
		// Map city names to area IDs (Moscow = 1)
		cityMap := map[string]string{
			"Moscow":    "1",
			"Москва":    "1",
			"SPb":       "2",
			"СПб":       "2",
			"Ekaterinburg": "3",
		}
		if areaID, exists := cityMap[rp.config.City]; exists {
			params = append(params, "area="+areaID)
		}
	}

	if rp.config.Experience != "" {
		params = append(params, "experience="+rp.config.Experience)
	}

	if rp.config.UpdateDays > 0 {
		params = append(params, "period="+strconv.Itoa(rp.config.UpdateDays))
	}

	if len(params) > 0 {
		return baseURL + "?" + strings.Join(params, "&")
	}
	return baseURL
}

// parseResume converts API response item to Resume struct
func (rp *ResumeParser) parseResume(item interface{}) Resume {
	// Type assertion and parsing logic
	data, _ := json.Marshal(item)
	var apiItem struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		UpdatedAt   string `json:"updated_at"`
		URL         string `json:"url"`
		Experience  []Job  `json:"experience"`
		Education   []Edu  `json:"education"`
		Skills      []struct {
			Name string `json:"name"`
		} `json:"skills"`
		Contact struct {
			Phone struct {
				Formatted string `json:"formatted"`
			} `json:"phone"`
			Email struct {
				Email string `json:"email"`
			} `json:"email"`
		} `json:"contact"`
	}
	json.Unmarshal(data, &apiItem)

	resume := Resume{
		ID:         apiItem.ID,
		Name:       strings.TrimSpace(apiItem.FirstName + " " + apiItem.LastName),
		Experience: apiItem.Experience,
		Education:  apiItem.Education,
		URL:        apiItem.URL,
	}

	// Parse skills
	for _, skill := range apiItem.Skills {
		resume.Skills = append(resume.Skills, skill.Name)
	}

	// Parse contact information
	resume.Contact = Contact{
		Phone: apiItem.Contact.Phone.Formatted,
		Email: apiItem.Contact.Email.Email,
	}

	// Parse update date
	if apiItem.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339, apiItem.UpdatedAt); err == nil {
			resume.LastUpdate = t
		}
	}

	return resume
}

// saveToCSV saves resumes to CSV file
func (rp *ResumeParser) saveToCSV(resumes []Resume, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Name", "Skills", "Experience", "Education", "Last Update", "Phone", "Email", "URL"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, resume := range resumes {
		record := []string{
			resume.ID,
			resume.Name,
			strings.Join(resume.Skills, "; "),
			rp.formatExperience(resume.Experience),
			rp.formatEducation(resume.Education),
			resume.LastUpdate.Format("2006-01-02 15:04:05"),
			resume.Contact.Phone,
			resume.Contact.Email,
			resume.URL,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// saveToJSON saves resumes to JSON file
func (rp *ResumeParser) saveToJSON(resumes []Resume, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resumes)
}

// saveToPostgreSQL generates SQL script for PostgreSQL
func (rp *ResumeParser) saveToPostgreSQL(resumes []Resume) error {
	filename := "resumes.sql"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write table creation scripts
	file.WriteString(`-- Resume Parser SQL Export
-- Execute this script in your PostgreSQL database

`)
	file.WriteString(`CREATE TABLE IF NOT EXISTS resumes (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500),
    skills TEXT,
    last_update TIMESTAMP,
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),
    url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

`)
	file.WriteString(`CREATE TABLE IF NOT EXISTS experience (
    id SERIAL PRIMARY KEY,
    resume_id VARCHAR(255) REFERENCES resumes(id),
    company VARCHAR(500),
    position VARCHAR(500),
    start_date VARCHAR(50),
    end_date VARCHAR(50)
);

`)
	file.WriteString(`CREATE TABLE IF NOT EXISTS education (
    id SERIAL PRIMARY KEY,
    resume_id VARCHAR(255) REFERENCES resumes(id),
    institution VARCHAR(500),
    faculty VARCHAR(500),
    specialty VARCHAR(500),
    year VARCHAR(50)
);

`)

	// Write data insertion scripts
	for _, resume := range resumes {
		// Escape single quotes for SQL
		escapedName := strings.ReplaceAll(resume.Name, "'", "''")
		escapedSkills := strings.ReplaceAll(strings.Join(resume.Skills, "; "), "'", "''")
		escapedPhone := strings.ReplaceAll(resume.Contact.Phone, "'", "''")
		escapedEmail := strings.ReplaceAll(resume.Contact.Email, "'", "''")
		escapedURL := strings.ReplaceAll(resume.URL, "'", "''")

		fmt.Fprintf(file, "INSERT INTO resumes (id, name, skills, last_update, contact_phone, contact_email, url) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s') ON CONFLICT (id) DO NOTHING;\n",
			resume.ID, escapedName, escapedSkills, resume.LastUpdate.Format("2006-01-02 15:04:05"), escapedPhone, escapedEmail, escapedURL)

		// Insert experience
		for _, exp := range resume.Experience {
			escapedCompany := strings.ReplaceAll(exp.Company, "'", "''")
			escapedPosition := strings.ReplaceAll(exp.Position, "'", "''")
			fmt.Fprintf(file, "INSERT INTO experience (resume_id, company, position, start_date, end_date) VALUES ('%s', '%s', '%s', '%s', '%s');\n",
				resume.ID, escapedCompany, escapedPosition, exp.StartDate, exp.EndDate)
		}

		// Insert education
		for _, edu := range resume.Education {
			escapedInstitution := strings.ReplaceAll(edu.Institution, "'", "''")
			escapedFaculty := strings.ReplaceAll(edu.Faculty, "'", "''")
			escapedSpecialty := strings.ReplaceAll(edu.Specialty, "'", "''")
			fmt.Fprintf(file, "INSERT INTO education (resume_id, institution, faculty, specialty, year) VALUES ('%s', '%s', '%s', '%s', '%s');\n",
				resume.ID, escapedInstitution, escapedFaculty, escapedSpecialty, edu.Year)
		}
	}

	rp.logger.Log("INFO", fmt.Sprintf("SQL script saved to %s", filename))
	return nil
}

// formatExperience formats experience for CSV output
func (rp *ResumeParser) formatExperience(experience []Job) string {
	var parts []string
	for _, exp := range experience {
		part := fmt.Sprintf("%s at %s (%s - %s)", exp.Position, exp.Company, exp.StartDate, exp.EndDate)
		parts = append(parts, part)
	}
	return strings.Join(parts, "; ")
}

// formatEducation formats education for CSV output
func (rp *ResumeParser) formatEducation(education []Edu) string {
	var parts []string
	for _, edu := range education {
		part := fmt.Sprintf("%s, %s (%s)", edu.Institution, edu.Specialty, edu.Year)
		parts = append(parts, part)
	}
	return strings.Join(parts, "; ")
}

// loadKeywordsFromFile loads keywords from a text file
func loadKeywordsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keywords []string
	scanner := json.NewDecoder(file)
	if err := scanner.Decode(&keywords); err != nil {
		// Try reading as plain text (one keyword per line)
		file.Seek(0, 0)
		content, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		keywords = strings.Split(strings.TrimSpace(string(content)), "\n")
	}

	return keywords, nil
}

// Parse executes the main parsing logic
func (rp *ResumeParser) Parse() error {
	rp.logger.Log("INFO", "Starting resume parsing...")

	var allResumes []Resume
	page := 0
	totalFound := 0

	for {
		url := rp.buildSearchURL(page)
		rp.logger.Log("INFO", fmt.Sprintf("Fetching page %d: %s", page, url))

		resp, err := rp.makeRequest(url)
		if err != nil {
			rp.logger.Log("ERROR", fmt.Sprintf("Failed to fetch page %d: %v", page, err))
			break
		}

		if page == 0 {
			totalFound = resp.Found
			rp.logger.Log("INFO", fmt.Sprintf("Total resumes found: %d", totalFound))
		}

		for _, item := range resp.Items {
			resume := rp.parseResume(item)
			
			// Check for duplicates
			if rp.processed[resume.ID] {
				continue
			}
			rp.processed[resume.ID] = true

			allResumes = append(allResumes, resume)
		}

		rp.logger.Log("INFO", fmt.Sprintf("Processed page %d, collected %d resumes so far", page, len(allResumes)))

		// Check if we've reached the last page
		if page >= resp.Pages-1 || len(resp.Items) == 0 {
			break
		}

		page++
	}

	rp.logger.Log("INFO", fmt.Sprintf("Finished parsing. Total resumes collected: %d", len(allResumes)))

	// Save results
	return rp.saveResults(allResumes)
}

// saveResults saves the parsed resumes to the specified output format
func (rp *ResumeParser) saveResults(resumes []Resume) error {
	switch rp.config.OutputFormat {
	case "csv":
		return rp.saveToCSV(resumes, rp.config.OutputFile)
	case "json":
		return rp.saveToJSON(resumes, rp.config.OutputFile)
	case "postgres", "sql":
		return rp.saveToPostgreSQL(resumes)
	default:
		return fmt.Errorf("unsupported output format: %s", rp.config.OutputFormat)
	}
}

// Close closes the parser and cleans up resources
func (rp *ResumeParser) Close() error {
	return rp.logger.Close()
}

func main() {
	// Command line flags
	var (
		apiToken     = flag.String("token", "", "hh.ru API token")
		keywords     = flag.String("keywords", "", "Search keywords (comma-separated)")
		keywordsFile = flag.String("keywords-file", "", "File with keywords (JSON array or newline-separated)")
		city         = flag.String("city", "Moscow", "City to search in")
		experience   = flag.String("experience", "", "Experience level (noExperience, between1And3, between3And6, moreThan6)")
		updateDays   = flag.Int("update-days", 7, "Filter by last update days")
		outputFormat = flag.String("format", "json", "Output format (csv, json, sql)")
		outputFile   = flag.String("output", "resumes.json", "Output file (for csv/json)")
		logFile      = flag.String("log", "parser.log", "Log file")
		rateLimit    = flag.Duration("rate", time.Second, "Rate limit between requests")
		
		// Database flags
		dbHost     = flag.String("db-host", "localhost", "PostgreSQL host")
		dbPort     = flag.Int("db-port", 5432, "PostgreSQL port")
		dbUser     = flag.String("db-user", "postgres", "PostgreSQL user")
		dbPassword = flag.String("db-password", "", "PostgreSQL password")
		dbName     = flag.String("db-name", "resumes", "PostgreSQL database name")
	)
	flag.Parse()

	// Validate required parameters
	if *apiToken == "" {
		log.Fatal("API token is required. Use -token flag or set HH_API_TOKEN environment variable")
	}

	// Load keywords
	var keywordList []string
	if *keywordsFile != "" {
		var err error
		keywordList, err = loadKeywordsFromFile(*keywordsFile)
		if err != nil {
			log.Fatalf("Failed to load keywords from file: %v", err)
		}
	} else if *keywords != "" {
		keywordList = strings.Split(*keywords, ",")
	}

	// Create configuration
	config := Config{
		APIToken:     *apiToken,
		Keywords:     keywordList,
		City:         *city,
		Experience:   *experience,
		UpdateDays:   *updateDays,
		OutputFormat: *outputFormat,
		OutputFile:   *outputFile,
		RateLimit:    *rateLimit,
		LogFile:      *logFile,
		DBConfig: DatabaseConfig{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Password: *dbPassword,
			DBName:   *dbName,
		},
	}

	// Create and run parser
	parser, err := NewResumeParser(config)
	if err != nil {
		log.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close()

	if err := parser.Parse(); err != nil {
		log.Fatalf("Parsing failed: %v", err)
	}

	fmt.Println("Resume parsing completed successfully!")
}
