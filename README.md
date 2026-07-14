# ElenaSite — Система онлайн-записи на услуги

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

**ElenaSite** — это полноценное веб-приложение для онлайн-записи клиентов на услуги (онлайн или оффлайн). Проект включает в себя REST API на Go, клиентскую часть для записи, административную панель и интеграцию с платежной системой ЮKassa (поддержка полной оплаты и предоплаты 50%).

---

## ✨ Основные возможности

- 📅 **Умное бронирование**: Проверка доступных временных слотов с защитой от двойного бронирования (используется расширение PostgreSQL `btree_gist` для исключения пересекающихся интервалов).
- 💳 **Интеграция с ЮKassa**: Создание платежей, обработка вебхуков, поддержка полной оплаты и предоплаты (50%).
- 🔐 **Безопасность и удобство**: Управление записью без обязательной регистрации (используются уникальные `cancellation_tokens` для отмены или переноса записи по ссылке из письма).
- 📧 **Уведомления**: Отправка подтверждений и напоминаний через SMTP (`gomail`).
- 🛠️ **Админ-панель**: Управление списком услуг (название, цена, длительность, формат) и просмотр всех записей.
- 📊 **Аналитика**: Встроенная поддержка Яндекс.Метрики.

---

## 🛠️ Стек технологий

| Компонент | Технологии |
|-----------|------------|
| **Backend** | Go (Golang), Gin, `pgx` / `sqlx`, `golang-migrate` |
| **Database** | PostgreSQL 16 (с расширением `btree_gist`) |
| **Frontend** | Vanilla HTML, CSS, JavaScript |
| **Инфраструктура** | Docker, Docker Compose |
| **Внешние API** | ЮKassa (YooKassa), SMTP-провайдер |

---

## 🚀 Быстрый старт

Для запуска проекта локально вам понадобятся **Docker** и **Docker Compose**.

### 1. Клонирование репозитория
```bash
git clone https://github.com/Artin-cell/elenasite.git
cd elenasite
```

### 2. Настройка переменных окружения
Скопируйте файл с примером и заполните его своими данными:
```bash
cp .env.example .env
```
*(См. раздел [Переменные окружения](#-переменные-окружения) ниже)*

### 3. Запуск через Docker Compose
```bash
docker-compose up --build -d
```
После запуска сервисы будут доступны по адресам:
- **Клиентская часть**: `http://localhost:8080`
- **Админ-панель**: `http://localhost:8080/admin` (или отдельный порт, если настроен)
- **База данных**: `localhost:5432`

---

## ⚙️ Переменные окружения

Создайте файл `.env` в корне проекта на основе `.env.example`:

```env
# --- Приложение ---
PORT=8080
GIN_MODE=release # или 'debug' для разработки
JWT_SECRET=your_super_secret_jwt_key_here

# --- База данных (PostgreSQL) ---
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_db_password
DB_NAME=elena_db
DB_SSL_MODE=disable

# --- ЮKassa (Платежи) ---
YOOKASSA_SHOP_ID=your_shop_id
YOOKASSA_SECRET_KEY=your_test_secret_key
YOOKASSA_RETURN_URL=http://localhost:8080/payment/success
YOOKASSA_WEBHOOK_URL=http://localhost:8080/api/v1/webhooks/yookassa

# --- Почта (SMTP) ---
SMTP_HOST=smtp.yandex.ru
SMTP_PORT=465
SMTP_USER=your_email@yandex.ru
SMTP_PASSWORD=your_app_password
SMTP_FROM_NAME="ElenaSite Booking"

# --- Яндекс.Метрика (опционально) ---
YANDEX_METRICA_ID=your_counter_id
```

---

## 🗄️ Миграции базы данных

Миграции выполняются автоматически при запуске контейнера (если настроен `entrypoint`). 
Для ручного применения миграций используйте `golang-migrate`:

```bash
# Вверх (применить все)
migrate -path migrations -database "postgres://postgres:your_secure_db_password@localhost:5432/elena_db?sslmode=disable" up

# Вниз (откатить последнюю)
migrate -path migrations -database "postgres://postgres:your_secure_db_password@localhost:5432/elena_db?sslmode=disable" down 1
```
> **Важно:** Убедитесь, что в базе данных включено расширение `btree_gist` (оно применяется в первой миграции).

---

## 📂 Структура проекта

```text
.
├── cmd/
│   └── main.go              # Точка входа в приложение
├── internal/
│   ├── config/              # Загрузка конфигурации и .env
│   ├── handlers/            # HTTP-обработчики (Gin)
│   ├── models/              # Структуры данных (Go)
│   ├── repository/          # Работа с БД (pgx/sqlx)
│   ├── services/            # Бизнес-логика (записи, платежи, почта)
│   └── middleware/          # JWT, CORS, логирование
├── migrations/              # SQL-файлы миграций (golang-migrate)
├── public/                  # Статические файлы (HTML, CSS, JS клиентской части)
├── admin-front/             # Исходные файлы административной панели
├── docker-compose.yml       # Конфигурация Docker-сервисов
├── Dockerfile               # Образ для Go-приложения
└── .env.example             # Пример переменных окружения
```

---

## 🧪 Тестирование

Для запуска unit-тестов выполните:
```bash
go test ./... -v
```
*(Рекомендуется настроить CI/CD через GitHub Actions для автоматического прогона тестов и линтера `golangci-lint` при каждом push).*

---

## 🗺️ Roadmap (Планы развития)

- [x] Базовая структура БД и миграции
- [x] Интеграция с ЮKassa (создание платежа)
- [ ] Полная реализация обработки вебхуков ЮKassa
- [ ] Связка клиентского JS с API бэкенда (полный цикл записи)
- [ ] Реализация CRUD для услуг в админ-панели
- [ ] Настройка автоматических email-уведомлений (напоминания за 24ч)
- [ ] Настройка CI/CD и деплой на продакшен (Nginx + Let's Encrypt)

---

## 🤝 Вклад в проект

Pull Request'ы приветствуются! Для серьезных изменений, пожалуйста, сначала откройте Issue, чтобы обсудить детали.

---

## 📄 Лицензия

Этот проект распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для подробностей.

---

## 📬 Контакты

Если у вас есть вопросы или предложения, вы можете открыть Issue в репозитории или связаться с автором: [Artin-cell](https://github.com/Artin-cell).