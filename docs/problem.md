# log-parser: полный разбор проблемы

Документ описывает, почему CLI `go run ./cmd/cli/main.go corporate_resources/` падает на свежих `.gz`-файлах, что именно ломается, какие конкретные устройства задеты и какие логи это теперь отражают. Это справочник: открываешь, ищешь WARN из CLI — находишь объяснение.

План фикса — в `docs/fix-plan.md`. Здесь только диагноз.

---

## Содержание

1. [TL;DR](#tldr)
2. [Ожидаемая модель одного устройства](#ожидаемая-модель-одного-устройства)
3. [Группа багов №1 — пропущенная `StationInformation` запись](#группа-багов-1--пропущенная-stationinformation-запись)
4. [Группа багов №2 — retry-асимметрия (несовпадение счёта внутри типа)](#группа-багов-2--retry-асимметрия-несовпадение-счёта-внутри-типа)
5. [Группа багов №3 — `ProductSN` вместо пустого `PCBANumber`](#группа-багов-3--productsn-вместо-пустого-pcbanumber)
6. [Каскад: почему один битый PCBA роняет весь файл](#каскад-почему-один-битый-pcba-роняет-весь-файл)
7. [Доказательства на реальных данных](#доказательства-на-реальных-данных)
8. [Что проверили и что не сломано](#что-проверили-и-что-не-сломано)
9. [Сводная таблица по файлам](#сводная-таблица-по-файлам)
10. [Какие логи добавлены и где их искать](#какие-логи-добавлены-и-где-их-искать)

---

## TL;DR

- **Симптом.** CLI падает с `mismatch in TestSteps count and TestStationRecords count for PCBA <X>`. Одна такая ошибка обрывает обработку всего файла — валидные устройства «после» в БД не попадают.
- **Первопричина.** В модели данных нет сущности «сессия» (запись станции + её шаги как одно целое). Парсер держит массивы шагов и записи станций в параллельных слайсах, сопоставляет их по индексу и только по номеру PCBA — без учёта типа станции. Живой лог этот инвариант нарушает.
- **Две группы багов** сливаются в один симптом:
  1. **№1** — StationInformation записи для какого-то типа (PCBA или Final) в логе **нет**, хотя шаги этого типа пришли. Частично вина завода (станок после обновления иногда не дёргает `/v1/stationinformation`), частично наша (не умеем это пережить).
  2. **№2** — StationInformation и шаги **есть**, но их счёт не совпадает в рамках одного типа (например, `1 Final-запись + 2 Final-массива шагов` из-за повторного прогона). Это чистая наша архитектурная ошибка — счёт и не должен совпадать позиционно.
- **Масштаб.** На четырёх свежих файлах — 8 случаев №1 и 15 случаев №2. Из-за каскадного `return err` в диспатчере теряем в БД **≈1500** валидных устройств, чтобы защититься от **23** битых.
- **Сопутствующее №3.** Записи `StationInformation` с пустым `PCBANumber` (ранняя фейлящаяся электроника) попадают в БД под `ProductSN`-ключом, засоряя таблицу `logistic_data` «фантомными» PCBA. Не роняет обработку, но портит отчётность.

---

## Ожидаемая модель одного устройства

Нормальный жизненный цикл одного TCU в логе:

```
POST /v1/stationinformation   Data { "TestStation": "Download", "TcuPCBANumber": <PCBA>, ... }
POST /v1/testdatas            Data [ { "TestStepName": "PCBA Scan", "TestMeasuredValue": <PCBA>, ... }, ... ]
POST /v1/stationinformation   Data { "TestStation": "PCBA",  "LogisticData": { "PCBANumber": <PCBA>, ... }, ... }
POST /v1/testdatas            Data [ { "TestStepName": "Compare PCBA Serial Number", "TestMeasuredValue": <PCBA>, ... }, ... ]
POST /v1/stationinformation   Data { "TestStation": "Final", "LogisticData": { "PCBANumber": <PCBA>, ... }, ... }
```

Пять сообщений: 1 Download + 2 StationInformation (PCBA + Final) + 2 TestDatas (шаги PCBA + шаги Final). Жизненный цикл устройства растягивается на часы/сутки — разные блоки могут попадать в разные файлы ротации.

Соответствие между шагами и типом станции однозначно определяется именем первого скан-шага массива:

| TestStepName | Тип сессии |
|---|---|
| `PCBA Scan` | `PCBA` (читаем серийник впервые) |
| `Compare PCBA Serial Number` | `Final` (сверяем с ранее записанным) |
| `Valid PCBA Serial Number` | `Final` |

Эта функция вынесена в `internal/services/parser/classifier.go`, `InferStationTypeFromSteps()` — используется и в парсере, и в диспатчере.

---

## Группа багов №1 — пропущенная `StationInformation` запись

### Что происходит

Завод прислал массив тест-шагов (`Data [ ... ]` с `PCBA Scan` / `Compare PCBA Serial Number`), но соответствующую `StationInformation` этого типа (`Data { "TestStation": "PCBA" ... }` / `"Final"`) **не прислал**. В файле её нет в принципе.

### Два подвида

| Подвид | Что пришло | Что не пришло |
|---|---|---|
| **№1a** | Download + PCBA-шаги + Final-запись + Final-шаги | `StationInformation` типа `PCBA` |
| **№1b** | (Download +) PCBA-запись + PCBA-шаги + Final-шаги | `StationInformation` типа `Final` |

Редкий сверхслучай: **ноль** StationInformation вообще (`T32444111` в 20260415) — пришли только шаги. Парсер отбрасывает такие шаги с WARN ещё до группировки (orphan), в БД ничего не попадает, диспатчер не трогается.

### Затронутые PCBA (8 штук)

- **20260411** — `H8444A11100T32343298` (№1a).
- **20260415** — только `orphan_no_station`: `H8444A11100T32444111`.
- **20260416** — `T32443984` (№1a), `T32444508` / `T33148238` / `T40148613` / `T33146841` / `T33146763` (№1b).
- **20260417** — `H8444A11100T33147078` (№1a).

### WARN в новом логе

```
WARN Missing station record of inferred type (Bug #1 signature)
  fields:
    array_index:        930
    pcba:               "H8444A11100T32343298"
    inferred_type:      "PCBA"
    station_types_seen: ["Final"]
    step_count:         47
    reason:             "test steps of type 'PCBA' arrived for this PCBA, but the
                         station record of that type was not in the log. Only the
                         listed station_types_seen were present."
```

- `inferred_type` отсутствует в `station_types_seen` ⇒ запись-партнёр не пришла.
- Эмитится в `internal/services/parser/json_parser.go` (`ParseMixedJSONArray`, второй проход). Использует структуру `stationTypesSeen map[PCBA]map[type]bool` (новая; раньше была `allStations map[PCBA]Record` с перезаписью, из-за которой WARN был ложно-срабатывающим на все «нормальные» двухстанционные устройства).

### Сквозной след для `T32343298` (20260411)

```
WARN Missing station record of inferred type (Bug #1 signature)
  {array_index:930, pcba:H8444A11100T32343298, inferred_type:PCBA,
   station_types_seen:[Final], step_count:47, ...}

INFO Parser classification summary
  {download_payloads:419, station_pcba_payloads:421, station_final_payloads:386,
   step_arrays_total:807, steps_matched_same_type:804,
   steps_matched_different_type:1,       <-- ровно 1 битый массив шагов
   steps_orphan_no_station:0, steps_missing_pcba_scan:2, ...}

WARN Group has more step arrays than station records (will fail in dispatcher)
  {group_key:H8444A11100T32343298,
   cause:bug1_missing_station_record_of_type,
   missing_station_types:[PCBA], asymmetric_types:[],
   station_record_count:1, step_array_count:2,
   stations_by_type:{Final:1}, steps_by_type:{Final:1, PCBA:1}}

WARN Dispatcher mismatch: more step arrays than station records
  {pcba:H8444A11100T32343298, station_record_count:1, step_array_count:2,
   failing_array_index:1, failing_array_inferred:Final, failing_array_steps:38,
   group_snapshot:{
     stations:[{index:0, station_type:Final, finished_at:20260410174848,
                pcba:H8444A11100T32343298, all_passed:true}],
     step_arrays:[{index:0, inferred_type:PCBA, scan_pcba:H8444A11100T32343298, step_count:47},
                  {index:1, inferred_type:Final, scan_pcba:H8444A11100T32343298, step_count:38}]}}

ERROR Error processing file
  {error:"failed to dispatch groups: mismatch ... for PCBA H8444A11100T32343298"}
```

---

## Группа багов №2 — retry-асимметрия (несовпадение счёта внутри типа)

### Что происходит

И `StationInformation`, и `Data [ ... ]` шагов для одного типа станции **есть**, но их счёт различается. Например, `1 Final-запись + 2 Final-массива шагов`: устройство Final-тестировали дважды, а `StationInformation` пришёл один раз (или наоборот — больше записей, чем массивов шагов).

В коде это ловит только группировщик (`len(TestSteps) > len(TestStationRecords)`). Парсер про это не знает — на его уровне «станция того типа у этого PCBA есть, значит шаги ок».

### Шаблоны, которые реально встречаются

| Шаблон | Смысл |
|---|---|
| `1 станция + 2 массива шагов` одного типа | Устройство прогнано дважды, `StationInformation` POST пришёл один раз |
| `2 станции + 3 массива шагов` | Два полных прогона + третий массив без пары |
| `3 станции + 4 массива шагов` | Три записи (первые две — `P013`, третья passed), четыре `testdatas` POST'а |

### Затронутые PCBA (15 штук)

- **20260415** — `T32646421` (1 PCBA + 2 PCBA-шагов), `T32645382` (3 PCBA + 4 PCBA-шагов).
- **20260416** (9) — `T32444267`, `T32443807`, `T33146990`, `T32443821`, `T32646747`, `T32443833`, `T32444542`, `T32444502`, `T40148920`. Большинство — шаблон `1 Final + 2 Final-шагов`.
- **20260417** (4) — `T32646736`, `T33147003`, `T33146892`, `T33146845`.

**№2 встречается почти вдвое чаще, чем №1.** Именно из-за этой группы падает 20260416/20260417 даже если бы мы починили №1 в изоляции.

### Сквозной след для `T32645382` (20260415)

Парсер молчит (запись типа PCBA есть, классификатор не видит асимметрии):

```
INFO Parser classification summary
  {..., steps_matched_same_type:822, steps_matched_different_type:0,
   steps_orphan_no_station:1, ...}
```

Группировщик классифицирует:

```
WARN Group has more step arrays than station records (will fail in dispatcher)
  {group_key:H8444A11100T32645382,
   cause:retry_asymmetry_same_type,
   missing_station_types:[], asymmetric_types:[PCBA],
   station_record_count:3, step_array_count:4,
   stations_by_type:{PCBA:3}, steps_by_type:{PCBA:4}}
```

Диспатчер с полным снапшотом:

```
WARN Dispatcher mismatch: more step arrays than station records
  {pcba:H8444A11100T32645382, station_record_count:3, step_array_count:4,
   failing_array_index:3, failing_array_inferred:PCBA,
   group_snapshot:{
     stations:[
       {index:0, station_type:PCBA, finished_at:20260414054408, error_codes:P013, all_passed:false},
       {index:1, station_type:PCBA, finished_at:20260414064358, error_codes:P013, all_passed:false},
       {index:2, station_type:PCBA, finished_at:20260414064904, error_codes:"",   all_passed:true}],
     step_arrays:[
       {index:0, inferred_type:PCBA, step_count:47},
       {index:1, inferred_type:PCBA, step_count:47},
       {index:2, inferred_type:PCBA, step_count:47},
       {index:3, inferred_type:PCBA, step_count:47}]}}
```

Три записи с разным `finished_at` (первые две `P013`, третья passed) — классический retry. Четыре массива шагов. Какой с какой станцией парный — только завод знает по таймстемпам; наш код этого не делает.

### Ещё один показательный: `T32646747` (20260416, микс)

```
WARN Group has more step arrays than station records
  {group_key:H8444A11100T32646747,
   cause:retry_asymmetry_same_type,
   asymmetric_types:[Final],
   stations_by_type:{Final:1, PCBA:1},
   steps_by_type:{Final:2, PCBA:1}}
```

Все типы есть, но Final-сессий прошло две, а `StationInformation Final` — один.

---

## Группа багов №3 — `ProductSN` вместо пустого `PCBANumber`

### Что происходит

Устройство фейлится рано (шаг `DUT Power On` или до PCBA-скана) — PCBA-серийник прочитать не успели. В логе прилетает `StationInformation` с `"PCBANumber": ""` и заполненным `"ProductSN": "YCOT1EBG...#"`. Код ошибки `F065`, `F202` и подобные.

Наш код в `internal/services/processor/grouper.go` делает фолбэк: если `PCBANumber` пуст, ключ группы = `ProductSN`. Это попадает в БД как строка в `logistic_data` с `pcba_number = "YCOT1EBG...#"`.

### Последствия

- **Не фатально** — обработка файла продолжается, диспатчер не подрывается.
- **Фоном портит данные**: в `logistic_data` создаются строки, где `pcba_number` — на самом деле ProductSN (не PCBA). API `/api/v1/pcbanumbers` отдаёт их как валидные PCBA, в отчётах они смешиваются с настоящими.

### Пример (из 20260411)

```
Apr 10 06:24:32 ... INFO ... Data {
  "TestStation": "Final",
  "IsAllPassed": false,
  "ErrorCodes": "F202",
  "LogisticData": {
    "PCBANumber": "",
    "ProductSN": "YCOT1EBG30900FB#",
    ...
  }
}
```

Группировщик в этом случае создаст группу с ключом `YCOT1EBG30900FB#`.

Этот случай **не** ломает обработку — задокументирован на будущее, чтобы при фиксе не забыть решить: (а) пропускать такие записи с WARN; (б) класть в отдельную таблицу `early_failures`; (в) оставить как есть с явным флагом. Рекомендация в `fix-plan.md` — (б).

---

## Каскад: почему один битый PCBA роняет весь файл

Баг не в том, что возникает несовпадение — оно естественно для живых логов. Баг в том, что код на него неадекватно реагирует. Четыре усиливающих фактора:

1. **`allStations` индексируется только по PCBA, без типа.**
   `internal/services/parser/json_parser.go`: `map[PCBA]TestStationRecordDTO`, перезаписывается на каждое новое вхождение. Когда для одного PCBA приходят и PCBA-, и Final-записи, в карте остаётся только последняя. Массивы шагов пропускаются в `results` по проверке «ключ в карте есть», тип не сверяется.

2. **`GroupedDataDTO` — параллельные слайсы.**
   `TestStationRecords []TestStationRecordDTO` и `TestSteps [][]TestStepDTO` ничем структурно не связаны. Подразумевается, что они парные и равной длины — но этот инвариант нигде не enforced.

3. **Диспатчер сшивает по индексу.**
   `internal/services/dispatcher/dispatcher.go`:
   ```go
   for i, stepsSlice := range group.TestSteps {
       if i >= len(testStationIDs) {
           return fmt.Errorf("mismatch...")
       }
       ...InsertTestSteps(..., testStationIDs[i])
   }
   ```
   Даже когда счёт сошёлся, привязка шагов к станции идёт **по позиции**, не по типу. PCBA-шаги могут молча залететь в Final-станцию — WARN `Dispatcher paired step array to station of different type` про это.

4. **Одна ошибка = отмена всего файла.**
   Любой `return err` из `DispatchGroups` отменяет уже собранные группы. Следующие ~500 валидных PCBA теряются вместе с одним сбоем.

Формально все три группы багов — это проявления одной архитектурной проблемы: **в модели данных нет понятия «сессия» (запись станции + её шаги как одно целое)**. Пока его нет, любое расхождение — фатальное.

---

## Доказательства на реальных данных

Чтобы не было «вы, может, что-то в коде неправильно считаете» — вот цитаты из сырых `.gz` для четырёх разных паттернов. Методика и скрипт воспроизводимости — в конце раздела.

### Proof A. `H8444A11100T32343298` в `mesrestapi.log-20260411.gz` — нет PCBA-записи

**Шаги PCBA-станции загружены** (строки 470008–470014):

```
470008: Apr 10 12:09:35 ... TestDatas Serving: /v1/testdatas
470009: Apr 10 12:09:35 ... Data  [
470010: Apr 10 12:09:35 ...  {
470011: Apr 10 12:09:35 ...    "TestStepName": "PCBA Scan",
470012: Apr 10 12:09:35 ...    "TestThresholdValue": "",
470013: Apr 10 12:09:35 ...    "TestMeasuredValue": "H8444A11100T32343298",
```

**Все `Inserting StationInformation` с этим PCBA** в файле:

```
Apr 10 17:48:54 ... Inserting StationInformation:  {0 703003734AA Final Tester_01 ... H8444A11100T32343298 ...}
```

Одна запись, тип `Final`. Количество PCBA-записей = **0**.

Вывод: устройство прошло PCBA-станцию в 12:09 (шаги в логе), Final-станцию в 17:48 (запись + шаги), но **`StationInformation` с `TestStation: "PCBA"` для него не отправлялся**.

### Proof B. `H8444A11100T32444111` в `mesrestapi.log-20260415.gz` — **ноль** записей станции

**Шаги PCBA-станции загружены** (строки 653092–653097):

```
653092: Apr 14 15:04:52 ... TestDatas Serving: /v1/testdatas
653093: Apr 14 15:04:52 ... Data  [
653094: Apr 14 15:04:52 ...  {
653095: Apr 14 15:04:52 ...    "TestStepName": "PCBA Scan",
653097: Apr 14 15:04:52 ...    "TestMeasuredValue": "H8444A11100T32444111",
```

**Все `Inserting StationInformation` с этим PCBA**: (пусто).

Ни одной записи. Парсер отбрасывает такие шаги как orphan, устройство не появляется в БД.

### Proof C. `H8444A11100T33147078` в `mesrestapi.log-20260417.gz` — тот же паттерн, что A

```
52894: Apr 16 06:21:37 ... TestDatas Serving: /v1/testdatas
52895: Apr 16 06:21:37 ... Data  [
52897: Apr 16 06:21:37 ...    "TestStepName": "PCBA Scan",
52899: Apr 16 06:21:37 ...    "TestMeasuredValue": "H8444A11100T33147078",
```

```
Apr 16 13:04:14 ... Inserting StationInformation:  {0 703003736AA Final Tester_01 ... H8444A11100T33147078 ...}
```

Только Final. PCBA-записи нет.

### Proof D. `H8444A11100T32444508` в `mesrestapi.log-20260416.gz` — зеркально: нет Final-записи

Шаги Final-станции (строки 589582–589584 в блоке):

```
589556: Apr 15 14:23:19 ... Data  [
589582: Apr 15 14:23:19 ...    "TestStepName": "Valid PCBA Serial Number",
589584: Apr 15 14:23:19 ...    "TestMeasuredValue": "H8444A11100T32444508",
```

`Valid PCBA Serial Number` — шаг Final-станции.

```
Apr 15 07:58:37 ... Inserting StationInformation:  {0 703003736AA PCBA Tester_01 ... H8444A11100T32444508 ...}
```

Только PCBA. Final-записи нет. Подвид №1b.

### Методика и воспроизведение

«Запись станции в логе есть» = в файле есть строка вида `INFO ... Inserting StationInformation: {0 ... PCBA|Final ... <PCBA_NUMBER> ...}` — это тот самый лог, который сервис пишет в момент фактического INSERT в `test_station_record`.

«Шаги станции в логе есть» = в файле есть `Data [ ... ]`-массив, **первый** сканирующий шаг которого (`PCBA Scan` → PCBA; `Compare PCBA Serial Number` / `Valid PCBA Serial Number` → Final) имеет `TestMeasuredValue` равный PCBA-номеру.

Скрипт `/tmp/prove_missing_stations.py` пробегает по всем четырём файлам и выдаёт список PCBA с расхождениями. Для ручной проверки любого PCBA:

```bash
gzcat corporate_resources/mesrestapi.log-<date>.gz | grep 'Inserting StationInformation' | grep <PCBA>
```

Если результат пуст или не содержит ни одного из двух типов — подтверждение №1.

---

## Что проверили и что не сломано

Перед тем, как утверждать «дело в №1/№2», мы прочёсали логи на другие возможные регрессии:

### Имена полей в payload'ах

Все ключи в `Data { ... }`-блоках и `Data [ ... ]`-массивах **точно** совпадают с полями DTO (`DownloadInfoDTO`, `TestStationRecordDTO`, `LogisticDataDTO`, `TestStepDTO`). **Ни одного переименования.** Сверяли `grep -oE '"[A-Za-z][A-Za-z0-9_]*"\s*:'` по всем файлам.

В логах также встречаются lowercase-ключи (`palletbarcode`, `xgroupbarcodes`, `userbarcode`, `trashstate`, `comment`, `elementbarcode` и т.п.) — это payload'ы других REST-эндпоинтов (`/v1/addpalletgroup`, `/v1/addelementtrash`). Они корректно отфильтровываются `FilterRelevantJsonBlocks` (нет поля `TestStation`), на наш парсинг не влияют.

### Значения `TestStation`

Ровно три: `Download`, `PCBA`, `Final`. Никаких новых типов не появилось.

### Значения `TestStepName`

В каждом файле 65 уникальных имён. Полное совпадение между 20260411 и 20260417 (diff пустой). Никаких новых или переименованных шагов.

### Типы `TestMeasuredValue`

Чаще всего строка, иногда число (`2`, `0`). Уже обработано — тип поля `interface{}`.

### Пустой файл

`mesrestapi.log-20260414.gz` — 0 байт. Парсер корректно логирует `failed to create gzip reader: EOF` и переходит к следующему. Отдельная проблема ротации на заводе, к коду отношения не имеет.

**Итог:** схема данных не дрейфовала. Причина падения — исключительно в том, как наш код сводит шаги и станции.

---

## Сводная таблица по файлам

Цифры из новых INFO-логов после `go run ./cmd/cli/main.go corporate_resources/`:

| Файл | Download | PCBA-станции | Final-станции | Массивов шагов | №1 | №2 | Всего mismatch-групп |
|---|---:|---:|---:|---:|---:|---:|---:|
| 20260411 | 419 | 421 | 386 | 807 | 1 | 0 | 1 |
| 20260414 | — пустой | — | — | — | — | — | — |
| 20260415 | 438 | 444 | 377 | 824 | 0 | 2 | 2 |
| 20260416 | 396 | 421 | 362 | 798 | 6 | 9 | 15 |
| 20260417 | 278 | 319 | 314 | 638 | 1 | 4 | 5 |
| **Итого** | | | | | **8** | **15** | **23** |

Устройств, которые падают **каскадом** после первого mismatch'а в каждом файле — порядка 200, 350, 550, 450. Суммарно это около **1500 валидных устройств**, которые мы сейчас не заливаем в БД ради защиты от 23 битых.

---

## Какие логи добавлены и где их искать

Все изменения — только логирование. Поведение парсинга и диспатчеризации **не менялось**; CLI падает ровно в тех же местах, но теперь с исчерпывающим контекстом.

### Новый файл `internal/services/parser/classifier.go`

Хелпер `InferStationTypeFromSteps(steps) (stationType, pcba)`. Вычисляет тип станции по первому сканирующему шагу и вытаскивает PCBA из его `TestMeasuredValue`. Используется и в парсере, и в диспатчере — чтобы вывод типа был согласованным.

### `internal/services/parser/json_parser.go`

Счётчики и финальный `INFO Parser classification summary` с полями:

- `download_payloads`, `station_pcba_payloads`, `station_final_payloads`, `station_records_with_empty_pcba`,
- `step_arrays_total`, `steps_matched_same_type`, `steps_matched_different_type`, `steps_orphan_no_station`, `steps_missing_pcba_scan`, `steps_unknown_infer`.

Структура `stationTypesSeen map[PCBA]map[type]bool`, чтобы различать реальный №1 («станции такого типа для этого PCBA не было вовсе») и шум от перезаписи `allStations` (массовое ложное срабатывание раньше — когда оба типа нормально приходили).

WARN'ы:

- `Missing station record of inferred type (Bug #1 signature)` — массив шагов типа X, в `stationTypesSeen[PCBA]` есть другие типы, но X — нет.
- `Test steps cannot be matched to any test station` — `stationTypesSeen[PCBA]` пуст (подвид «ноль записей»).
- `Test step array lacks PCBA identifier` (DEBUG) — сканирующего шага нет.
- DEBUG при перезаписи `allStations` для одного PCBA дважды.
- DEBUG при пустом `PCBANumber` в station record.

### `internal/services/processor/grouper.go`

Итоговый `INFO Grouping summary`:

- `groups_total`, `groups_with_download`, `groups_with_station`, `groups_with_steps`,
- `stations_by_type`, `step_arrays_by_type`, `groups_with_mismatch`.

WARN на каждую аварийную группу (`len(TestSteps) > len(TestStationRecords)`):

```
WARN Group has more step arrays than station records (will fail in dispatcher)
  group_key, station_record_count, step_array_count, has_download,
  stations_by_type, steps_by_type,
  cause, missing_station_types, asymmetric_types
```

Поле `cause` — это ключ для классификации:

- `bug1_missing_station_record_of_type` — есть типы шагов, для которых нет ни одной станции того же типа.
- `retry_asymmetry_same_type` — тип присутствует и в станциях, и в шагах, но счёт различается.
- `mixed` — одновременно и то, и другое.

### `internal/services/dispatcher/dispatcher.go`

Добавлено:

- `groupKey(group)` — возвращает лучший доступный идентификатор группы (PCBA → TcuPCBA → `productsn:<SN>` → `<unknown>`). Теперь ни одна ошибка не печатает пустой ключ.
- `describeGroup(group)` — компактный JSON-снапшот группы (станции с типами и `finished_at` + массивы шагов с `inferred_type`, `scan_pcba`, `step_count`). Попадает в каждое ключевое WARN/ERROR.
- `INFO Dispatch starting` / `INFO Dispatch finished` — границы работы.
- `DEBUG Dispatching group` на каждую группу.
- `WARN Dispatcher mismatch: more step arrays than station records` **перед** фатальным `return` — с полным `group_snapshot`.
- `WARN Dispatcher paired step array to station of different type` — когда счёт совпал, но типы нет (молчаливая раньше ситуация, когда PCBA-шаги уходили в Final-станцию).
- Все пути ошибок теперь логируют `ERROR` с `err` и структурированным контекстом.

### Конфиг

`configs/config.yaml` — `logger.level: info`. Все описанные WARN/INFO-саммари видны без переключения. Для DEBUG (поблочные решения, перезаписи map) — `level: debug`.

---

## Что дальше

Этот документ — диагноз. Фикс — в `docs/fix-plan.md`. Ключевое из плана:

1. Массивы шагов сопоставлять со станциями по паре **(PCBA, StationType)**, а не только по PCBA — закрывает №1.
2. В рамках одной (PCBA, StationType) пары счёт может быть неравным — делаем FIFO-сшивку по порядку появления в логе; необъединённые остатки либо оставляем как «сессию-сироту», либо синтезируем минимальную станцию. Это закрывает №2.
3. Ошибка в одной группе логируется и **не прерывает** обработку файла — спасает ~1500 валидных устройств, которые сегодня теряются каскадом.
4. Для №3 (ProductSN-фолбэк) — решение отдельное, в плане рекомендована отдельная сущность `early_failures`.
