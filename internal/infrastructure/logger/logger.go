package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel - уровень логирования
type LogLevel int

const (
	DEBUG LogLevel = iota // Отладочная информация
	INFO                  // Информационные сообщения
	WARN                  // Предупреждения
	ERROR                 // Ошибки
	FATAL                 // Критические ошибки
)

// String - строковое представление уровня логирования
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger - интерфейс логгера
// Определяет методы для записи логов различных уровней
type Logger interface {
	// Debug - отладочная информация
	Debug(message string, fields map[string]interface{})
	
	// Info - информационные сообщения
	Info(message string, fields map[string]interface{})
	
	// Warn - предупреждения
	Warn(message string, fields map[string]interface{})
	
	// Error - ошибки
	Error(message string, err error)
	
	// Fatal - критические ошибки (завершают программу)
	Fatal(message string, err error)
	
	// Close - закрытие логгера и освобождение ресурсов
	Close() error
}

// fileLogger - реализация логгера для записи в файл
type fileLogger struct {
	file     *os.File    // Файл для записи логов
	logger   *log.Logger // Встроенный логгер Go
	minLevel LogLevel    // Минимальный уровень для записи
}

// New - создание нового файлового логгера
func New(filename string) (Logger, error) {
	return NewWithLevel(filename, INFO)
}

// NewWithLevel - создание нового файлового логгера с указанным уровнем
func NewWithLevel(filename string, level LogLevel) (Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл логов %s: %w", filename, err)
	}

	logger := log.New(file, "", 0) // Убираем стандартные префиксы

	return &fileLogger{
		file:     file,
		logger:   logger,
		minLevel: level,
	}, nil
}

// Debug - запись отладочной информации
func (l *fileLogger) Debug(message string, fields map[string]interface{}) {
	if l.minLevel <= DEBUG {
		l.writeLog(DEBUG, message, fields, nil)
	}
}

// Info - запись информационного сообщения
func (l *fileLogger) Info(message string, fields map[string]interface{}) {
	if l.minLevel <= INFO {
		l.writeLog(INFO, message, fields, nil)
	}
}

// Warn - запись предупреждения
func (l *fileLogger) Warn(message string, fields map[string]interface{}) {
	if l.minLevel <= WARN {
		l.writeLog(WARN, message, fields, nil)
	}
}

// Error - запись ошибки
func (l *fileLogger) Error(message string, err error) {
	if l.minLevel <= ERROR {
		fields := make(map[string]interface{})
		if err != nil {
			fields["error"] = err.Error()
		}
		l.writeLog(ERROR, message, fields, err)
	}
}

// Fatal - запись критической ошибки
func (l *fileLogger) Fatal(message string, err error) {
	fields := make(map[string]interface{})
	if err != nil {
		fields["error"] = err.Error()
	}
	l.writeLog(FATAL, message, fields, err)
	os.Exit(1)
}

// Close - закрытие файла логов
func (l *fileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// writeLog - внутренний метод для записи лога
func (l *fileLogger) writeLog(level LogLevel, message string, fields map[string]interface{}, err error) {
	// Формирование временной метки
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// Базовая строка лога
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), message)
	
	// Добавление дополнительных полей
	if len(fields) > 0 {
		fieldsStr := ""
		for key, value := range fields {
			if fieldsStr != "" {
				fieldsStr += ", "
			}
			fieldsStr += fmt.Sprintf("%s=%v", key, value)
		}
		logLine += fmt.Sprintf(" | %s", fieldsStr)
	}
	
	// Запись в файл
	l.logger.Println(logLine)
	
	// Дублирование в stdout для важных сообщений
	if level >= WARN {
		fmt.Println(logLine)
	}
}

// consoleLogger - реализация логгера для вывода в консоль
type consoleLogger struct {
	minLevel LogLevel
}

// NewConsole - создание консольного логгера
func NewConsole() Logger {
	return NewConsoleWithLevel(INFO)
}

// NewConsoleWithLevel - создание консольного логгера с указанным уровнем
func NewConsoleWithLevel(level LogLevel) Logger {
	return &consoleLogger{
		minLevel: level,
	}
}

// Debug - запись отладочной информации в консоль
func (l *consoleLogger) Debug(message string, fields map[string]interface{}) {
	if l.minLevel <= DEBUG {
		l.writeConsoleLog(DEBUG, message, fields, nil)
	}
}

// Info - запись информационного сообщения в консоль
func (l *consoleLogger) Info(message string, fields map[string]interface{}) {
	if l.minLevel <= INFO {
		l.writeConsoleLog(INFO, message, fields, nil)
	}
}

// Warn - запись предупреждения в консоль
func (l *consoleLogger) Warn(message string, fields map[string]interface{}) {
	if l.minLevel <= WARN {
		l.writeConsoleLog(WARN, message, fields, nil)
	}
}

// Error - запись ошибки в консоль
func (l *consoleLogger) Error(message string, err error) {
	if l.minLevel <= ERROR {
		fields := make(map[string]interface{})
		if err != nil {
			fields["error"] = err.Error()
		}
		l.writeConsoleLog(ERROR, message, fields, err)
	}
}

// Fatal - запись критической ошибки в консоль
func (l *consoleLogger) Fatal(message string, err error) {
	fields := make(map[string]interface{})
	if err != nil {
		fields["error"] = err.Error()
	}
	l.writeConsoleLog(FATAL, message, fields, err)
	os.Exit(1)
}

// Close - закрытие консольного логгера (ничего не делает)
func (l *consoleLogger) Close() error {
	return nil
}

// writeConsoleLog - внутренний метод для записи лога в консоль
func (l *consoleLogger) writeConsoleLog(level LogLevel, message string, fields map[string]interface{}, err error) {
	timestamp := time.Now().Format("15:04:05")
	
	// Цветовое кодирование для консоли
	var colorCode string
	switch level {
	case DEBUG:
		colorCode = "\033[37m" // Белый
	case INFO:
		colorCode = "\033[36m" // Голубой
	case WARN:
		colorCode = "\033[33m" // Желтый
	case ERROR:
		colorCode = "\033[31m" // Красный
	case FATAL:
		colorCode = "\033[35m" // Пурпурный
	}
	resetColor := "\033[0m"
	
	logLine := fmt.Sprintf("%s[%s] [%s] %s%s", colorCode, timestamp, level.String(), message, resetColor)
	
	if len(fields) > 0 {
		fieldsStr := ""
		for key, value := range fields {
			if fieldsStr != "" {
				fieldsStr += ", "
			}
			fieldsStr += fmt.Sprintf("%s=%v", key, value)
		}
		logLine += fmt.Sprintf(" | %s", fieldsStr)
	}
	
	fmt.Println(logLine)
}
