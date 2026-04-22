# Эскалация на завод: пропущенный `POST /v1/stationinformation`

## TL;DR

На свежих логах (`mesrestapi.log-20260411/15/16/17`) для **1–2 % устройств** со стороны станков (PCBA-тестер `Pandora-PCBA-01` и Final-тестер `Pandora-CP-01`) **перестал приходить** один из двух стандартных вызовов:

```
POST /v1/stationinformation   HTTP/1.1
Content-Type: application/json

{ "TestStation": "PCBA",  "LogisticData": { "PCBANumber": "<S/N>", ... }, ... }
```

или, зеркально, его Final-версия:

```
POST /v1/stationinformation   HTTP/1.1
Content-Type: application/json

{ "TestStation": "Final", "LogisticData": { "PCBANumber": "<S/N>", ... }, ... }
```

При этом **второй** вызов от того же станка — `POST /v1/testdatas` с массивом тест-шагов этого же устройства — приходит **нормально**. То есть станок провёл тест и отправил результаты шагов, но саму «обёртку» со StationInformation (метаданные сессии) не отправил.

Скорее всего — регрессия после последнего обновления ПО на стороне станков. Нам нужна ваша помощь проверить, почему для части устройств `/v1/stationinformation` не дёргается.

---

## Ожидаемый протокол (как было раньше и как должно быть)

Для каждого устройства, которое проходит станцию, от станка прилетает **пара** запросов в определённом порядке:

```
1) POST /v1/stationinformation    — метаданные сессии (тип станции, тестер, время, PCBA, LogisticData)
2) POST /v1/testdatas             — массив отдельных тест-шагов этой сессии
```

(Порядок может быть обратный в зависимости от станка, но оба должны быть на каждое устройство, в рамках одного рабочего цикла устройства на станции.)

Для каждого PCBA в нормальном производственном цикле (прошивка → PCBA-тестер → Final-тестер) это даёт пять запросов суммарно:

```
POST /v1/stationinformation   { "TestStation": "Download", "TcuPCBANumber": "<S/N>", ... }
POST /v1/testdatas            [ ..., { "TestStepName": "PCBA Scan",                 "TestMeasuredValue": "<S/N>" }, ... ]
POST /v1/stationinformation   { "TestStation": "PCBA",     "LogisticData": { "PCBANumber": "<S/N>", ... }, ... }
POST /v1/testdatas            [ ..., { "TestStepName": "Compare PCBA Serial Number","TestMeasuredValue": "<S/N>" }, ... ]
POST /v1/stationinformation   { "TestStation": "Final",    "LogisticData": { "PCBANumber": "<S/N>", ... }, ... }
```

## Что мы сейчас видим в логах

### Пример 1. `H8444A11100T32343298` (файл `mesrestapi.log-20260411.gz`, 2026-04-10)

```
11:51:47   POST /v1/stationinformation   { "TestStation": "Download", ... }       ← пришло
12:09:35   POST /v1/testdatas            [ "PCBA Scan" / T32343298, ... ]         ← пришло
           ↓ ждём: POST /v1/stationinformation  { "TestStation": "PCBA",  ... }
           ↓ НЕ ПРИШЛО от Pandora-PCBA-01
17:48:51   POST /v1/testdatas            [ "Compare PCBA Serial Number" / T32343298, ... ] ← пришло
17:48:54   POST /v1/stationinformation   { "TestStation": "Final", ... }          ← пришло
```

Итог: 4 запроса вместо 5. Отсутствует `POST /v1/stationinformation` типа `PCBA` от Pandora-PCBA-01.

### Пример 2. `H8444A11100T32444508` (файл `mesrestapi.log-20260416.gz`, 2026-04-15) — зеркально

