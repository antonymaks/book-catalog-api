# Book Catalog API

Учебный проект книжного каталога на Go.

Основная цель проекта — реализовать одинаковую функциональность с использованием REST и GraphQL, а также PostgreSQL и MongoDB. Это позволяет сравнивать разные варианты API и хранения данных на одной предметной области.

## Стек

- Go
- REST API
- GraphQL (gqlgen)
- PostgreSQL
- MongoDB
- Docker / Docker Compose
- JWT
- HTML, CSS, JavaScript

## Функционал

В проекте реализованы:

- получение, поиск и фильтрация книг;
- CRUD книг и авторов;
- REST и GraphQL API;
- работа с PostgreSQL и MongoDB;
- регистрация и авторизация пользователей;
- JWT-аутентификация;
- роли `user` и `admin`;
- персональный список чтения;
- административные операции для книг и авторов;
- простой web-интерфейс для проверки работы приложения.

Для обеих баз данных используется один набор из 100 книг.

Проект можно запускать в четырех конфигурациях:

- REST + PostgreSQL;
- GraphQL + PostgreSQL;
- REST + MongoDB;
- GraphQL + MongoDB.

## Запуск

Для запуска необходим Docker.

Клонировать репозиторий:

```bash
git clone https://github.com/antonymaks/book-catalog-api.git
cd book-catalog-api
```

Запустить проект:

```bash
docker compose up --build
```

При первом запуске Docker создаст PostgreSQL и MongoDB, загрузит тестовые данные и запустит два экземпляра приложения.

После запуска доступны:

```text
PostgreSQL: http://localhost:8080
MongoDB:    http://localhost:8081
```

GraphQL:

```text
http://localhost:8080/graphql
http://localhost:8081/graphql
```

GraphQL Playground:

```text
http://localhost:8080/playground
http://localhost:8081/playground
```

Тестовый администратор:

```text
Логин: admin
Пароль: admin123
```

Остановка проекта:

```bash
docker compose down
```

Для полного удаления контейнеров вместе с данными:

```bash
docker compose down -v
```