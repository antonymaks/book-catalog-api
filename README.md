# Book Catalog API

Учебный проект книжного каталога на Go.

Проект разрабатывается для сравнения двух подходов к построению API: REST и GraphQL. На текущем этапе реализованы обе API-ветки с PostgreSQL. Они используют общие репозитории и предоставляют одинаковый основной функционал.

## Что уже работает

* CRUD для книг и авторов через REST и GraphQL
* получение пользователей
* списки чтения пользователей
* поиск книг по названию и описанию
* фильтрация по автору
* сортировка и пагинация
* вложенные запросы GraphQL
* обработка ошибок API
* проверка связей и ограничений PostgreSQL
* простой web-интерфейс для работы с каталогом

## Стек

* Go
* PostgreSQL
* pgx
* REST API
* GraphQL (gqlgen)
* HTML, CSS, JavaScript
* Postman
* Git

В дальнейшем планируется подключение MongoDB, авторизация и сравнительное тестирование.

## Структура проекта

```text
book-catalog-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── apperror/
│   ├── database/
│   ├── domain/
│   ├── repository/
│   │   └── postgres/
│   └── rest/
├── graph/
│   ├── model/
│   ├── converter.go
│   ├── generated.go
│   ├── resolver.go
│   ├── schema.graphqls
│   └── schema.resolvers.go
├── web/
│   ├── index.html
│   ├── style.css
│   └── app.js
├── migrations/
│   ├── 001_init.sql
│   └── 002_seed.sql
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

`graph` содержит GraphQL-схему, resolver'ы, сгенерированный код и преобразование моделей.

`web` содержит простой интерфейс каталога.

`internal/database` отвечает за подключение к базе данных.

`internal/apperror` содержит общие ошибки приложения.

В `migrations` находятся SQL-файлы для создания таблиц и добавления тестовых данных.

## База данных

Сейчас проект использует PostgreSQL.

Основные таблицы:

| Таблица        | Назначение                           |
| -------------- | ------------------------------------ |
| `authors`      | авторы книг                          |
| `books`        | каталог книг                         |
| `users`        | пользователи                         |
| `reading_list` | книги в списках чтения пользователей |

Связь между авторами и книгами имеет вид `1:N`: у одного автора может быть несколько книг, каждая книга относится к одному автору.

`reading_list` связывает пользователей и книги отношением `M:N`.

При удалении книги или пользователя связанные записи из `reading_list` удаляются автоматически. Удалить автора, у которого существуют книги, нельзя.

## REST API

Сервер по умолчанию запускается на `http://localhost:8080`.

### Books

| Метод    | Endpoint      | Описание             |
| -------- | ------------- | -------------------- |
| `GET`    | `/books`      | получить список книг |
| `GET`    | `/books/{id}` | получить книгу по ID |
| `POST`   | `/books`      | добавить книгу       |
| `PUT`    | `/books/{id}` | изменить книгу       |
| `DELETE` | `/books/{id}` | удалить книгу        |

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

| Метод    | Endpoint        | Описание                    |
| -------- | --------------- | --------------------------- |
| `GET`    | `/authors`      | получить авторов            |
| `GET`    | `/authors/{id}` | получить автора и его книги |
| `POST`   | `/authors`      | добавить автора             |
| `PUT`    | `/authors/{id}` | изменить автора             |
| `DELETE` | `/authors/{id}` | удалить автора              |

### Users

| Метод | Endpoint      | Описание               |
| ----- | ------------- | ---------------------- |
| `GET` | `/users`      | получить пользователей |
| `GET` | `/users/{id}` | получить пользователя  |

### Reading list

| Метод    | Endpoint                             | Описание                |
| -------- | ------------------------------------ | ----------------------- |
| `GET`    | `/users/{id}/reading-list`           | получить список чтения  |
| `POST`   | `/users/{id}/reading-list`           | добавить книгу в список |
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

Доступные поля сортировки: `id`, `title`, `author`.

Порядок: `asc`, `desc`.

Пагинация:

```http
GET /books?limit=5&offset=0
```

Параметры можно использовать вместе:

```http
GET /books?author=Достоевский&sort=title&order=asc&limit=5&offset=0
```

## GraphQL API

GraphQL доступен по адресу:

```text
http://localhost:8080/graphql
```

Для проверки запросов используется GraphQL Playground:

```text
http://localhost:8080/playground
```

GraphQL использует те же PostgreSQL-репозитории, что и REST. Это позволяет сравнивать два подхода к API на одинаковых данных и операциях.

### Получение книг

```graphql
query {
  books {
    id
    title
    description
    author {
      id
      name
    }
  }
}
```

Получение одной книги:

```graphql
query {
  book(id: "1") {
    id
    title
    description
    author {
      name
    }
  }
}
```

### Поиск, фильтрация и пагинация

```graphql
query {
  books(
    filter: {
      author: "Достоевский"
      sort: "title"
      order: "asc"
      limit: 5
      offset: 0
    }
  ) {
    id
    title
    author {
      name
    }
  }
}
```

Поддерживаются параметры `search`, `author`, `authorId`, `sort`, `order`, `limit` и `offset`.

### Авторы и вложенные данные

```graphql
query {
  author(id: "1") {
    id
    name
    books {
      id
      title
    }
  }
}
```

GraphQL позволяет выбирать только необходимые поля и получать связанные данные в одном запросе.

### Пользователи

```graphql
query {
  users {
    id
    username
    role
  }
}
```

