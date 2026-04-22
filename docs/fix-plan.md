# План фундаментального фикса log-parser

Документ описывает целевое состояние парсера и список изменений. Цель — не «залатать до понедельника», а убрать архитектурные предпосылки, из-за которых мы спустя месяц снова сюда возвращаемся.

> Сначала — что сломано на уровне архитектуры, потом — целевая модель, потом — пошаговый план работ, тесты, миграции БД, логирование. В конце — открытые вопросы, на которые нужны решения до старта.

---

## 1. Что не так структурно (корневая критика)

1. **Параллельные слайсы вместо сессии.**
   `GroupedDataDTO.TestStationRecords []` и `TestSteps [][]` держатся независимо и неявно сопоставляются по индексу. Ни парсер, ни группировщик не гарантируют, что они реально парные. Отсюда все текущие `mismatch`-ошибки.

2. **Парсер знает тип шагов, но теряет его.**
   В `ParseMixedJSONArray` из шагов уже извлекается PCBA через `"PCBA Scan"` / `"Compare PCBA Serial Number"` / `"Valid PCBA Serial Number"`. По этому же шагу можно однозначно определить **тип станции** (PCBA / Final) — но мы это не фиксируем.

3. **Карта `allStations` ключом по PCBA — без типа и с перезаписью.**
   Несколько станций одного PCBA (retry, PCBA + Final) перезаписывают друг друга. В `results` они попадают все, но логически «потеряны» для сопоставления.

4. **Диспатчер all-or-nothing на файл.**
   Первая ошибка в одной группе → `return`, весь файл теряется. Даже корректно распарсенные PCBA в БД не уходят.

5. **БД не идемпотентна.**
   Migration `002_remove_unique_constraints` сняла все unique-ограничения. `GetOrInsertLogisticData` вставляет безусловно. Повторный прогон CLI по тому же файлу плодит дубликаты во всех таблицах. Нигде нет `created_at` — понять «какие строки свежие» невозможно.

6. **Транзакций нет.**
   Падение на середине группы оставляет БД в полусостоянии (логистика вставлена, станция нет, шаги нет).

7. **Валидатор — пустой файл.**
   `processor/validator.go` содержит только `package processor`. Никакой валидации в пайплайне нет.

8. **Сообщения об ошибках искажают PCBA.**
   Ошибка диспатчера печатает `DownloadInfo.TcuPCBANumber`, а не ключ группы. Отсюда строки `for PCBA ` (пусто) в логах.

9. **`FilterRelevantJsonBlocks` + `ParseMixedJSONArray` делают одну работу дважды.**
   Первый — валидирует типы блоков, второй — переваливает их снова. Лишняя поверхность.

10. **Наблюдаемости ноль.**
    Нет ни агрегированной сводки по файлу (сколько блоков, сколько сессий, сколько сирот), ни метрик, ни per-PCBA аудита. Диагностика только по грепу текстовых сообщений.

---

## 2. Целевая модель: сессия как единица

Реальная модель данных на заводе — такая:

```
Device (идентификатор = PCBA)
  ├── DownloadInfo         (0..N, обычно 1)
  └── TestSession          (0..N, по одной на попытку прогона на станции)
        ├── Station record (метаданные: время, результат, коды ошибок)
        └── TestSteps      (шаги этой попытки)
```

**Сессия — атомарная единица пайплайна.** Парсер выдаёт именно сессии. Группировщик собирает их по PCBA. Диспатчер вставляет по одной сессии. Если в сессии есть только station record — шагов нет. Если только шаги — нет station record (сирота). Всегда ясно, что нарушено.

### Новая форма DTO

```go
// Одна попытка прогона устройства на станции.
// Минимум одно из {Record, Steps} должно быть непустым.
type TestSessionDTO struct {
    StationType string                // "PCBA" | "Final" (обязательно)
    PCBANumber  string                // для группировки
    Record      *TestStationRecordDTO // nil если сессия-сирота (шаги без записи)
    Steps       []TestStepDTO         // nil/[] если сессия без шагов
    LogSeq      int                   // порядковый номер в исходном логе
}

type GroupedDataDTO struct {
    PCBANumber    string              // ключ группы (или ProductSN если PCBA пуст)
    ProductSN     string
    DownloadInfos []DownloadInfoDTO
    Sessions      []TestSessionDTO    // в порядке появления в логе
}
```

