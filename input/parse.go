// Пакет input предоставляет парсер ввода для терминала:
// клавиатура, мышь, фокус окна. Не зависит от term/raw — может
// использоваться в любой реализации RawTerminal.
package input

// Parse разбирает байты в одно из событий:
// *WindowFocusEvent, *MouseEvent или *KeyboardEvent.
//
// Порядок важен: focus и mouse проверяются раньше клавиатуры,
// потому что их последовательности начинаются с ESC,
// который иначе ушёл бы в клавиатурный парсер.
//
// Возвращает nil, если данные не распознаны.
func Parse(data []byte) any {
	if ev := ParseFocus(data); ev != nil {
		return ev
	}
	if ev := ParseMouse(data); ev != nil {
		return ev
	}
	if ev := ParseKeyboard(data); ev != nil {
		return ev
	}
	return nil
}