```
07:58:37   POST /v1/stationinformation   { "TestStation": "PCBA",  ... }          ← пришло
           POST /v1/testdatas            [ "PCBA Scan" / T32444508, ... ]         ← пришло
14:23:19   POST /v1/testdatas            [ "Valid PCBA Serial Number" / T32444508, ... ] ← пришло
           ↓ ждём: POST /v1/stationinformation  { "TestStation": "Final", ... }
           ↓ НЕ ПРИШЛО от Pandora-CP-01
```

Итог: отсутствует `POST /v1/stationinformation` типа `Final` от Pandora-CP-01.

### Пример 3. `H8444A11100T32444111` (файл `mesrestapi.log-20260415.gz`, 2026-04-14) — крайний случай

```
15:04:52   POST /v1/testdatas            [ "PCBA Scan" / T32444111, ... ]         ← пришло
           (больше ничего от этого устройства в файле нет)
```

Ни одного `POST /v1/stationinformation` для этого PCBA. Не пришёл ни Download, ни PCBA, ни Final. Возможно, устройство было протестировано только частично и цикл прервался до следующих станций — но минимум `POST /v1/stationinformation { "TestStation": "PCBA", ... }` после `POST /v1/testdatas` с PCBA Scan должен был прийти.

---

## Затронутые устройства (из 4 файлов)

**Нет `POST /v1/stationinformation { "TestStation": "PCBA", ... }`** (хотя `testdatas` с PCBA Scan пришёл):

- `H8444A11100T32343298` (2026-04-10, Pandora-PCBA-01)
- `H8444A11100T32443984` (2026-04-15, Pandora-PCBA-01)
- `H8444A11100T33147078` (2026-04-16, Pandora-PCBA-01)

**Нет `POST /v1/stationinformation { "TestStation": "Final", ... }`** (хотя `testdatas` с `Compare PCBA Serial Number` / `Valid PCBA Serial Number` пришёл):

- `H8444A11100T32444508` (2026-04-15, Pandora-CP-01)
- `H8444A11100T33148238` (2026-04-15, Pandora-CP-01)
- `H8444A11100T40148613` (2026-04-15, Pandora-CP-01)
- `H8444A11100T33146841` (2026-04-15, Pandora-CP-01)
- `H8444A11100T33146763` (2026-04-15, Pandora-CP-01)

**Ноль `POST /v1/stationinformation` (ни одного типа)** — пришли только `testdatas`:

- `H8444A11100T32444111` (2026-04-14, Pandora-PCBA-01)

Масштаб: **9** устройств из **~596** (~1.5%) в одном файле. Суммарно по четырём файлам — 9 устройств с пропущенным `stationinformation` POST'ом.

---

## Что хотелось бы узнать

1. Есть ли у вас на стороне станков лог исходящих HTTP-запросов? Если да — для перечисленных PCBA проверьте, был ли `POST /v1/stationinformation` собственно сформирован и отправлен, или он упал до отправки (DNS, таймаут, сериализация).
2. Если он не формировался — при каком условии станок его пропускает? (Например, если устройство зафейлилось на определённом шаге до конца сессии.)
3. Это поведение появилось после недавнего обновления ПО станка? (На старых логах из `corporate_resources/` за более ранние даты такого почти не было.)
4. Можно ли на стороне станка добавить ретраи на `POST /v1/stationinformation`? (Сейчас у нас нет способа понять, это баг передачи или намеренный пропуск.)

---

## Что будем делать мы со своей стороны

Параллельно с вашим разбором мы дорабатываем парсер так, чтобы:

- 1 такой пропуск больше **не ронял обработку всего файла** (сейчас роняет, из-за чего около 1500 валидных устройств не попадают в БД).
- Для устройств с пропущенным `stationinformation` мы **синтезируем минимальную запись** в БД (с флагом `is_synthetic=true`), чтобы сохранить привязанные шаги. Если даже вы починитесь на стороне станка, эти строки у нас останутся — можно будет посчитать, как часто это случалось в переходный период.

Детали — в `docs/problem.md` и `docs/fix-plan.md` нашего репозитория.
