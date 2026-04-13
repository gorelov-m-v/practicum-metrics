# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` - адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` - порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Сборка с информацией о версии

При сборке можно задать версию, дату и коммит через флаги линковщика (`-ldflags`):

```bash
go build -ldflags "\
  -X main.buildVersion=v1.0.0 \
  -X main.buildDate=$(date +%F) \
  -X main.buildCommit=$(git rev-parse --short HEAD)" \
  -o server ./cmd/server

go build -ldflags "\
  -X main.buildVersion=v1.0.0 \
  -X main.buildDate=$(date +%F) \
  -X main.buildCommit=$(git rev-parse --short HEAD)" \
  -o agent ./cmd/agent
```

Если переменные не заданы, используется значение по умолчанию `N/A`.

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Профилирование и оптимизация (iter17)

### Бенчмарки

Добавлены бенчмарки для ключевых компонентов системы:
- **Storage** - `UpdateGauge`, `UpdateCounter`, `GetGauge`, `GetCounter`, `GetAllGauges`, `GetAllCounters`, `UpdateMetricsBatch`
- **Service** - `UpdateGauge`, `UpdateCounter`, `GetGauge`, `GetCounter`, `UpdateMetricsBatch`
- **Handler** - `UpdateMetricJSON` (gauge/counter), `GetMetricJSON`, `UpdateMetricsBatch`, `ListMetrics`

Запуск: `go test -bench=. -benchmem ./internal/...`

### pprof

Профилировщик `net/http/pprof` подключён к серверу по пути `/debug/pprof/`.

### Анализ base-профиля: что было неэффективно

При снятии профиля памяти (`profiles/base.pprof`) через бенчмарки обнаружены следующие проблемы:

1. **`ListMetrics` - 1864 MB cum, 1203 allocs/op, 60728 B/op**
   - Слайс `metrics` создавался без указания capacity (`var metrics []metricData`), что приводило к многократным реаллокациям и копированиям при `append`.
   - Для форматирования значений метрик использовался `fmt.Sprintf("%g", value)` и `fmt.Sprintf("%d", value)`. Пакет `fmt` использует рефлексию и внутренние буферы, что порождает лишние аллокации на каждый вызов.
   - Строковые константы типа метрик (`string(storage.Gauge)`) пересоздавались на каждой итерации цикла.

2. **`UpdateMetricsBatch` (через сервис) - 348 allocs/op, 42256 B/op**
   - При batch-обновлении в memory storage сервис вызывал `UpdateGauge`/`UpdateCounter` в цикле по одной метрике. Каждый вызов:
     - создавал новый `context.WithTimeout` (4 аллокации, 272 B на вызов);
     - захватывал и отпускал мьютекс отдельно для каждой метрики.
   - При 100 метриках в батче это давало 100 лишних контекстов и 100 lock/unlock вместо одного.

3. **`context.WithDeadlineCause` - 1071 MB cum**
   - Каждый вызов сервисного метода создавал `context.WithTimeout`, даже когда операция - простая запись в map за мьютексом (~30 нс). Таймаут в 5 секунд на операцию, выполняющуюся за наносекунды, порождал аллокации, которые стоили дороже самой операции.

### Что было исправлено

1. **`ListMetrics` - предварительная аллокация и замена `fmt` на `strconv`**
   - Слайс создаётся с известным capacity: `make([]metricData, 0, len(gauges)+len(counters))` - исключены реаллокации.
   - `fmt.Sprintf("%g", value)` заменён на `strconv.FormatFloat(value, 'g', -1, 64)`.
   - `fmt.Sprintf("%d", value)` заменён на `strconv.FormatInt(value, 10)`.
   - Строковые константы типов метрик вынесены из цикла.
   - **Результат**: 60728 B/op -> 55895 B/op, 1203 allocs/op -> 1148 allocs/op.

2. **`UpdateMetricsBatch` - интерфейс `BatchUpdater` для MemStorage**
   - Добавлен интерфейс `storage.BatchUpdater` с методом `UpdateMetricsBatch`.
   - `MemStorage` реализует этот интерфейс - весь batch обрабатывается за один захват мьютекса.
   - Сервис проверяет, поддерживает ли storage `BatchUpdater`, и если да - вызывает batch-метод напрямую, минуя N отдельных `UpdateGauge`/`UpdateCounter` с N `context.WithTimeout`.
   - **Результат**: устранены N лишних `context.WithTimeout` и N lock/unlock при batch-обновлении.

### Результат сравнения профилей

```
$ go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

Type: alloc_space
Showing nodes accounting for -906108.63kB, 3.63% of 24934952.11kB total
      flat  flat%   sum%        cum   cum%
-400343.49kB  1.61%  1.61% -400343.49kB  1.61%  bufio.NewReaderSize (inline)
-135994.99kB  0.55%  2.15% -104158.01kB  0.42%  handler.(*MetricHandler).ListMetrics
-84728.10kB  0.34%  2.49% -84728.10kB  0.34%  encoding/json.(*Decoder).refill
-84507.38kB  0.34%  2.83% -84507.38kB  0.34%  net/textproto.MIMEHeader.Set (inline)
-65555.54kB  0.26%  2.77% -65555.54kB  0.26%  net/http.Header.Clone (inline)
-65539.69kB  0.26%  3.03% -53250.38kB  0.21%  context.WithDeadlineCause
-46606.22kB  0.19%  3.22% -46606.22kB  0.19%  net/http.(*Request).WithContext (inline)
-6656.10kB 0.027%  3.65% -7168.98kB 0.029%  fmt.Sprintf
```

Отрицательные значения показывают уменьшение потребления памяти после оптимизаций. Суммарное снижение alloc_space составило ~906 MB за время работы бенчмарков.
