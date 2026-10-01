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

// ParseOne разбирает одно событие из начала data.
// Возвращает событие и количество съеденных байт.
// Если данных недостаточно (неполная последовательность) —
// возвращает (nil, 0), и вызывающий должен накопить ещё.
func ParseOne(data []byte) (any, int) {
	if len(data) == 0 {
		return nil, 0
	}

	if len(data) >= 3 && data[0] == 27 && data[1] == '[' && (data[2] == 'I' || data[2] == 'O') {
		return ParseFocus(data[:3]), 3
	}

	if len(data) >= 3 && data[0] == 27 && data[1] == '[' && data[2] == '<' {
		for i := 3; i < len(data); i++ {
			if data[i] == 'M' || data[i] == 'm' {
				if ev := ParseMouse(data[:i+1]); ev != nil {
					return ev, i + 1
				}
				return nil, i + 1
			}
		}
		return nil, 0
	}

	if data[0] == 27 {
		if len(data) == 1 {
			return nil, 0
		}
		if data[1] == '[' {
			for i := 2; i < len(data); i++ {
				if data[i] >= 0x40 && data[i] <= 0x7E {
					if ev := ParseKeyboard(data[:i+1]); ev != nil {
						return ev, i + 1
					}
					return nil, i + 1
				}
			}
			return nil, 0
		}
		if data[1] == 'O' {
			if len(data) < 3 {
				return nil, 0
			}
			if ev := ParseKeyboard(data[:3]); ev != nil {
				return ev, 3
			}
			return nil, 3
		}

		if ev := ParseKeyboard(data[:2]); ev != nil {
			return ev, 2
		}
		return nil, 2
	}

	if data[0] < 0x80 {
		if ev := ParseKeyboard(data[:1]); ev != nil {
			return ev, 1
		}
		return nil, 1
	}

	var size int
	switch {
	case data[0]&0xE0 == 0xC0:
		size = 2
	case data[0]&0xF0 == 0xE0:
		size = 3
	case data[0]&0xF8 == 0xF0:
		size = 4
	default:
		return nil, 1 // битый байт
	}
	if len(data) < size {
		return nil, 0 // ждём
	}
	if ev := ParseKeyboard(data[:size]); ev != nil {
		return ev, size
	}
	return nil, size
}
