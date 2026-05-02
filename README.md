```markdown
# Vessel 🚢

[![Go Reference](https://pkg.go.dev/badge/github.com/krasilovalex/vessel.svg)](https://pkg.go.dev/github.com/krasilovalex/vessel)

**Vessel** — это легковесная Go-библиотека (обертка над официальным Docker SDK) для управления локальными dev-окружениями и создания крутых CLI-утилит.

Главная цель Vessel: **Великолепный Developer Experience (DX)**. 
Больше никаких километров YAML-конфигов и громоздких Dockerfile. Управляйте контейнерами и собирайте образы на лету прямо из Go-кода через удобный Fluent API.

## Фичи ✨

- 🚀 **Fluent API**: аккумулирование ошибок и цепочки вызовов.
- 🐳 **Без YAML**: поднимайте базы данных и сервисы в пару строк кода.
- 🛠 **Убийца Dockerfile (Builder API)**: динамическая генерация `Dockerfile` и контекста сборки прямо в оперативной памяти (in-memory tar).
- 🧹 **Автоочистка**: удобные методы для graceful остановки и удаления контейнеров.

## Установка
```bash
go get [github.com/krasilovalex/vessel](https://github.com/krasilovalex/vessel)
```

## Быстрый старт

Поднимаем PostgreSQL 15 с пробросом портов и монтированием volume:
```go
package main

import (
	"context"
	"log"
	"os"

	"[github.com/krasilovalex/vessel](https://github.com/krasilovalex/vessel)"
)

func main() {
	ctx := context.Background()
	cwd, _ := os.Getwd()

	// Настраиваем контейнер через Fluent API
	pg := vessel.NewContainer("postgres:15-alpine").
		WithName("my-dev-db").
		WithPort("5432", "5432").
		WithEnv("POSTGRES_PASSWORD", "secret").
		WithVolume(cwd+"/pgdata", "/var/lib/postgresql/data")

	// Запускаем
	if err := pg.Up(ctx); err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
	}

	// ... работаем с БД ...

	// Останавливаем и удаляем
	_ = pg.Stop(ctx)
	_ = pg.Remove(ctx)
}
```

## Статус проекта
Библиотека находится в активной разработке. API может меняться.
```