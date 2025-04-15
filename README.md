# Avito PVZ Service

Сервис для работы с пунктами выдачи заказов.

# Быстрый старт
## Требования
- Docker 20.10+
- Docker Compose 2.20+

## Развертывание в Docker

1. Клонировать репозиторий:
```
git clone https://github.com/maksemen2/pvz-service.git
cd pvz-service
```
2. Запустить сервис:
```
make deploy

# Или

docker compose up --build 
```
API сервер будет доступен по адресу http://localhost:8080, сервер метрик по адресу http://localhost:9000, а gRPC сервер по адресу localhost:3000

⚠️ Важно: генерируемые файлы (моки, dto и grpc) не добавлены в репозиторий. При развертии проекта в Docker они будут сгенерированы автоматически. 
Для генерации вручную воспользуйтесь командой 
```make generate```
Для этого у вас должны быть установлены protoc, oapi-codegen, mockgen и protoc-gen-go. 
Для их установки воспользуйтесь командами:
```
apt update && apt install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install go.uber.org/mock/mockgen@latest
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
```

Или целью Makefile
```
make gen-install
```

## Выполнение задания
### Основное задание
1. Реализованы все требования, указанные в описании задания.
2. Реализован интеграционный тест, описанный в задании ([app_test.go](internal/app/app_test.go))

### Дополнительные задания
1. Реализована пользовательская авторизация по методам /register и /login ([auth.go](internal/delivery/http/handlers/auth.go))
2. Реализован gRPC метод для получения всех пвз ([grpc](internal/delivery/grpc))
3. В проект добавлен Prometheus ([metrics](internal/pkg/metrics)), он доступен на 9000 порту по ручке /metrics. Пример вывода:
```
# HELP business_products_added_total Total number of added products
# TYPE business_products_added_total counter
business_products_added_total 150
# HELP business_pvz_created_total Total number of created PVZs
# TYPE business_pvz_created_total counter
business_pvz_created_total 93
# HELP business_receptions_created_total Total number of created receptions
# TYPE business_receptions_created_total counter
business_receptions_created_total 27
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/pvz",status="OK"} 16
http_requests_total{method="POST",path="/pvz",status="Created"} 93
# HELP http_response_time_seconds Duration of HTTP requests
# TYPE http_response_time_seconds histogram
http_response_time_seconds_bucket{method="GET",path="/pvz",le="0.1"} 16
http_response_time_seconds_bucket{method="GET",path="/pvz",le="0.5"} 16
http_response_time_seconds_bucket{method="GET",path="/pvz",le="1"} 16
```
4. Настроено логирование ([logger](internal/pkg/logger)) с 4 уровнями: "debug", "info", "warn", "error" или "silent".
5. Настроена генерация DTO по OpenAPI схеме, а так же генерация моков для тестирования и кода gRPC сервера. Цели для генерации можно увидеть в файле [Makefile](Makefile)

## Тестирование:
- Юнит-тесты: testify
- Интеграционные тесты: testcontainers

Хендлеры, мидлвари и сервисы покрыты юнит-тестами.
Библиотеки покрыты юнит-тестами.
Репозитории и приложение покрыты интеграционными тестами

Покрытие кода тестами свыше 75%.

## Makefile цели
```
make lint # Запуск линтера
make lint-fix # Запуск линтера с автоисправлением
make generate # Генерация моков, dto и grpc схем
make unit-tests # Запуск юнит-тестов
make integration-tests # Запуск интеграционных тестов
make test # Запуск всех тестов
```

## Стек
- Go 1.23
- Gin
- СУБД: Postgresql
- Логирование: zap
- Генерация кода: oapi-codegen, protoc, mockgen


## Структура проекта
```
├───cmd # Точка входа
├───config # Конфиг
├───docs # oapi схема и .proto файл
├───internal
│   ├───app # Инкапсуляция зависимостей и их инициализация
│   ├───common
│   │   └───errors # типовые http ответы с ошибками
│   ├───delivery
│   │   ├───grpc # gRPC сервер
│   │   │   ├───handlers
│   │   │   ├───pvz_v1
│   │   │   └───server
│   │   └───http # HTTP сервер
│   │       ├───handlers
│   │       ├───httpdto
│   │       ├───routes
│   │       └───server
│   ├───domain
│   │   ├───errors # Доменные ошибки
│   │   ├───models # Доменные модели
│   │   └───repositories # Интерфейсы репозиториев
│   │
│   ├───pkg
│   │   ├───auth # Функционал для авторизации
│   │   │   └───jwt # Имплементация JWT-менеджера токенов
│   │   │
│   │   ├───database # Функционал для работы с БД (PostgreSQL)
│   │   ├───logger # Логирование
│   │   ├───metrics # Prometheus метрики
│   │   └───testhelpers # Вспомогательные функции для тестов
│   ├───repository
│   │   ├───errors # Ошибки репозиториев
│   │   └───postgresql # Реализация репозиториев для Postgresql
│   └───service # Реализации сервисов
│
└───migrations # Миграции БД
```
