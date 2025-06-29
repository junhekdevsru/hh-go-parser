# Парсер резюме HH.ru

Комплексное Go-приложение для парсинга резюме из API hh.ru с расширенной фильтрацией, несколькими форматами вывода и корпоративными функциями.

## Возможности

- **Интеграция с API**: OAuth2/API аутентификация с API hh.ru
- **Расширенная фильтрация**: Ключевые слова, города, уровни опыта, даты обновления
- **Несколько форматов вывода**: CSV, JSON, скрипты PostgreSQL
- **Ограничение запросов**: Настраиваемое ограничение скорости (по умолчанию: 1 запрос/сек)
- **Логирование**: Комплексное отслеживание ошибок и мониторинг прогресса
- **Предотвращение дубликатов**: Кэширование ID резюме
- **CLI интерфейс**: Гибкие аргументы командной строки
- **Только стандартная библиотека**: Нет внешних зависимостей, кроме драйвера PostgreSQL (опционально)

## Требования

- Go 1.21 или выше
- API токен или OAuth2 учетные данные hh.ru
- PostgreSQL (при использовании SQL формата вывода)

## Установка

1. Клонировать или скачать исходный код:
```bash
git clone <repository-url>
cd hh-resume-parser
```

2. Инициализировать модуль Go:
```bash
go mod init hh-resume-parser
go mod tidy
```

3. Собрать приложение:
```bash
go build -o hh-parser main.go
```

## Аутентификация API

### Получение API токена hh.ru

