
Сервис для обработки транзакций между кошельками на Go и SQLite.

---

### Локальный запуск

make run-local

### Запуск в Docker

1. Сборка:


make docker-build

2. Запуск:


make docker-run

---

## API

- `GET /api/wallet/{address}/balance` — получить баланс кошелька.
- `POST /api/send` — перевод средств между кошельками (ожидается JSON):
    ```
    {
      "from": "64-символьный адрес отправителя",
      "to": "64-символьный адрес получателя",
      "amount": 100.5
    }
    ```
- `GET /api/transactions?count=N` — получить последние N транзакций.


---

## Структура

- Точка входа: `cmd/main.go`
- Конфиги: `config/local.yaml`, `config/docker.yaml`
- HTTP API и обработчики: `/internal/http-server`
- Модели: `/internal/models/transaction`
- Логгер: `/internal/lib/logger/sl`
- Хранилище: `/internal/storage/sqlite`
- Makefile для всех задач проекта

---

## Зависимости

- `gorilla/mux` — роутинг.
- `cleanenv` — работа с конфигами.
- `go-sqlite3` — работа с SQLite.

---