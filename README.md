# 🚢 Vessel: The Go-Native Docker Orchestrator

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/vessel.svg)](https://pkg.go.dev/github.com/yourusername/vessel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Status: MVP](https://img.shields.io/badge/Status-MVP-brightgreen.svg)]()

*(🇷🇺 Русская версия документации находится [ниже](#-vessel-на-русском))*

**YAML is dead. Long live Go.**

Vessel is a high-performance, programmable Docker orchestrator built entirely in Go. It replaces static, bloated `docker-compose.yml` files with dynamic, type-safe, and infinitely scalable Go code. Whether you have 3 microservices or 300, Vessel builds, connects, and launches your environment concurrently with a beautiful terminal UI.

## ⚡ Why Vessel? (The End of YAML Hell)
*   **Programmable:** Need to start 10 identical workers? Use a `for` loop. Don't copy-paste 150 lines of YAML.
*   **Type-Safe:** Catch infrastructure errors at compile-time. Your IDE provides auto-completion for your entire deployment.
*   **Dynamic Secrets:** Fetch passwords from AWS Secrets Manager or HashiCorp Vault directly in your Go script *before* injecting them into containers.
*   **Blazing Fast:** Starts containers concurrently using Go channels and `errgroup`.

## 📦 Installation

```bash
go get [github.com/krasilovalex/vessel](https://github.com/krasilovalex/vessel)
```

## ✨ Core Features
*   🛠️ **In-Memory Multi-Stage Builds:** Compile Go, Node.js, or Python apps directly into lightweight images (~15MB) on the fly.
*   🧠 **Smart DAG Dependencies:** Channel-based `DependsOn` ensures your backend naturally waits for your databases to be fully ready.
*   🌐 **Batteries Included:** Automatic Docker Networks (`bridge`) and Named Volumes out of the box.
*   🎨 **Sleek TUI:** Real-time parallel spinners powered by `pterm`.

---

## 🚀 The Power of Vessel: Real-World Examples

### Example 1: The Polyglot Fleet (Go, Node.js, Python)
Vessel doesn't care what language your apps use. It builds and orchestrates anything.

```go
// 1. Build a Node.js Frontend locally
_ = vessel.NewBuilder("my-front:latest").
    From("node:18-alpine").Workdir("/app").CopyDir("./frontend", ".").
    Run("npm install").Cmd("npm start").Build(ctx)

// 2. Describe a Python ML Worker (from an external registry)
pythonWorker := vessel.NewContainer("my-registry/py-ml-worker:v1.2")

// 3. Describe the Go API and link them
goApi := vessel.NewContainer("my-go-api:latest").
    WithPort("8080", "8080").
    DependsOn(pythonWorker) // Go waits for Python

fleet := vessel.NewFleet(goApi, pythonWorker, vessel.NewContainer("my-front:latest"))
fleet.Up(ctx)
```

### Example 2: The "YAML Killer" (7 Microservices in a Loop)
Imagine writing a Compose file for 5 identical microservices. With Vessel, it's just a few lines of elegant Go code:

```go
db := vessel.NewContainer("postgres:15").WithVolume("pg-data", "/var/lib/postgresql/data")
redis := vessel.NewContainer("redis:7-alpine").WithVolume("redis-data", "/data")

services := []*vessel.Container{}
names := []string{"auth", "users", "billing", "notifications", "analytics"}

// Look at this magic. No YAML copy-pasting.
for _, name := range names {
    _ = vessel.NewBuilder(name+":latest").
        From("golang:1.22-alpine").CopyDir("./services/"+name, ".").
        Run("go build -o server main.go").Build(ctx)
    
    // Every service depends on DB and Redis
    c := vessel.NewContainer(name+":latest").DependsOn(db, redis)
    services = append(services, c)
}

// Gateway waits for ALL 5 microservices
gateway := vessel.NewContainer("nginx:alpine").WithPort("80", "80").DependsOn(services...)

allContainers := append([]*vessel.Container{db, redis, gateway}, services...)
vessel.NewFleet(allContainers...).Up(ctx)
```

## 🗺️ Roadmap
- [x] Multi-stage dynamic builds
- [x] Ephemeral bridge networks & volume management
- [x] Concurrent TUI execution
- [x] `DependsOn` dependency graph
- [ ] Active port-pinging Healthchecks
- [ ] Real-time stdout/stderr log streaming
- [ ] Label-based garbage collection for dangling resources

## 📄 License
MIT License.

---

# 🚢 Vessel (на русском)

**YAML мёртв. Да здравствует Go.**

Vessel — это высокопроизводительный, программируемый оркестратор Docker, написанный на чистом Go. Он заменяет раздутые и статичные файлы `docker-compose.yml` на динамичный, строго типизированный код. Неважно, у вас 3 микросервиса или 300 — Vessel соберет, свяжет и запустит всё параллельно с невероятно красивым интерфейсом в терминале.

## ⚡ Почему Vessel? (Конец эпохи YAML-лапши)
*   **Программируемость:** Нужно запустить 10 одинаковых воркеров? Используйте цикл `for`. Хватит копипастить 150 строк YAML.
*   **Строгая типизация:** Ваша IDE (Goland, VS Code) сама подскажет методы и подсветит ошибки архитектуры еще до запуска.
*   **Динамические секреты:** Получайте пароли из AWS или Vault напрямую в скрипте и динамически передавайте их в контейнеры при старте.
*   **Невероятная скорость:** Конкурентный запуск сотен контейнеров благодаря горутинам и `errgroup`.

## ✨ Главные фичи
*   🛠️ **Multi-Stage сборки в памяти:** Компилируйте Go, Node.js или Python прямо на лету, получая образы весом 15MB.
*   🧠 **Умный граф зависимостей:** Метод `DependsOn` под капотом использует каналы Go, чтобы бэкенд терпеливо ждал старта баз данных.
*   🌐 **Всё включено:** Встроенная маршрутизация Docker Networks и поддержка Named Volumes.
*   🎨 **Стильный TUI:** Параллельные спиннеры и логирование на базе `pterm`.

---

## 🚀 Мощь Vessel: Реальные примеры

### Пример 1: Полиглот-архитектура (Go, Node.js, Python)
Vessel всё равно, на чем написан ваш код. Это универсальный клей для любых технологий.

```go
// 1. Собираем Node.js Frontend из исходников локально
_ = vessel.NewBuilder("my-front:latest").
    From("node:18-alpine").Workdir("/app").CopyDir("./frontend", ".").
    Run("npm install").Cmd("npm start").Build(ctx)

// 2. Описываем Python ML Worker (скачается из приватного Registry)
pythonWorker := vessel.NewContainer("my-registry/py-ml-worker:v1.2")

// 3. Описываем Go API и связываем их
goApi := vessel.NewContainer("my-go-api:latest").
    WithPort("8080", "8080").
    DependsOn(pythonWorker) // Go дождется старта Python!

fleet := vessel.NewFleet(goApi, pythonWorker, vessel.NewContainer("my-front:latest"))
fleet.Up(ctx)
```

### Пример 2: Убийца YAML (Запуск 7 микросервисов в цикле)
Представьте, как выглядит `docker-compose.yml` для 5 микросервисов. В Vessel это пара строк элегантного кода:

```go
db := vessel.NewContainer("postgres:15").WithVolume("pg-data", "/var/lib/postgresql/data")
redis := vessel.NewContainer("redis:7-alpine").WithVolume("redis-data", "/data")

services := []*vessel.Container{}
names := []string{"auth", "users", "billing", "notifications", "analytics"}

// Оцените магию. Никакого копипаста!
for _, name := range names {
    _ = vessel.NewBuilder(name+":latest").
        From("golang:1.22-alpine").CopyDir("./services/"+name, ".").
        Run("go build -o server main.go").Build(ctx)
    
    // Каждый сервис зависит от БД и Кэша
    c := vessel.NewContainer(name+":latest").DependsOn(db, redis)
    services = append(services, c)
}

// Gateway ждет, пока поднимутся ВСЕ 5 микросервисов
gateway := vessel.NewContainer("nginx:alpine").WithPort("80", "80").DependsOn(services...)

allContainers := append([]*vessel.Container{db, redis, gateway}, services...)
vessel.NewFleet(allContainers...).Up(ctx)
```

## 🗺️ План развития (Roadmap)
- [x] Динамическая Multi-stage сборка
- [x] Авто-сети (`bridge`) и управление томами
- [x] Параллельный запуск и TUI
- [x] Умный граф запуска `DependsOn`
- [ ] Настоящие Healthchecks (пинг портов)
- [ ] Стриминг `stdout/stderr` прямо в терминал
- [ ] Сборка мусора (Dangling resources) по лейблам

## 📄 Лицензия
MIT License.