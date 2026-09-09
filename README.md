# Book Catalog API

Учебный проект книжного каталога на Go для сравнения REST и GraphQL, а также PostgreSQL и MongoDB.

На текущем этапе REST и GraphQL реализованы для PostgreSQL. API использует общие интерфейсы репозиториев. MongoDB запущена в Docker и заполнена подготовленным набором из 100 книг. Подключение MongoDB как полноценного хранилища API является следующим этапом.

## Возможности

- CRUD книг и авторов через REST и GraphQL.
- Получение пользователей и управление списками чтения.
- Поиск по названию и описанию, фильтрация по автору.
- Сортировка и пагинация.
- Получение связанных данных через GraphQL.
- Обработка ошибок и проверка внешних связей.
- Простой web-интерфейс, работающий через REST.
- Общие repository-интерфейсы для независимости API от конкретной БД.
- Подготовка и импорт единого датасета в PostgreSQL и MongoDB.

## Стек

Go, PostgreSQL, MongoDB, Docker Compose, pgx, MongoDB Go Driver v2, gqlgen, HTML, CSS, JavaScript, Postman, Git.

## Архитектура

```text
REST ─────────┐
              ├── Repository interfaces
GraphQL ──────┘           │
                         ├── PostgreSQL (реализовано)
                         └── MongoDB (репозитории в разработке)
```

REST-хендлеры и GraphQL resolver'ы зависят от интерфейсов `BookRepository`, `AuthorRepository` и `UserRepository`, а не от конкретных PostgreSQL-структур. Это позволит подключить MongoDB без переписывания API-слоя. Отдельный service layer пока не добавлялся: общую бизнес-логику планируется выделять по мере необходимости.

### Структура проекта

```text
book-catalog-api/
├── cmd/
│   ├── server/
│   └── importer/
│       ├── main.go
│       ├── load_postgres.go
│       └── load_mongo.go
├── data/
│   ├── raw/                 # исходные данные, не хранятся в Git
│   └── prepared/
│       └── books.csv
├── internal/
│   ├── apperror/
│   ├── database/
│   ├── domain/
│   ├── repository/
│   │   ├── repository.go
│   │   └── postgres/
│   └── rest/
├── graph/
├── migrations/
├── web/
├── docker-compose.yml
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── gqlgen.yml
└── README.md
```

Файлы импортёров являются отдельными исполняемыми программами. Их следует запускать по имени файла, а не командой `go run ./cmd/importer`, поскольку в каталоге может находиться несколько функций `main`.

## Модель данных

### PostgreSQL

| Таблица | Назначение |
| --- | --- |
| `authors` | Авторы книг |
| `books` | Книги и ссылки на авторов |
| `users` | Пользователи, роли и хеши паролей |
| `reading_list` | Связь пользователей с книгами |

Один автор может иметь много книг, но каждая книга относится к одному автору. Список чтения реализует связь многие-ко-многим между пользователями и книгами. Удаление книги или пользователя удаляет связанные записи списка чтения. Удаление автора, у которого есть книги, ограничено внешним ключом.

### MongoDB

Используются коллекции `authors` и `books`. Для сопоставимости с PostgreSQL у документов сохранены числовые идентификаторы и поле `author_id`.

```json
{
  "id": 1,
  "name": "Author Name"
}
```

```json
{
  "id": 1,
  "author_id": 1,
  "title": "Book title",
  "description": "Book description."
}
```

MongoDB автоматически добавляет служебное поле `_id`. Числовое `id` используется как идентификатор предметной области. В дальнейшем будут добавлены коллекции пользователей и списков чтения, а также индексы и репозитории для полного API-паритета.

## Датасет

