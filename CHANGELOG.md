# Changelog

## [1.1.0] — 2026-09-24

### Added
- Пакет `terminfo` — детект возможностей терминала (`Info`, `ColorBits`).
- Downsampling цветов: TrueColor → 256 → 16 → 8 под палитру терминала.
- `Terminal.DrawString(x, y, style, str)` — рисует строку с поддержкой `\n` и `\r\n`.
- `Terminal.Info()` — возвращает `terminfo.Info` текущего терминала.
- Оптимизация перемещения курсора: `\033[H` для (0,0), `\033[nC` для близких сдвигов вправо.
- Поддержка `Dim`, `Strike`, `Hidden` в `Style`.