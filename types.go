package acell

import (
	"github.com/romanSPB15/acell/input"
	"github.com/romanSPB15/acell/term"
)

// Типы событий и клавиатуры.
type (
	// Key — код клавиши (стрелки, F1–F12, Ctrl+буква и т.д.).
	Key = input.Key

	// KeyboardEvent — событие нажатия клавиши.
	KeyboardEvent = input.KeyboardEvent

	// MouseEvent — событие мыши: клик, движение, скролл.
	MouseEvent = input.MouseEvent

	// MouseAction — тип действия мыши (нажатие, отпускание, скролл).
	MouseAction = input.MouseAction

	// ResizeEvent — событие изменения размера окна терминала.
	ResizeEvent = term.ResizeEvent

	// Point — координаты (X, Y) в системе ячеек терминала.
	Point = input.Point

	// RawTerminal — низкоуровневый интерфейс терминала.
	RawTerminal = term.RawTerminal

	// WindowFocusEvent — событие фокуса окна терминала.
	WindowFocusEvent = input.WindowFocusEvent
)

// Функциональные клавиши F1–F12.
const (
	KeyF1  = input.KeyF1
	KeyF2  = input.KeyF2
	KeyF3  = input.KeyF3
	KeyF4  = input.KeyF4
	KeyF5  = input.KeyF5
	KeyF6  = input.KeyF6
	KeyF7  = input.KeyF7
	KeyF8  = input.KeyF8
	KeyF9  = input.KeyF9
	KeyF10 = input.KeyF10
	KeyF11 = input.KeyF11
	KeyF12 = input.KeyF12
)

// KeyUnknown — неизвестная клавиша.
const KeyUnknown Key = input.KeyUnknown

// Управляющие комбинации Ctrl+A … Ctrl+Z.
const (
	KeyCtrlA = input.KeyCtrlA
	KeyCtrlB = input.KeyCtrlB
	KeyCtrlC = input.KeyCtrlC
	KeyCtrlD = input.KeyCtrlD
	KeyCtrlE = input.KeyCtrlE
	KeyCtrlF = input.KeyCtrlF
	KeyCtrlG = input.KeyCtrlG
	KeyCtrlH = input.KeyCtrlH
	KeyCtrlI = input.KeyCtrlI
	KeyCtrlJ = input.KeyCtrlJ
	KeyCtrlK = input.KeyCtrlK
	KeyCtrlL = input.KeyCtrlL
	KeyCtrlM = input.KeyCtrlM
	KeyCtrlN = input.KeyCtrlN
	KeyCtrlO = input.KeyCtrlO
	KeyCtrlP = input.KeyCtrlP
	KeyCtrlQ = input.KeyCtrlQ
	KeyCtrlR = input.KeyCtrlR
	KeyCtrlS = input.KeyCtrlS
	KeyCtrlT = input.KeyCtrlT
	KeyCtrlU = input.KeyCtrlU
	KeyCtrlV = input.KeyCtrlV
	KeyCtrlW = input.KeyCtrlW
	KeyCtrlX = input.KeyCtrlX
	KeyCtrlY = input.KeyCtrlY
	KeyCtrlZ = input.KeyCtrlZ
)

// Клавиши редактирования и навигации.
const (
	KeyEnter        = input.KeyEnter
	KeySpace        = input.KeySpace
	KeyPgUp         = input.KeyPgUp
	KeyPgDown       = input.KeyPgDown
	KeySlash        = input.KeySlash
	KeyReverseSlash = input.KeyReverseSlash
	KeyTab          = input.KeyTab
	KeyShiftTab     = input.KeyShiftTab
	KeyEsc          = input.KeyEsc

	KeyBackspace = input.KeyBackspace
	KeyDelete    = input.KeyDelete
	KeyInsert    = input.KeyInsert
	KeyHome      = input.KeyHome
	KeyEnd       = input.KeyEnd

	KeyArrowUp    = input.KeyArrowUp
	KeyArrowRight = input.KeyArrowRight
	KeyArrowDown  = input.KeyArrowDown
	KeyArrowLeft  = input.KeyArrowLeft
)

// Действия мыши.
const (
	MousePress     = input.MousePress
	MouseRelease   = input.MouseRelease
	MouseMove      = input.MouseMove
	MouseWheelUp   = input.MouseWheelUp
	MouseWheelDown = input.MouseWheelDown

	// NoButton — отсутствие нажатой кнопки мыши.
	NoButton = input.NoButton
)

// MouseWheelLeft и MouseWheelRight поддерживаются не всеми
// терминалами (SGR buttons 66/67). Требуют tilt-колесо мыши
// или горизонтальный жест трекпада.
const (
	MouseWheelLeft  = input.MouseWheelLeft
	MouseWheelRight = input.MouseWheelRight
)