Первоначальные 12 тестовых книг заменены подготовленным набором из 100 книг на основе [CMU Book Summary Dataset](https://www.kaggle.com/datasets/ymaricar/cmu-book-summary-dataset). Исходный набор содержит названия, авторов и сюжетные описания.

Подготовка включает выбор необходимых полей, удаление записей с пустыми значениями, устранение дублей и сокращение описания до первого предложения. Результат сохраняется в `data/prepared/books.csv` с колонками:

```csv
title,author,description
```

Один и тот же подготовленный файл используется для обеих БД. Это позволяет сравнивать их на одинаковых исходных данных. Для строгого сравнения необходимо также сверять идентификаторы и связи, поскольку независимые импортёры могут назначить авторам разные ID.

Исходный датасет хранится в `data/raw/` и не добавляется в Git. Подготовленный CSV можно хранить в репозитории как воспроизводимый тестовый набор с учётом условий лицензии исходных данных.

## REST API

Базовый адрес: `http://localhost:8080`.

### Книги

| Метод | Endpoint | Описание |
| --- | --- | --- |
| GET | `/books` | Список книг |
| GET | `/books/{id}` | Книга по ID |
| POST | `/books` | Создание книги |
| PUT | `/books/{id}` | Обновление книги |
| DELETE | `/books/{id}` | Удаление книги |

Пример создания:

```http
POST /books
Content-Type: application/json

{
  "author_id": 1,
  "title": "New Book",
  "description": "Book description"
}
```

### Авторы

| Метод | Endpoint | Описание |
| --- | --- | --- |
| GET | `/authors` | Список авторов |
| GET | `/authors/{id}` | Автор и его книги |
| POST | `/authors` | Создание автора |
| PUT | `/authors/{id}` | Обновление автора |
| DELETE | `/authors/{id}` | Удаление автора |

### Пользователи и списки чтения

| Метод | Endpoint | Описание |
| --- | --- | --- |
| GET | `/users` | Список пользователей |
| GET | `/users/{id}` | Пользователь по ID |
| GET | `/users/{id}/reading-list` | Список чтения |
| POST | `/users/{id}/reading-list` | Добавление книги |
| DELETE | `/users/{id}/reading-list/{book_id}` | Удаление книги из списка |

Пример добавления:

```http
POST /users/2/reading-list
Content-Type: application/json

{
  "book_id": 5
}
```

### Поиск, фильтрация и пагинация

```text
GET /books?search=book
GET /books?author=Rowling
GET /books?author_id=1
GET /books?sort=title&order=asc
GET /books?limit=5&offset=0
GET /books?author=Rowling&sort=title&order=asc&limit=5&offset=0
```

Поиск выполняется по названию и описанию. Поддерживаются сортировка по `id`, `title`, `author`, порядок `asc`/`desc` и параметры `limit`/`offset`.

### Ошибки

REST использует HTTP-коды 400, 404, 409 и 500. Например, попытка удалить автора с книгами возвращает 409 Conflict:

```json
{
  "error": "author cannot be deleted because they have books"
}
```

Ошибки отсутствующих записей, конфликтов и некорректных ссылок обрабатываются через общие ошибки приложения. GraphQL возвращает ошибки в стандартном поле `errors`.

## GraphQL API

Endpoint: `http://localhost:8080/graphql`.

Playground: `http://localhost:8080/playground`.

GraphQL использует те же репозитории и предметную модель, что и REST. Клиент может выбирать необходимые поля и получать связанные данные одним запросом.

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

```graphql
query {
  book(id: "1") {
    id
    title
    author { name }
  }
}
```

### Поиск и фильтрация

```graphql
query {
  books(filter: {
    search: "book"
    sort: "title"
    order: "asc"
    limit: 5
    offset: 0
  }) {
    id
    title
    author { name }
  }
}
```

Фильтр поддерживает `search`, `author`, `authorId`, `sort`, `order`, `limit` и `offset`.

### Авторы и пользователи

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
      author { name }
    }
  }
}
```

### Mutations

Создание книги:

```graphql
mutation {
  createBook(input: {
    authorId: "1"
    title: "New Book"
    description: "Book description"
  }) {
    id
    title
    author { name }
  }
}
```

Обновление книги:

```graphql
mutation {
  updateBook(id: "1", input: {
    authorId: "1"
    title: "Updated Book"
    description: "Updated description"
  }) {
    id
    title
  }
}
```

Удаление книги:

```graphql
mutation {
  deleteBook(id: "1")
}
```

Создание, обновление и удаление автора:

```graphql
mutation {
  createAuthor(input: { name: "New Author" }) {
    id
    name
  }
}
```

```graphql
mutation {
  updateAuthor(id: "1", input: { name: "Updated Author" }) {
    id
    name
  }
}
```

```graphql
mutation {
  deleteAuthor(id: "1")
}
```

Список чтения:

```graphql
query {
  readingList(userId: "2") {
    id
    title
    author { name }
  }
}
```

```graphql
mutation {
  addBookToReadingList(userId: "2", bookId: "5")
}
```

```graphql
mutation {
  removeBookFromReadingList(userId: "2", bookId: "5")
}
```

### Функциональный паритет

| Возможность | REST | GraphQL |
| --- | --- | --- |
| Получение книг и авторов | ✓ | ✓ |
| Поиск и фильтрация | ✓ | ✓ |
| Сортировка и пагинация | ✓ | ✓ |
| CRUD книг | ✓ | ✓ |
| CRUD авторов | ✓ | ✓ |
| Получение пользователей | ✓ | ✓ |
| Получение reading list | ✓ | ✓ |
| Добавление и удаление из reading list | ✓ | ✓ |
| Получение связанных данных | ✓ | ✓ |

Основные операции реализованы в обоих API. Корректность произвольных вложенных GraphQL-запросов и отсутствие N+1-запросов будут отдельно проверяться при подготовке нагрузочных сценариев.

## Web-интерфейс

Интерфейс доступен по адресу `http://localhost:8080/`. Он позволяет просматривать каталог, искать и фильтровать книги, добавлять их в список чтения и удалять из него.