Это **минимальный** контракт. Всё остальное (конвертер, диспатчер, API) поднимается на этом фундаменте.

### Алгоритм спаривания шагов и станций

Один проход по блокам в **порядке появления в логе**:

- `Station(pcba, type, record)` — пытаемся достать из очереди *непарных шагов* для `(pcba, type)` самый старый массив. Нашли → эмитируем `TestSessionDTO{Record, Steps}`. Нет → кладём station в свою очередь ожидания.
- `Steps(pcba, type, steps)` — симметрично. Пытаемся достать из очереди непарных станций для `(pcba, type)`. Нашли → полная сессия. Нет → кладём steps в очередь.
- `Download(pcba, info)` — просто прибавляется к `DownloadInfos` соответствующей группы.

В конце прохода:
- Оставшиеся в очередях непарные станции → сессии с `Steps=nil` (`orphan_station`).
- Оставшиеся непарные шаги → сессии с `Record=nil` (`orphan_steps`).

Это корректно обрабатывает все найденные в аудите паттерны: отсутствующие станции, retry'и, пустой PCBA с фоллбэком на ProductSN.

---

## 3. Пошаговый план работ

### Фаза 0 — hotfix, выкатывается первым

Цель: чтобы CLI сегодня вечером дошёл до конца файлов и залил то, что смог.

1. `dispatcher.go`: ловить ошибку на уровне группы → `logger.Warn` с ключом группы и `continue`. Не валить файл.
2. `dispatcher.go:96`: печатать ключ группы (PCBA/ProductSN), а не `DownloadInfo.TcuPCBANumber`.
3. При `len(TestSteps) > len(TestStationRecords)` — WARN, обработка продолжается с тем, что есть.

Фаза 0 **не** решает проблему корректности (данные частично теряются), но восстанавливает поток данных и не трогает архитектуру. Можно мерджить в тот же день.

### Фаза 1 — новый контракт DTO и спаривание в парсере

1. Добавить `TestSessionDTO`, обновить `GroupedDataDTO`.
2. В `json_parser.go`:
   - Переименовать/рефакторить `ParseMixedJSONArray` → `ParseLogBlocks` с явным возвратом `([]TestSessionDTO, []DownloadInfoDTO, Stats, error)`.
   - Реализовать FIFO-спаривание `(PCBA, StationType)` с учётом порядка в логе.
   - Размечать тип станции по имени скан-шага (`PCBA Scan` → `PCBA`; `Compare PCBA Serial Number`/`Valid PCBA Serial Number` → `Final`).
   - Fallback: если скан-шага нет, но первый шаг даёт косвенную подсказку — попробовать, иначе пометить сессию как `orphan_steps` с `StationType="Unknown"` и эмитировать WARN.
3. `grouper.go`: группировать `TestSessionDTO` по PCBA (или ProductSN), не разбивая шаги и станции.
4. `converter/` + services: принимать `TestSessionDTO`. Вставка одной сессии = Insert station record (если есть) → Insert steps с этим station_id (если есть). Если `Record=nil` — применять политику из Вопроса №1 ниже (drop vs синтезировать).
5. `dispatcher.go`: итерировать сессии, не параллельные слайсы.

### Фаза 2 — валидатор как реальный слой

Сейчас `validator.go` пуст. Наполняем:

- Группа валидна, если: есть либо `DownloadInfos`, либо `Sessions`.
- Каждая сессия валидна, если: `StationType != ""` и хотя бы одно из `Record`/`Steps` непусто.
- WARN-алерты (не фатальные):
  - Сессия-сирота без station record;
  - Сессия-сирота без шагов;
  - Дубль по `(pcba, station_type, test_finished_time)` в рамках одной группы (странно, но не повод падать);
  - PCBA пуст, группировка по ProductSN — пометить группу `is_phantom=true`.

Валидатор возвращает не `error`, а `ValidationReport{Warnings, HardErrors}`. Диспатчер решает, вставлять или пропустить.

### Фаза 3 — БД: идемпотентность и аудит

Миграция `003_idempotency_and_audit.sql`:

