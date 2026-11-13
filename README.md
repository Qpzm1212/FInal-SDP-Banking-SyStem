# Banking System (SDP Final Project)

Web banking app built with Go 

## Quickstart

Requirements: Go 1.20+ 

Run the server:
```bash
go run ./cmd/api/main.go
```
Open `http://localhost:8080` in your browser.

Run all tests:
```bash
go test -v ./...
```

---


---

## Division of Work

| Student | Patterns | Core Files | Other Work |
|---|---|---|---|
| Shynarbek Nuraziz (Student A) | Singleton, Factory Method, Facade | `internal/db/singleton.go`, `internal/patterns/factory.go`, `internal/service/facade.go` | Models (`internal/models/data.go`), HTTP API (`cmd/api/main.go`), DB & service tests |
| Asset Kamza (Student B) | Builder, Strategy, Observer | `internal/patterns/builder.go`, `internal/patterns/strategy.go`, `internal/patterns/observer.go` | Web UI (`web/index.html`), pattern unit tests, project report |

---

