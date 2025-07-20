# Marketplace API

REST API условного маркетплейса, реализованного на Go с использованием Gin, PostgreSQL, JWT и Docker.

## Описание

- Регистрация пользователей
- Авторизация пользователей с получением JWT токена
- Размещение объявлений
- Лента объявлений с фильтрами и пагинацией
- Swagger-документация
- Защищённые эндпоинты (только с токеном)

## Технологии

- Go + Gin
- PostgreSQL
- Goose (миграции)
- JWT (аутентификация)
- Swagger (документация)
- Docker + Docker Compose

## Запуск контейнера

```bash
docker compose -f build/docker-compose.yml up -d --build
```

Swagger-документация доступна по адресу

[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)