1. `ALTER TABLE ... ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()` во все четыре таблицы.
2. Денормализация `pcba_number` в `test_station_record` (дубликат из logistic_data) — для естественного UNIQUE и для запросов без JOIN. Бэкфил на существующих данных одноразовым `UPDATE ... FROM logistic_data`.
3. Уникальные индексы (частичные, чтобы не блокировать существующие дубли):
   - `UNIQUE (tcu_pcba_number, download_finished_time)` на `download_info`;
   - `UNIQUE (pcba_number, test_station, test_finished_time)` на `test_station_record`;
   - `UNIQUE (pcba_number, product_sn)` на `logistic_data` (ловим дубли 1:1);
   - `test_step` — без констрейнта (дубли шагов нормальны при retry), но поверх всё равно `created_at`.
4. Новая таблица `processed_files` для трекинга:
   ```sql
   CREATE TABLE processed_files (
       id SERIAL PRIMARY KEY,
       file_path TEXT NOT NULL,
       file_hash TEXT NOT NULL UNIQUE,
       file_size BIGINT NOT NULL,
       status TEXT NOT NULL, -- 'success' | 'partial' | 'failed'
       summary JSONB NOT NULL,
       started_at TIMESTAMPTZ NOT NULL,
       finished_at TIMESTAMPTZ NOT NULL
   );
   ```
   CLI перед обработкой файла смотрит `processed_files.file_hash` — если уже `success`, пропускает. Если `partial` — предлагает `--force` для повторной обработки с ON CONFLICT DO NOTHING.

Все `Insert*`-репозитории переходят на `ON CONFLICT DO NOTHING RETURNING id`, с возвратом id существующей записи при конфликте (через отдельный SELECT в случае пустого RETURNING).

### Фаза 4 — транзакции

- Каждая группа (PCBA) — одна транзакция. Падение на сессии откатывает всю группу. Следующая группа обрабатывается отдельно.
- Файл целиком — не транзакция (слишком крупно, долгие блокировки).

### Фаза 5 — структурированное логирование и сводка по файлу

1. Унифицировать лог-поля: `event`, `file`, `file_hash`, `pcba`, `product_sn`, `station_type`, `log_seq`, `reason`, `count`.

2. Ключевые события:
   - `file_start` — входим в файл, хеш посчитан.
   - `block_dropped` — блок не прошёл фильтр (DEBUG-уровень).
   - `session_emitted` — парсер выдал сессию (TRACE).
   - `session_orphan_steps` — шаги без станции (WARN).
   - `session_orphan_station` — станция без шагов (INFO/WARN).
   - `retry_detected` — вторая+ попытка на станции для PCBA.
   - `group_failed` — группа упала в диспатчере (WARN).
   - `file_summary` — агрегат по итогам файла (INFO).

3. **`file_summary`** — структура JSON для ops-мониторинга:

   ```json
   {
     "event": "file_summary",
     "file": "mesrestapi.log-20260411.gz",
     "file_hash": "…",
     "duration_ms": 1340,
     "blocks": {"total": 3702, "download": 419, "pcba_station": 421,
                "final_station": 386, "pcba_steps": 421, "final_steps": 386,
                "other": 1669, "dropped": 0},
     "pcbas_grouped": 446,
     "sessions": {"full": 800, "orphan_steps": 3, "orphan_station": 12, "retries": 3},
     "inserted": {"logistic_data": 446, "test_station_record": 812,
                  "test_step": 34379, "download_info": 419},
     "skipped_duplicates": {"logistic_data": 0, "test_station_record": 0},
     "groups": {"ok": 445, "failed": 1},
     "status": "partial"
   }
   ```

   Копия этой структуры сохраняется в `processed_files.summary` — можно строить исторические дашборды прямо из БД.

4. Уровни:
   - TRACE/DEBUG — поблочные решения.
   - INFO — file_start / file_summary / retries (выборочно).
   - WARN — orphans, group_failed, phantom PCBA.
   - ERROR — IO/DB ошибки.
   - FATAL — только запуск (конфиг, БД).

5. Exit-коды CLI:
   - 0 — всё ок.
   - 2 — были warn-события (partial processing), но CLI дошёл до конца.
   - 1 — fatal (не удалось подключиться к БД / разобрать конфиг / ни один файл не обработан).

### Фаза 6 — тесты

