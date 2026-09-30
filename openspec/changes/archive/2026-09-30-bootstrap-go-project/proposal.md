## Why

Кода приложения в репозитории нет: собрано только окружение разработки. Первое изменение цепочки
реализации должно дать собираемый проект Extruder — без него нечем проверить ни одно последующее
изменение: приёмочному прогону не во что компилироваться, а целям `go-build`, `go-test` и `go-lint`
нечего собирать.

Изменение открывает этап «Подготовка» в цепочке из восьми изменений — порядок и место изменения
в цепочке описывает [`project-plan.md`](../../../docs/project-plan.md#порядок-реализации).

## What Changes

- заводится go-модуль `github.com/united-software-platform/extruder`, строка версии языка —
  `go 1.27`, без директивы `toolchain`: патч-версия остаётся свойством машины
  ([DEC-040](../../../docs/extruder-design.md#устройство-extruder));
- появляется точка входа `cmd/extruder`: печатает строку версии инструмента и завершается кодом
  `0`; строка версии живёт здесь же, технического пакета вне модулей раскладка не заводит
  ([DEC-039](../../../docs/extruder-design.md#устройство-extruder));
- закладывается каркас пакетов двух модулей самого Extruder — `profile` и `generation`, в каждом
  три слоя ([DEC-028](../../../docs/extruder-design.md#устройство-extruder),
  [DEC-039](../../../docs/extruder-design.md#устройство-extruder),
  [DEC-020](../../../docs/extruder-design.md#раскладки));
- добавляется конфигурация линтера `.golangci.yml`
  ([DEC-038](../../../docs/extruder-design.md#устройство-extruder));
- появляется первый тест — на строку версии: цель `go-test` перестаёт быть целью без предмета.

Логики опроса, плана генерации, модели проекта, рендера и записи изменение не содержит: они
принадлежат изменениям `add-acceptance-run`, `add-profile-survey`, `build-generation-pipeline`
и `add-atomic-write`.

## Capabilities

### New Capabilities

- `extruder-cli`: инструмент как исполняемый файл — сборка в единственный бинарник и его запуск.
  Поведение опроса, генерации и записи в эту capability на этом изменении не входит: оно
  добавляется последующими изменениями цепочки.

### Modified Capabilities

Нет: `openspec/specs/` пуст, изменение вводит первую capability.

## Impact

- **Новые файлы проекта:** `go.mod`, `cmd/extruder/main.go`, `cmd/extruder/main_test.go`,
  `internal/profile/{domain,application,infrastructure}/doc.go`,
  `internal/generation/{domain,application,infrastructure}/doc.go`, `.golangci.yml`.
- **Файлы окружения не правятся:** `Makefile`, `tools/host-runner/runner.py` и `README.md`
  приведены задачей `EXT-006` заранее — цели `go-build`, `go-test`, `go-lint`, `go-tools`
  и каталоги инструментов уже на месте
  ([DEC-033](../../../docs/extruder-design.md#проверки-и-поставка)).
- **Проверка:** на хосте через раннер — `tools/host-runner/call.sh go-build`, `go-test`,
  `go-lint`; линтер ставится целью `go-tools`. Без запущенного раннера изменение непроверяемо.
- **Внешние зависимости:** ни одной. Первая зависимость модуля появится не раньше, чем её
  потребует последующее изменение; опрос будет собран на стандартной библиотеке
  ([DEC-041](../../../docs/extruder-design.md#опрос)).
- **Каталог `bin/`** уже игнорируется git, артефакт сборки в историю не попадает.
