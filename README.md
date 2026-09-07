# Book Catalog API

Учебный проект книжного каталога на Go.

Проект разрабатывается для сравнения двух подходов к построению API: REST и GraphQL. На текущем этапе реализована REST-часть приложения с PostgreSQL.

## Что уже работает

- CRUD для книг
- CRUD для авторов
- получение пользователей
- списки чтения пользователей
- поиск книг по названию и описанию
- фильтрация по автору
- сортировка
- пагинация
- обработка ошибок API
- проверка связей и ограничений PostgreSQL

## Стек

- Go
- PostgreSQL
- pgx
- REST API
- Postman
- Git

В дальнейшем в проект будут добавлены GraphQL и MongoDB.

## Структура проекта

```text
book-catalog-api/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── apperror/
│   ├── database/
│   ├── domain/
│   ├── repository/
│   │   └── postgres/
│   └── rest/
│
├── migrations/
│   ├── 001_init.sql
│   └── 002_seed.sql
│
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

`cmd/server` содержит точку запуска приложения.

`internal/domain` содержит основные модели данных.

`internal/repository/postgres` отвечает за работу с PostgreSQL.

`internal/rest` содержит HTTP-обработчики REST API.

`internal/database` отвечает за подключение к базе данных.

`internal/apperror` содержит общие ошибки приложения.

В `migrations` находятся SQL-файлы для создания таблиц и добавления тестовых данных.

## База данных

Сейчас проект использует PostgreSQL.

Основные таблицы:

| Таблица | Назначение |
|---|---|
| `authors` | авторы книг |
| `books` | каталог книг |
| `users` | пользователи |
| `reading_list` | книги в списках чтения пользователей |

Связь между авторами и книгами имеет вид `1:N`: у одного автора может быть несколько книг, каждая книга относится к одному автору.

`reading_list` связывает пользователей и книги отношением `M:N`.

При удалении книги или пользователя связанные записи из `reading_list` удаляются автоматически. Удалить автора, у которого существуют книги, нельзя.

## REST API

Сервер по умолчанию запускается на:

```text
http://localhost:8080
```

### Books

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/books` | получить список книг |
| `GET` | `/books/{id}` | получить книгу по ID |
| `POST` | `/books` | добавить книгу |
| `PUT` | `/books/{id}` | изменить книгу |
| `DELETE` | `/books/{id}` | удалить книгу |

Пример создания книги:

```http
POST /books
Content-Type: application/json
```

```json
{
  "author_id": 6,
  "title": "Марсианские хроники",
  "description": "Сборник связанных рассказов Рэя Брэдбери"
}
```

### Authors

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/authors` | получить авторов |
| `GET` | `/authors/{id}` | получить автора и его книги |
| `POST` | `/authors` | добавить автора |
| `PUT` | `/authors/{id}` | изменить автора |
| `DELETE` | `/authors/{id}` | удалить автора |

### Users

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/users` | получить пользователей |
| `GET` | `/users/{id}` | получить пользователя |

### Reading list

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/users/{id}/reading-list` | получить список чтения |
| `POST` | `/users/{id}/reading-list` | добавить книгу в список |
| `DELETE` | `/users/{id}/reading-list/{book_id}` | удалить книгу из списка |

Для добавления книги:

```http
POST /users/2/reading-list
Content-Type: application/json
```

```json
{
  "book_id": 5
}
```

## Поиск и фильтрация

`GET /books` поддерживает несколько query-параметров.

Поиск по названию и описанию:

```http
GET /books?search=война
```

Фильтрация по имени автора:

```http
GET /books?author=Достоевский
```

Фильтрация по ID автора:

```http
GET /books?author_id=1
```

Сортировка:

```http
GET /books?sort=title&order=asc
```

Доступные поля сортировки:

```text
id
title
author
```

Порядок:

```text
asc
desc
```

Пагинация:

```http
GET /books?limit=5&offset=0
```

Параметры можно использовать вместе:

```http
GET /books?author=Достоевский&sort=title&order=asc&limit=5&offset=0
```

## Обработка ошибок

API возвращает HTTP-коды в зависимости от результата запроса.

Например:

```text
400 Bad Request
404 Not Found
409 Conflict
500 Internal Server Error
```

Попытка получить несуществующую книгу:

```http
GET /books/99999
```

вернёт:

```json
{
  "error": "book not found"
}
```

Если попытаться повторно добавить одну книгу в список чтения:

```json
{
  "error": "book is already in reading list"
}
```

При попытке удалить автора, у которого есть книги:

```http
DELETE /authors/1
```

API вернёт:

```text
409 Conflict
```

```json
{
  "error": "author cannot be deleted because they have books"
}
```

## Запуск

Для работы проекта нужны Go и PostgreSQL.

### 1. Клонировать репозиторий

```bash
git clone <repository-url>
cd book-catalog-api
```

### 2. Установить зависимости

```bash
go mod download
```

### 3. Запустить PostgreSQL

При использовании WSL:

```bash
sudo service postgresql start
```

### 4. Создать пользователя и базу

Подключиться к PostgreSQL:

```bash
sudo -u postgres psql
```

Создать пользователя:

```sql
CREATE USER book_user WITH PASSWORD 'your_password';
```

Создать базу:

```sql
CREATE DATABASE book_catalog OWNER book_user;
```

Выйти:

```text
\q
```

### 5. Выполнить миграции

Создание таблиц:

```bash
psql -h localhost -U book_user -d book_catalog -f migrations/001_init.sql
```

Добавление тестовых данных:

```bash
psql -h localhost -U book_user -d book_catalog -f migrations/002_seed.sql
```

### 6. Создать `.env`

В репозитории находится пример конфигурации:

```text
.env.example
```

Создать локальный файл:

```bash
cp .env.example .env
```

Пример:

```env
DATABASE_URL=postgres://book_user:your_password@127.0.0.1:5432/book_catalog?sslmode=disable
SERVER_PORT=8080
```

В `.env` нужно указать пароль пользователя PostgreSQL.

Сам `.env` не хранится в Git.

### 7. Запустить сервер

```bash
go run ./cmd/server
```

При успешном запуске:

```text
Connected to PostgreSQL
Server started on http://localhost:8080
```

Проверить работу сервера:

```http
GET http://localhost:8080/health
```

## Тестирование

REST API проверяется через Postman.

Для базовой проверки можно использовать:

```http
GET http://localhost:8080/books
```

```http
GET http://localhost:8080/authors
```

```http
GET http://localhost:8080/users/2/reading-list
```

После запуска API работает локально на порту, указанном в `SERVER_PORT`.

## Дальше

REST API с PostgreSQL является первой частью проекта.

Следующие этапы:

1. добавить GraphQL API;
2. подключить MongoDB;
3. реализовать авторизацию пользователей;
4. добавить web-интерфейс;
5. подготовить одинаковые сценарии запросов для REST и GraphQL;
6. сравнить полученные реализации;
7. добавить Docker.

## Статус проекта

В разработке.

На данный момент готова основная REST-часть приложения и работа с PostgreSQL. CRUD, поиск, фильтрация, сортировка, пагинация и пользовательские списки чтения работают.
