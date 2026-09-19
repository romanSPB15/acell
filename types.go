package acell

import "github.com/romanSPB15/acell/term"

// Реэкспорт типов из term для удобства пользователей acell.
type (
	Key           = term.Key
	KeyboardEvent = term.KeyboardEvent
	MouseEvent    = term.MouseEvent
	MouseAction   = term.MouseAction
	Point         = term.Point
	RawTerminal   = term.RawTerminal
)

// Реэкспорт клавиш.
const (
	KeyUnknown Key = term.KeyUnknown

	KeyF1  = term.KeyF1
	KeyF2  = term.KeyF2
	KeyF3  = term.KeyF3
	KeyF4  = term.KeyF4
	KeyF5  = term.KeyF5
	KeyF6  = term.KeyF6
	KeyF7  = term.KeyF7
	KeyF8  = term.KeyF8
	KeyF9  = term.KeyF9
	KeyF10 = term.KeyF10
	KeyF11 = term.KeyF11
	KeyF12 = term.KeyF12

	KeyCtrlA = term.KeyCtrlA
	KeyCtrlB = term.KeyCtrlB
	KeyCtrlC = term.KeyCtrlC
	KeyCtrlD = term.KeyCtrlD
	KeyCtrlE = term.KeyCtrlE
	KeyCtrlF = term.KeyCtrlF
	KeyCtrlG = term.KeyCtrlG
	KeyCtrlH = term.KeyCtrlH
	KeyCtrlI = term.KeyCtrlI
	KeyCtrlJ = term.KeyCtrlJ
	KeyCtrlK = term.KeyCtrlK
	KeyCtrlL = term.KeyCtrlL
	KeyCtrlM = term.KeyCtrlM
	KeyCtrlN = term.KeyCtrlN
	KeyCtrlO = term.KeyCtrlO
	KeyCtrlP = term.KeyCtrlP
	KeyCtrlQ = term.KeyCtrlQ
	KeyCtrlR = term.KeyCtrlR
	KeyCtrlS = term.KeyCtrlS
	KeyCtrlT = term.KeyCtrlT
	KeyCtrlU = term.KeyCtrlU
	KeyCtrlV = term.KeyCtrlV
	KeyCtrlW = term.KeyCtrlW
	KeyCtrlX = term.KeyCtrlX
	KeyCtrlY = term.KeyCtrlY
	KeyCtrlZ = term.KeyCtrlZ

	KeyEnter        = term.KeyEnter
	KeySpace        = term.KeySpace
	KeyPgUp         = term.KeyPgUp
	KeyPgDown       = term.KeyPgDown
	KeySlash        = term.KeySlash
	KeyReverseSlash = term.KeyReverseSlash
	KeyTab          = term.KeyTab
	KeyShiftTab     = term.KeyShiftTab

	KeyBackspace = term.KeyBackspace
	KeyDelete    = term.KeyDelete
	KeyInsert    = term.KeyInsert
	KeyHome      = term.KeyHome
	KeyEnd       = term.KeyEnd

	KeyArrowUp    = term.KeyArrowUp
	KeyArrowRight = term.KeyArrowRight
	KeyArrowDown  = term.KeyArrowDown
	KeyArrowLeft  = term.KeyArrowLeft

	KeyEsc = term.KeyEsc
)

// Реэкспорт действий мыши и константы "нет кнопки".
const (
	MousePress      = term.MousePress
	MouseRelease    = term.MouseRelease
	MouseMove       = term.MouseMove
	MouseWheelUp    = term.MouseWheelUp
	MouseWheelDown  = term.MouseWheelDown
	MouseWheelLeft  = term.MouseWheelLeft
	MouseWheelRight = term.MouseWheelRight

	NoButton = term.NoButton
)
