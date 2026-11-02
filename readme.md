# taskmenu

Консольное меню для [`Taskfile.yml`](https://taskfile.dev/). Показывает список задач и позволяет запускать их по номерам, в любом порядке и по несколько за раз.

```
1 build Build the project
2 test Run tests
3 lint Run linters
4 deploy Deploy to production
> 3 2

task: [lint] golangci-lint run
task: [test] go test ./...

Done in 2.341s
Press Enter to continue...
```

## Возможности

- Список задач берётся из `task --list-all --json` — то же, что видит `task --list-all`.
- Можно ввести сразу несколько номеров: `4 2 1` → `task task-4 task-2 task-1`.

## Установка

1. Разместить `taskmenu` в ту же папку, где лежит `task`, либо в любое место, которое есть в `PATH`. Утилита вызывает `task` из `PATH`, так что он должен быть доступен.
2. Запускать из папки с `Taskfile.yml`.

### Требования

- [`task`](https://taskfile.dev/installation/) должен быть установлен и доступен в `PATH`.
- `Taskfile.yml` — в текущей рабочей директории.

## Использование

```
taskmenu              запустить интерактивное меню
taskmenu -h|--help    показать справку
taskmenu -v|--version показать версию
```

В меню:

- Введите **один или несколько номеров** через пробел — задачи выполнятся по порядку.
- После выполнения — нажмите **Enter**, чтобы вернуться к меню.
- **Ctrl+C** — выход из программы.

### Пример

```
$ taskmenu
1 build Build the project
2 test Run tests
3 lint Run linters
4 deploy Deploy to production
> 3 2
```

Запустит `task lint test`.