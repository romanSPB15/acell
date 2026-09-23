# acell

[![Go Reference](https://pkg.go.dev/badge/github.com/romanSPB15/acell.svg)](https://pkg.go.dev/github.com/romanSPB15/acell)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://img.shields.io/badge/coverage-96%25-green)](./calc-coverage-noterm.ps1)
[![Test](https://github.com/romanSPB15/acell/actions/workflows/test.yaml/badge.svg)](https://github.com/romanSPB15/acell/actions/workflows/test.yaml)

**Низкоуровневый терминальный слой для Go.**

- 🚀 **~5100 FPS** на бенчмарке — в 2 раза быстрее tcell v3
- 🎨 Рисование напрямую в `[][]Cell` — без ANSI-строк на горячем пути
- 🎯 Полная поддержка мыши: клик, отпускание, движение, скролл (SGR 1006)
- ⌨ Полная поддержка клавиатуры: стрелки, F1–F12, Ctrl/Alt/Shift, Shift+Tab
- 🎁 Windows без WSL, без CGO
- 📦 Всего две зависимости — `x/sys` и `x/term`
- ✅ 96% покрытие ядра тестами

<h3 align="center"><pre>go get -u github.com/romanSPB15/acell</pre></h3>

## Быстрый старт

```go
package main

import (
    "time"
    "github.com/romanSPB15/acell"
)

func main() {
    in, out := acell.Default()
    term := acell.New(in, out)
    defer term.Close()

    // term.Buf уже размечен под размер терминала и заполнен пробелами.
    term.Buf[0][0] = acell.Cell{Char: 'H', Style: acell.Style{Fg: "32", Args: acell.Bold}}
    term.Buf[0][1] = acell.Cell{Char: 'i'}
    term.Flush()

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case ev := <-term.Events():
            switch e := ev.(type) {
            case *acell.KeyboardEvent:
                if e.Rune == 'q' || e.Key == acell.KeyCtrlC {
                    return
                }
            case *acell.MouseEvent:
                if e.Action == acell.MousePress {
                    term.Buf[e.Pos.Y][e.Pos.X] = acell.Cell{Char: 'X'}
                    term.Flush()
                }
            case *acell.ResizeEvent:
                term.Buf = acell.NewBuf(e.Width, e.Height)
                term.Flush()
            }
        case <-ticker.C:
            // перерисовка ...
            term.Flush()
        }
    }
}
```

## Как это работает

Модель рендера — буфер `[][]Cell` (`Terminal.Buf`) плюс теневая копия
(`oldBuf`). Каждый `Flush`:

1. Проходит по обоим буферам клетка за клеткой.
2. Для каждой изменившейся клетки пишет `CUP` (только если курсор ещё
   не там), затем дельту стиля, затем руну.
3. Кеширует предыдущий стиль и позицию курсора, чтобы не писать лишние
   последовательности.
4. Оборачивает кадр в mode 2026 (`CSI ?2026h` … `CSI ?2026l`) для
   атомарного обновления.

Весь вывод собирается в переиспользуемый `builder.Builder` — без промежуточных строк.

## Производительность

Стресс-бенчмарк — терминал 120×30, виджет 80×24, 300 изменяющихся клеток
на кадр, Windows 10 x64, Windows Terminal, с I/O:

| Реализация                            | Raw     | С виджетами         |
|---------------------------------------|---------|---------------------|
| **acell**                             | ~5100   | —                   |
| **tui-compose v4** (acell + виджеты)  | —       | ~2300               |
| raw tcell v3                          | ~2450   | —                   |
| metaspartan/gotui (tcell + виджеты)   | —       | ~1100               |


## Пакеты

| Пакет     | Назначение                                       |
|-----------|--------------------------------------------------|
| `acell`   | `Terminal`, `Cell`, `Style`, diff-рендер         |
| `ansi`    | Парсинг ANSI-последовательностей: `Strip`, `Find`|
| `builder` | Аналог `strings.Builder` с расширенным API       |
| `term`    | `RawTerminal` и работа с терминалом              |

## Покрытие тестами

- Ядро: **96.2%**
- С учётом терминального слоя: 58.9%

```
./calc-coverage.ps1          # полное покрытие, включая term
./calc-coverage-noterm.ps1   # только ядро
```

## Лицензия

[MIT](./LICENSE)