Сейчас frontend использует REST и тестового пользователя. Авторизация и переключение между REST/GraphQL и PostgreSQL/MongoDB пока не реализованы.

## Запуск PostgreSQL

При локальной установке PostgreSQL в WSL:

```bash
sudo service postgresql start
```

Создание пользователя и базы (выполняется администратором PostgreSQL):

```sql
CREATE USER book_user WITH PASSWORD 'your_password';
CREATE DATABASE book_catalog OWNER book_user;
```

Применение миграций из корня проекта:

```bash
psql -h localhost -U book_user -d book_catalog -f migrations/001_init.sql
psql -h localhost -U book_user -d book_catalog -f migrations/002_seed.sql
```

Актуальная `002_seed.sql` должна содержать только необходимые тестовые данные пользователей, если старый набор из 12 книг больше не используется. Для существующей базы повторно применять миграции без необходимости не нужно.

## Запуск MongoDB

MongoDB запускается через Docker Compose:

```bash
docker compose up -d mongodb
```

Проверка:

```bash
docker ps
```

Подключение к контейнеру:

```bash
docker exec -it book-catalog-mongo mongosh \
  -u book_user \
  -p book_password \
  --authenticationDatabase admin
```

В `mongosh`:

```javascript
use book_catalog
db.runCommand({ ping: 1 })
show collections
db.books.countDocuments()
db.authors.countDocuments()
```

Docker volume сохраняет данные между обычными остановками и перезапусками контейнера. Команда `docker compose down` не удаляет volume, если не указан флаг `-v`.

## Переменные окружения

Создать `.env` на основе примера:

```bash
cp .env.example .env
```

Пример конфигурации:

```env
DATABASE_URL=postgres://book_user:your_password@127.0.0.1:5432/book_catalog?sslmode=disable
MONGO_URL=mongodb://book_user:your_password@127.0.0.1:27017/?authSource=admin
MONGO_DATABASE=book_catalog
SERVER_PORT=8080
```

Пароли должны соответствовать локальной конфигурации. `.env` не добавляется в Git. При изменении пароля MongoDB в Compose следует учитывать, что переменные инициализации не изменяют пароль уже созданного пользователя в существующем volume.

## Подготовка данных

Исходный CMU-файл размещается в `data/raw/`. Подготовщик читает TSV, выбирает название, автора и описание, удаляет неподходящие записи и сохраняет 100 книг в CSV.

Запуск из корня проекта:

```bash
go run ./cmd/importer/main.go
```

Результат:

```text
data/prepared/books.csv
```

Если файл был переименован, необходимо синхронизировать пути в подготовщике и обоих импортёрах.

## Импорт в PostgreSQL

```bash
go run ./cmd/importer/load_postgres.go
```

Импортёр читает подготовленный CSV, ищет существующих авторов по имени, создаёт недостающих и добавляет книги через PostgreSQL-репозитории.

Текущий импортёр не следует считать полностью идемпотентным: повторный запуск может создать дубли книг. Для повторного развёртывания будет добавлена проверка уникальности или отдельный воспроизводимый сценарий загрузки.

Проверка:

```sql
SELECT COUNT(*) FROM books;
SELECT COUNT(*) FROM authors;
```