```graphql
query {
  user(id: "2") {
    id
    username
    role
    readingList {
      id
      title
      author {
        name
      }
    }
  }
}
```

### CRUD книг

Создание:

```graphql
mutation {
  createBook(
    input: {
      authorId: "6"
      title: "Марсианские хроники"
      description: "Сборник рассказов Рэя Брэдбери"
    }
  ) {
    id
    title
    author {
      name
    }
  }
}
```

Изменение:

```graphql
mutation {
  updateBook(
    id: "13"
    input: {
      authorId: "6"
      title: "Марсианские хроники"
      description: "Обновлённое описание"
    }
  ) {
    id
    title
  }
}
```

Удаление:

```graphql
mutation {
  deleteBook(id: "13")
}
```

### CRUD авторов

```graphql
mutation {
  createAuthor(input: { name: "Стивен Кинг" }) {
    id
    name
  }
}
```

```graphql
mutation {
  updateAuthor(
    id: "9"
    input: { name: "Stephen King" }
  ) {
    id
    name
  }
}
```

```graphql
mutation {
  deleteAuthor(id: "9")
}
```

### Reading list

Получение списка чтения:

```graphql
query {
  readingList(userId: "2") {
    id
    title
    author {
      name
    }
  }
}
```

Добавление книги:

```graphql
mutation {
  addBookToReadingList(userId: "2", bookId: "5")
}
```

Удаление книги:

```graphql
mutation {
  removeBookFromReadingList(userId: "2", bookId: "5")
}
```

## Паритет REST и GraphQL

На текущем этапе оба API предоставляют одинаковые основные возможности.

| Возможность               | REST | GraphQL |
| ------------------------- | ---- | ------- |
| Список книг               | ✓    | ✓       |
| Книга по ID               | ✓    | ✓       |
| Поиск                     | ✓    | ✓       |
| Фильтрация                | ✓    | ✓       |
| Сортировка                | ✓    | ✓       |
| Пагинация                 | ✓    | ✓       |
| CRUD книг                 | ✓    | ✓       |
| Список авторов            | ✓    | ✓       |
| Автор по ID               | ✓    | ✓       |
| CRUD авторов              | ✓    | ✓       |
| Список пользователей      | ✓    | ✓       |
| Пользователь по ID        | ✓    | ✓       |
| Получение reading list    | ✓    | ✓       |
| Добавление в reading list | ✓    | ✓       |
| Удаление из reading list  | ✓    | ✓       |

Оба API работают с одной базой данных. Различается способ организации запросов и получения данных, что будет использоваться при дальнейшем сравнении.

## Обработка ошибок

REST API возвращает HTTP-коды в зависимости от результата запроса:

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

API вернёт `409 Conflict`:

```json
{
  "error": "author cannot be deleted because they have books"
}
```

В GraphQL ошибки возвращаются в поле `errors` ответа. Например, при попытке удалить автора с книгами:

```graphql
mutation {
  deleteAuthor(id: "1")
}
```

будет возвращена ошибка:

```text
author cannot be deleted because they have books
```

## Web-интерфейс

В проекте есть простой web-интерфейс для работы с каталогом. Он доступен по адресу:

```text
http://localhost:8080/
```

Через интерфейс можно просматривать книги, использовать поиск и фильтрацию, а также работать со списком чтения. На текущем этапе frontend использует REST API. В дальнейшем планируется добавить переключение между REST и GraphQL.

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

В репозитории находится пример конфигурации `.env.example`.

Создать локальный файл:

```bash
cp .env.example .env
```

Пример:

```env
DATABASE_URL=postgres://book_user:your_password@127.0.0.1:5432/book_catalog?sslmode=disable
SERVER_PORT=8080
```

В `.env` нужно указать пароль пользователя PostgreSQL. Сам `.env` не хранится в Git.

### 7. Запустить сервер

```bash
go run ./cmd/server
```

Проверить работу сервера:

```http
GET http://localhost:8080/health
```

После запуска доступны REST API, GraphQL, Playground и web-интерфейс.

## Тестирование

REST API проверяется через Postman, GraphQL через Playground.

Для базовой проверки REST:

```http
GET http://localhost:8080/books
```

```http
GET http://localhost:8080/authors
```

```http
GET http://localhost:8080/users/2/reading-list
```

Для GraphQL можно выполнить запрос:

```graphql
query {
  books(filter: { limit: 5 }) {
    id
    title
    author {
      name
    }
  }
}
```

Также проверены операции создания, изменения и удаления книг и авторов, работа со списками чтения и обработка ошибок.

## Дальше

Основной функционал REST и GraphQL с PostgreSQL реализован.

Следующие этапы:

1. Добавить переключение REST/GraphQL в web-интерфейсе.
2. Вынести общие интерфейсы репозиториев и сервисный слой.
3. Подключить MongoDB и реализовать аналогичные операции.
4. Добавить выбор базы данных через конфигурацию.
5. Реализовать авторизацию пользователей и роли.
6. Подготовить одинаковые сценарии для сравнения REST и GraphQL.
7. Провести измерения и сравнить полученные результаты.
8. Добавить Docker и итоговую документацию.

## Статус проекта

В разработке.

На данный момент реализованы REST и GraphQL API с PostgreSQL. Основной функционал обеих API-веток приведён к паритету. CRUD, поиск, фильтрация, сортировка, пагинация, пользователи и списки чтения работают. Следующий этап посвящён подготовке архитектуры к подключению MongoDB и дальнейшему сравнению реализаций.