1. **Unit**:
   - `ParseLogBlocks` на синтетических входах: нормальная сессия, orphan_steps, orphan_station, retry с разной асимметрией, пустой PCBA → ProductSN, смесь.
   - Валидатор на тех же входах.

2. **Integration** (в `tests/integration/`):
   - Реальный PostgreSQL (в Docker), парсим синтетический `.gz` с 5–10 PCBA разных форм.
   - Проверяем содержимое всех четырёх таблиц после прогона.
   - Повторный прогон — количество строк не меняется.
   - Модифицируем файл (другой хеш) — обработка идёт заново.

3. **Regression fixture**:
   - Мини-файл, вырезанный из `mesrestapi.log-20260416.gz` (файл с максимальным числом проблемных PCBA), в который положены 3–4 «неудобных» PCBA. Добавить в репо как `tests/fixtures/regression_20260416.gz`. Любая регрессия ловится.

### Фаза 7 — API (минимум)

- `/api/v1/pcba?pcbanumber=X` должен возвращать **все** сессии по PCBA (сейчас, судя по коду, возвращает первую или агрегат — уточнить после фазы 3).
- Новый эндпоинт `/api/v1/files` — список обработанных файлов из `processed_files` для мониторинга.
- Эндпоинт `/api/v1/pcba/orphans` — список PCBA с неполными сессиями (для QA и завода — возможность поднять, почему станция не записалась).

---

## 4. Риски и митиграции

| Риск | Митигация |
|---|---|
| Миграция 003 на живой БД: бэкфил `pcba_number` в `test_station_record` может быть долгим | Делать в off-peak, либо в две миграции: 003a — добавить колонку nullable + триггер, 003b — заполнить и сделать NOT NULL + UNIQUE |
| Уникальные индексы отвергают существующие дубли | Частичные индексы `WHERE created_at > '2026-04-22'` на переходный период + отдельный скрипт дедупа перед полной активацией |
| Смена формы `GroupedDataDTO` затрагивает API-обработчики | Проверить, что DTO не сериализуется «как есть» в API (судя по `swagger_models.go` — свой слой, ок) |
| Логика «сирот» потеряет данные, если выбрать drop | Вопрос №1 ниже — нужен ответ до реализации |

---

## 5. Решения по открытым вопросам (зафиксированы)

Принятые ответы по итогам обсуждения — именно так будем кодить:

1. **Сессии-сироты (шаги без station record) → синтезировать минимальную запись.**
   Механика:
   - Создаём `TestStationRecordDTO` с `TestStation = inferredType`, `IsAllPassed = false`, `ErrorCodes = "SYNTH_NO_STATION"`.
   - `LogisticData` **заимствуем** у соседней session в этой же группе (например, если пропала PCBA-запись — берём `LogisticData` из Final-записи того же PCBA).
   - Если соседней session нет (редкий случай как `T32444111`) — шаги **пропускаем** с WARN `sessions_lost_no_logistic_data`; по этому событию завод может разбираться, почему станок вообще ничего не прислал.
   - В миграции 003 добавляется колонка `is_synthetic BOOLEAN NOT NULL DEFAULT FALSE` на `test_station_record`. Синтетические пишутся с `is_synthetic = TRUE` — чтобы отчётность могла отличить их от настоящих.

2. **Сессии-сироты (station record без шагов) — вставляем как есть** (station record + пустой набор шагов). Валидное событие.

3. **Phantom PCBA (пустой `PCBANumber`, ProductSN вместо него) → пропускать с WARN.**
   Парсер ещё до группировки эмитит `station_skipped_empty_pcba` c полями `station_type`, `product_sn`, `error_codes`. В БД не попадают. Отдельной `early_failures` таблицы в этой итерации не делаем — если потом понадобится, добавим.

4. **Идемпотентность → row-level `ON CONFLICT DO NOTHING`, без блокирующего file-hash.**
   Таблица `processed_files` остаётся как **аудит-лог** (хранит `summary` каждого прогона), но она не блокирует повторный запуск. Повторный прогон безопасен: существующие строки сохраняются, пропущенные добавляются. `--force` не нужен.

5. **Retry на одной и той же станции** — каждая попытка отдельной строкой в `test_station_record` со своим `test_finished_time`. FIFO-сшивка с массивами шагов по порядку появления в логе.