После очистки старых тестовых записей каталог содержит 100 книг. Идентификаторы книг и авторов приведены к последовательным значениям, а связи сохранены.

## Импорт в MongoDB

```bash
go run ./cmd/importer/load_mongo.go
```

Импортёр читает тот же CSV и создаёт документы в коллекциях `authors` и `books`. Текущая версия очищает эти две коллекции перед загрузкой, поэтому запускать её следует только для тестовой базы, данные которой можно пересоздать.

Проверка в `mongosh`:

```javascript
use book_catalog

db.books.countDocuments()
db.authors.countDocuments()

db.books.find().sort({ id: 1 }).limit(5)
```

Проверка связи с автором:

```javascript
db.books.aggregate([
  {
    $lookup: {
      from: "authors",
      localField: "author_id",
      foreignField: "id",
      as: "author"
    }
  },
  { $limit: 5 }
])
```

В MongoDB загружены 100 книг. Репозитории для обслуживания REST и GraphQL ещё предстоит реализовать.

## Запуск приложения

Из корня проекта:

```bash
go mod download
go run ./cmd/server
```

Проверка:

```bash
curl http://localhost:8080/health
```

После запуска доступны REST API, GraphQL, Playground и web-интерфейс. На текущем этапе сервер использует PostgreSQL. Наличие `MONGO_URL` в `.env` само по себе не переключает API на MongoDB.

## Тестирование

REST проверяется через Postman, GraphQL через Playground. Проверены CRUD книг и авторов, поиск, фильтрация, сортировка, пагинация, получение пользователей, операции reading list и основные ошибки.

Примеры REST:

```text
GET /books
GET /books?limit=5&offset=0
GET /authors
GET /users
GET /users/2/reading-list
```

Пример GraphQL:

```graphql
query {
  books(filter: { limit: 5 }) {
    id
    title
    author { name }
  }
}
```

Проверка сборки:

```bash
go fmt ./...
go test ./...
```

Автоматизированные интеграционные и нагрузочные тесты будут добавлены на следующих этапах. Ручные проверки не заменяют полноценный набор автотестов.

## План сравнительного исследования

Планируется сравнить четыре комбинации:

| API | База данных |
| --- | --- |
| REST | PostgreSQL |
| GraphQL | PostgreSQL |
| REST | MongoDB |
| GraphQL | MongoDB |

Для сравнения будут использоваться одинаковые бизнес-сценарии: получение списка книг, поиск и фильтрация, связанные данные, CRUD и операции со списками чтения.

Планируемые метрики: среднее и медианное время ответа, p95, количество HTTP-запросов на пользовательскую операцию, объём ответа, пропускная способность, ошибки и потребление ресурсов. Для корректности измерений необходимо использовать одинаковые данные, сопоставимые индексы, повторные прогоны, прогрев и восстановление состояния после операций записи.

Отдельно будет изучено, в каких сценариях GraphQL уменьшает количество запросов или объём передаваемых данных, а также как особенности PostgreSQL и MongoDB влияют на выполнение одинаковых операций. Заранее не предполагается, что какой-либо API или база данных всегда быстрее.

## Дальнейшие этапы

1. Реализовать MongoDB BookRepository, AuthorRepository и UserRepository.
2. Добавить коллекции пользователей и списков чтения в MongoDB.
3. Реализовать одинаковую обработку ошибок и проверку связей.
4. Добавить выбор БД через `DB_DRIVER` при запуске сервера.
5. Запустить два экземпляра API с PostgreSQL и MongoDB на разных портах.
6. Проверить полный функциональный паритет обоих хранилищ.
7. Добавить переключение REST/GraphQL и PostgreSQL/MongoDB во frontend.
8. Реализовать авторизацию, JWT и роли пользователей.
9. Добавить интеграционные и нагрузочные тесты.
10. Подготовить Docker Compose для всех компонентов.
11. Провести сравнительные измерения и проанализировать результаты.
12. Подготовить итоговый отчёт.

## Текущий статус

REST и GraphQL работают с PostgreSQL через общие repository-интерфейсы. PostgreSQL содержит очищенный набор из 100 книг. MongoDB запущена в Docker и заполнена тем же подготовленным CSV. Подключение MongoDB к API, переключение баз данных, авторизация и сравнительные измерения пока находятся в разработке.

Следующий основной этап: реализация MongoDB-репозиториев и подключение MongoDB как полноценного backend-хранилища.