1. Зарегистрируйтесь на [dev.hh.ru](https://dev.hh.ru/)
2. Создайте новое приложение
3. Получите свой API токен или учетные данные OAuth2
4. Установите переменную окружения (необязательно):
```bash
export HH_API_TOKEN="your_token_here"
```

## Использование

### Примеры

**Парсинг резюме разработчиков Go в Москве:**
```bash
./hh-parser -token="YOUR_TOKEN" -keywords="Go developer" -city="Moscow" -format="json"
```

**Поиск с фильтром по опыту и пользовательским выводом:**
```bash
./hh-parser -token="YOUR_TOKEN" -keywords="golang,backend" -experience="between1And3" -output="results.csv" -format="csv"
```

**Генерация скрипта PostgreSQL:**
```bash
./hh-parser -token="YOUR_TOKEN" -keywords="Go,Golang" -format="sql"
```

### Расширенное использование

**Загрузка ключевых слов из файла:**
```bash
echo '["Go", "Golang", "Backend", "API"]' > keywords.json
./hh-parser -token="YOUR_TOKEN" -keywords-file="keywords.json" -format="json"
```

**Пользовательское ограничение скорости и логирование:**
```bash
./hh-parser -token="YOUR_TOKEN" -keywords="Go" -rate="2s" -log="custom.log" -update-days="3"
```

## Параметры командной строки

### Обязательные
- `-token string`: hh.ru API токен

### Фильтры поиска
- `-keywords string`: Ключевые слова для поиска (разделенные запятыми)
- `-keywords-file string`: Файл с ключевыми словами (JSON массив или разделенные новой строкой)
- `-city string`: Город для поиска (по умолчанию: "Москва")
- `-experience string`: Уровень опыта:
  - `noExperience`: Без опыта
  - `between1And3`: 1-3 года
  - `between3And6`: 3-6 лет  
  - `moreThan6`: Более 6 лет
- `-update-days int`: Фильтр по дням последнего обновления (по умолчанию: 7)

### Параметры вывода
- `-format string`: Формат вывода - csv, json, sql (по умолчанию: "json")
- `-output string`: Файл вывода для csv/json (по умолчанию: "resumes.json")

### Системные параметры
- `-rate duration`: Ограничение скорости между запросами (по умолчанию: 1s)
- `-log string`: Путь к файлу журнала (по умолчанию: "parser.log")

### Параметры базы данных (для SQL формата)
- `-db-host string`: Хост PostgreSQL (по умолчанию: "localhost")
- `-db-port int`: Порт PostgreSQL (по умолчанию: 5432)
- `-db-user string`: Пользователь PostgreSQL (по умолчанию: "postgres")
- `-db-password string`: Пароль PostgreSQL
- `-db-name string`: Имя базы данных (по умолчанию: "resumes")

## Форматы вывода

### Формат JSON
```json
[
  {
    "id": "12345",
    "name": "John Doe",
    "skills": ["Go", "Docker", "Kubernetes"],
    "experience": [
      {
        "company": "Tech Corp",
        "position": "Backend Developer",
        "start_date": "2022-01",
        "end_date": "2024-01"
      }
    ],
    "education": [
      {
        "institution": "Moscow State University",
        "specialty": "Computer Science",
        "year": "2021"
      }
    ],
    "last_update": "2024-01-15T10:30:00Z",
    "contact": {
      "phone": "+7-xxx-xxx-xxxx",
      "email": "john@example.com"
    },
    "url": "https://hh.ru/resume/12345"
  }
]
```

### Формат CSV
Содержит столбцы: ID, Имя, Навыки, Опыт, Образование, Последнее обновление, Телефон, Email, URL

### Формат SQL
Генерирует скрипт, совместимый с PostgreSQL, с:
- Командами создания таблиц
- Командами INSERT с данными
- Правильным экранированием SQL

## Формат файла ключевых слов

### JSON массив:
```json
["Go", "Golang", "Backend", "API", "Docker", "Kubernetes"]
```

### Обычный текст (разделенный новой строкой):
```
Go
Golang
Backend Developer
API Development
Microservices
```

## Логирование

Приложение создает подробные журналы, включая:
- Информацию о запросах/ответах
- Отслеживание прогресса
- Подробности об ошибках
- Статистику обработки

Пример вывода журнала:
```
2024/01/15 10:30:00 [INFO] Начало парсинга резюме...
2024/01/15 10:30:01 [INFO] Получение страницы 0: https://api.hh.ru/resumes?page=0&text=Go+developer&area=1
2024/01/15 10:30:02 [INFO] Всего найдено резюме: 1250
2024/01/15 10:30:02 [INFO] Обработана страница 0, собрано 20 резюме
2024/01/15 10:30:04 [INFO] Завершено парсинг. Всего собрано резюме: 156
```

## Схема базы данных

При использовании SQL формата вывода создаются следующие таблицы:

```sql
CREATE TABLE resumes (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500),
    skills TEXT,
    last_update TIMESTAMP,
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),
    url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE experience (
    id SERIAL PRIMARY KEY,
    resume_id VARCHAR(255) REFERENCES resumes(id),
    company VARCHAR(500),
    position VARCHAR(500),
    start_date VARCHAR(50),
    end_date VARCHAR(50)
);

CREATE TABLE education (
    id SERIAL PRIMARY KEY,
    resume_id VARCHAR(255) REFERENCES resumes(id),
    institution VARCHAR(500),
    faculty VARCHAR(500),
    specialty VARCHAR(500),
    year VARCHAR(50)
);
```

## Ограничение запросов

Приложение включает встроенное ограничение скорости для соблюдения лимитов API hh.ru:
- По умолчанию: 1 запрос в секунду
- Настраивается с помощью флага `-rate`
- Автоматическая логика повторной попытки для ошибок ограничения скорости

## Обработка ошибок

- Тайм-ауты сети и повторные попытки
- Соблюдение ограничения скорости API
- Обработка недопустимого JSON ответа
- Управление ошибками ввода/вывода файлов
- Комплексное логирование ошибок

## Разработка

### Запуск тестов
```bash
go test ./...
```

### Структура кода
- `main.go`: Основное приложение со всей функциональностью
- `go.mod`: Определение модуля
- Только стандартная библиотека (без внешних зависимостей)

### Поддержка Docker

Создайте `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o hh-parser main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/hh-parser .
CMD ["./hh-parser"]
```

Собрать и запустить:
```bash
docker build -t hh-parser .
docker run -v $(pwd):/data hh-parser -token="YOUR_TOKEN" -keywords="Go" -output="/data/resumes.json"
```

## Планирование с помощью Cron

Добавьте в crontab для ежедневного выполнения:
```bash
# Ежедневно в 2 часа ночи
0 2 * * * /path/to/hh-parser -token="YOUR_TOKEN" -keywords="Go developer" -format="json" -output="/data/daily-$(date +\%Y\%m\%d).json"
```

## Используемые API конечные точки

- `GET /resumes` - Поиск резюме
- Параметры: `text`, `area`, `experience`, `period`, `page`
- Лимит запросов: 1000 запросов в час с одного IP

## Устранение неполадок

### Распространенные проблемы

1. **Проблемы с токеном API**
   - Убедитесь, что токен действителен и не истек
   - Проверьте разрешения и квоты API

2. **Ограничение скорости**
   - Увеличьте задержку с помощью флага `-rate`
   - Мониторинг использования квоты API

3. **Сетевые проблемы**
   - Проверьте подключение к интернету
   - Убедитесь, что API hh.ru работает

4. **Большие объемы результатов**
   - Используйте более специфичные ключевые слова
   - Добавьте фильтры по уровню опыта
   - Ограничьте диапазон дней обновления

### Режим отладки
Добавьте подробное логирование, проверив файл журнала на наличие подробной информации о запросах/ответах.

## Участие

1. Форкните репозиторий
2. Создайте ветку функции
3. Добавьте тесты для новой функциональности
4. Отправьте запрос на извлечение

## Лицензия

MIT License - см. файл LICENSE для подробностей

## Поддержка

По вопросам и проблемам:
- Проверьте файл журнала на наличие подробной информации об ошибках
- Убедитесь, что токен API действителен и имеет необходимые разрешения
- Ознакомьтесь с документацией API hh.ru
- Проверьте подключение к сети и ограничения скорости