6. **Миграция 003 с денормализацией `pcba_number`** — принято. Конкретика:

   **До:** `test_station_record` ссылается на `logistic_data` через FK `logistic_data_id`, `pcba_number` живёт только в `logistic_data`.

   **После:**
   ```sql
   test_station_record (
     ...существующие колонки...,
     logistic_data_id INTEGER NOT NULL REFERENCES logistic_data(id),
     pcba_number    TEXT        NOT NULL,     -- NEW: дубль из logistic_data
     created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- NEW
     is_synthetic   BOOLEAN     NOT NULL DEFAULT FALSE   -- NEW (см. п.1)
   )
   CREATE UNIQUE INDEX test_station_record_natural_key
     ON test_station_record(pcba_number, test_station, test_finished_time)
     WHERE is_synthetic = FALSE;
   ```

   Шаги миграции:
   1. `ADD COLUMN pcba_number TEXT` (nullable).
   2. `UPDATE test_station_record SET pcba_number = l.pcba_number FROM logistic_data l WHERE logistic_data_id = l.id`.
   3. `ALTER COLUMN pcba_number SET NOT NULL`.
   4. `ADD COLUMN created_at`, `ADD COLUMN is_synthetic`.
   5. `CREATE UNIQUE INDEX` частичный (без синтетики, чтобы несколько синтетических записей для одного PCBA не конфликтовали).

   Репо-слой в `TestStationRecordRepository.Insert` добавляет `pcba_number` в INSERT. API, использующее `GetByPCBANumber`, сможет делать прямой SELECT без JOIN.

7. **Фаза 0 (hotfix) — сегодня отдельным PR**, всё остальное — последующими. Hotfix в коде: изменить `return err` на `WARN + continue` в `DispatchGroups`. Ничего больше.

8. **Эскалация заводу** — документ `docs/factory-escalation.md` содержит:
   - точное описание того, какой HTTP-запрос (`POST /v1/stationinformation`) не дёргается;
   - таблицу затронутых PCBA по файлам;
   - ожидаемый нормальный протокол;
   - список вопросов станочной команде.

9. **`validator.go` пуст** — да, интерпретируем как «надо реализовать». Наполняем в Фазе 2 как описано ниже.

10. **`TestMeasuredValue interface{}`** — оставляем, в БД всё равно хранится строкой.

---

### Детализация Фазы 2 (валидатор)

```go
type ValidationReport struct {
    Warnings   []ValidationIssue
    HardErrors []ValidationIssue
}

type ValidationIssue struct {
    GroupKey string                     // PCBA или productsn:<SN>
    Kind     string                     // см. список ниже
    Message  string
    Context  map[string]interface{}
}

func Validate(groups []dto.GroupedDataDTO) ValidationReport
```

`Kind` — перечисление:

- `orphan_steps_synthesized` — для шагов без записи синтезирована станция (WARN).
- `orphan_steps_dropped_no_logistic` — шаги отброшены, т.к. заимствовать LogisticData не у кого (WARN, помечается в `file_summary`).
- `retry_asymmetry_paired_fifo` — retry-асимметрия сшита по FIFO (WARN, диагностика).
- `phantom_pcba_skipped` — пропущена запись с пустым PCBA (WARN).
- `empty_group` — в группе нет ни Download, ни Sessions (HardError, крайне редкое).
- `session_with_unknown_station_type` — не смогли вывести тип (HardError).

`HardErrors` → группа пропускается с ERROR. `Warnings` → проходит в диспатчер, попадают в `file_summary.warnings[]` и в `processed_files.summary`.

---

## 6. Оценка объёма

| Фаза | Часы |
|---|---:|
| 0 (hotfix) | 1–2 |
| 1 (новый контракт + парсер) | 6–8 |
| 2 (валидатор) | 2–3 |
| 3 (миграции + идемпотентность репо) | 4–6 |
| 4 (транзакции) | 1–2 |
| 5 (структурированные логи + сводка) | 3–4 |
| 6 (тесты + fixture) | 4–6 |
| 7 (API endpoints) | 2–3 |
| **Итого** | **23–34 ч** |

Предложение: Фаза 0 сегодня, Фазы 1+2+4 завтра, Фазы 3+5+6 послезавтра-следующий, Фаза 7 — отдельным PR позже.
