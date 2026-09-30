package acell

import "github.com/romanSPB15/acell/builder"

const (
	Fg16Black   = "30"
	Fg16Red     = "31"
	Fg16Green   = "32"
	Fg16Yellow  = "33"
	Fg16Blue    = "34"
	Fg16Magenta = "35"
	Fg16Cyan    = "36"
	Fg16White   = "37"

	Fg16Grey         = "90"
	Fg16LightRed     = "91"
	Fg16LightGreen   = "92"
	Fg16LightYellow  = "93"
	Fg16LightBlue    = "94"
	Fg16LightMagenta = "95"
	Fg16LightCyan    = "96"
	Fg16LightWhite   = "97"

	Fg16Default = "39"

	Fg16BrightBlack = Fg16Grey
	Fg16BrightWhite = Fg16LightWhite
	Fg16BrightRed   = Fg16LightRed
	Fg16BrightGreen = Fg16LightGreen
)

const (
	Bg16Black   = "40"
	Bg16Red     = "41"
	Bg16Green   = "42"
	Bg16Yellow  = "43"
	Bg16Blue    = "44"
	Bg16Magenta = "45"
	Bg16Cyan    = "46"
	Bg16White   = "47"

	Bg16Grey         = "100"
	Bg16LightRed     = "101"
	Bg16LightGreen   = "102"
	Bg16LightYellow  = "103"
	Bg16LightBlue    = "104"
	Bg16LightMagenta = "105"
	Bg16LightCyan    = "106"
	Bg16LightWhite   = "107"

	Bg16Default = "49"

	Bg16BrightBlack = Bg16Grey
	Bg16BrightWhite = Bg16LightWhite
	Bg16BrightRed   = Bg16LightRed
	Bg16BrightGreen = Bg16LightGreen
)

const (
	FgGrey1  = "38;5;232" // rgb(8,8,8)
	FgGrey2  = "38;5;233" // rgb(18,18,18)
	FgGrey3  = "38;5;234" // rgb(28,28,28)
	FgGrey4  = "38;5;235" // rgb(38,38,38)
	FgGrey5  = "38;5;236" // rgb(48,48,48)
	FgGrey6  = "38;5;237" // rgb(58,58,58)
	FgGrey7  = "38;5;238" // rgb(68,68,68)
	FgGrey8  = "38;5;239" // rgb(78,78,78)
	FgGrey9  = "38;5;240" // rgb(88,88,88)
	FgGrey10 = "38;5;241" // rgb(98,98,98)
	FgGrey11 = "38;5;242" // rgb(108,108,108)
	FgGrey12 = "38;5;243" // rgb(118,118,118)
	FgGrey13 = "38;5;244" // rgb(128,128,128)
	FgGrey14 = "38;5;245" // rgb(138,138,138)
	FgGrey15 = "38;5;246" // rgb(148,148,148)
	FgGrey16 = "38;5;247" // rgb(158,158,158)
	FgGrey17 = "38;5;248" // rgb(168,168,168)
	FgGrey18 = "38;5;249" // rgb(178,178,178)
	FgGrey19 = "38;5;250" // rgb(188,188,188)
	FgGrey20 = "38;5;251" // rgb(198,198,198)
	FgGrey21 = "38;5;252" // rgb(208,208,208)
	FgGrey22 = "38;5;253" // rgb(218,218,218)
	FgGrey23 = "38;5;254" // rgb(228,228,228)
	FgGrey24 = "38;5;255" // rgb(238,238,238)
)

const (
	BgGrey1  = "48;5;232" // rgb(8,8,8)
	BgGrey2  = "48;5;233" // rgb(18,18,18)
	BgGrey3  = "48;5;234" // rgb(28,28,28)
	BgGrey4  = "48;5;235" // rgb(38,38,38)
	BgGrey5  = "48;5;236" // rgb(48,48,48)
	BgGrey6  = "48;5;237" // rgb(58,58,58)
	BgGrey7  = "48;5;238" // rgb(68,68,68)
	BgGrey8  = "48;5;239" // rgb(78,78,78)
	BgGrey9  = "48;5;240" // rgb(88,88,88)
	BgGrey10 = "48;5;241" // rgb(98,98,98)
	BgGrey11 = "48;5;242" // rgb(108,108,108)
	BgGrey12 = "48;5;243" // rgb(118,118,118)
	BgGrey13 = "48;5;244" // rgb(128,128,128)
	BgGrey14 = "48;5;245" // rgb(138,138,138)
	BgGrey15 = "48;5;246" // rgb(148,148,148)
	BgGrey16 = "48;5;247" // rgb(158,158,158)
	BgGrey17 = "48;5;248" // rgb(168,168,168)
	BgGrey18 = "48;5;249" // rgb(178,178,178)
	BgGrey19 = "48;5;250" // rgb(188,188,188)
	BgGrey20 = "48;5;251" // rgb(198,198,198)
	BgGrey21 = "48;5;252" // rgb(208,208,208)
	BgGrey22 = "48;5;253" // rgb(218,218,218)
	BgGrey23 = "48;5;254" // rgb(228,228,228)
	BgGrey24 = "48;5;255" // rgb(238,238,238)
)

const (
	FgBlack = "38;5;16"  // rgb(0,0,0)
	FgWhite = "38;5;231" // rgb(255,255,255)

	FgNavy        = "38;5;17"  // rgb(0,0,95)
	FgDarkBlue    = "38;5;18"  // rgb(0,0,135)
	FgMediumBlue  = "38;5;19"  // rgb(0,0,175)
	FgBlue        = "38;5;21"  // rgb(0,0,255)
	FgDodgerBlue  = "38;5;27"  // rgb(0,95,255)
	FgAzure       = "38;5;33"  // rgb(0,135,255)
	FgDeepSkyBlue = "38;5;39"  // rgb(0,175,255)
	FgSkyBlue     = "38;5;117" // rgb(135,215,255)
	FgRoyalBlue   = "38;5;63"  // rgb(95,95,255)
	FgCornflower  = "38;5;69"  // rgb(95,135,255)
	FgSlateBlue   = "38;5;99"  // rgb(135,95,255)
	FgLightBlue   = "38;5;153" // rgb(175,215,255)
	FgLightSteel  = "38;5;152" // rgb(175,215,215)

	FgTeal       = "38;5;30" // rgb(0,135,135)
	FgDarkCyan   = FgTeal
	FgCyan       = "38;5;51" // rgb(0,255,255)
	FgAqua       = FgCyan
	FgLightSea   = "38;5;37" // rgb(0,175,175)
	FgTurquoise  = "38;5;44" // rgb(0,215,215)
	FgSteelBlue  = "38;5;67" // rgb(95,135,175)
	FgCadetBlue  = "38;5;66" // rgb(95,135,135)
	FgPowderBlue = FgLightSteel
	FgPaleCyan   = "38;5;195" // rgb(215,255,255)

	FgDarkGreen    = "38;5;22" // rgb(0,95,0)
	FgForestGreen  = "38;5;28" // rgb(0,135,0)
	FgGreen        = "38;5;46" // rgb(0,255,0)
	FgLime         = FgGreen
	FgSpringGreen  = "38;5;47"  // rgb(0,255,95)
	FgMediumSpring = "38;5;48"  // rgb(0,255,135)
	FgGreenYellow  = "38;5;118" // rgb(135,255,0)
	FgChartreuse   = FgGreenYellow
	FgLightGreen   = "38;5;120" // rgb(135,255,135)
	FgPaleGreen    = "38;5;121" // rgb(135,255,175)
	FgSeaGreen     = "38;5;29"  // rgb(0,135,95)
	FgMediumSea    = "38;5;35"  // rgb(0,175,95)
	FgLimeGreen    = "38;5;40"  // rgb(0,215,0)

	FgOliveDrab = "38;5;100" // rgb(135,135,0)
	FgOlive     = "38;5;100" // rgb(135,135,0)

	FgDarkKhaki     = "38;5;143" // rgb(175,175,95)
	FgKhaki         = "38;5;222" // rgb(255,215,135)
	FgYellow        = "38;5;226" // rgb(255,255,0)
	FgPaleYellow    = "38;5;229" // rgb(255,255,175)
	FgLightYellow   = "38;5;230" // rgb(255,255,215)
	FgDarkGoldenrod = "38;5;136" // rgb(175,135,0)
	FgGoldenrod     = "38;5;178" // rgb(215,175,0)

	FgOrangeRed  = "38;5;202" // rgb(255,95,0)
	FgDarkOrange = "38;5;208" // rgb(255,135,0)
	FgOrange     = "38;5;214" // rgb(255,175,0)
	FgGold       = "38;5;220" // rgb(255,215,0)
	FgChocolate  = "38;5;166" // rgb(215,95,0)
	FgSaddle     = "38;5;130" // rgb(175,95,0)
	FgPeru       = "38;5;173" // rgb(215,135,95)
	FgTan        = "38;5;179" // rgb(215,175,95)
	FgSandy      = "38;5;215" // rgb(255,175,95)
	FgPeach      = "38;5;223" // rgb(255,215,175)

	FgDarkRed   = "38;5;88"  // rgb(135,0,0)
	FgFirebrick = "38;5;124" // rgb(175,0,0)
	FgRed       = "38;5;196" // rgb(255,0,0)
	FgCrimson   = "38;5;161" // rgb(215,0,95)
	FgTomato    = "38;5;203" // rgb(255,95,95)
	FgCoral     = "38;5;209" // rgb(255,135,95)
	FgSalmon    = "38;5;210" // rgb(255,135,135)
	FgDeepPink  = "38;5;198" // rgb(255,0,135)
	FgHotPink   = "38;5;205" // rgb(255,95,175)
	FgPink      = "38;5;212" // rgb(255,135,215)
	FgLightPink = "38;5;217" // rgb(255,175,175)

	FgIndigo       = "38;5;54" // rgb(95,0,135)
	FgBlueViolet   = "38;5;56" // rgb(95,0,215)
	FgPurple       = "38;5;90" // rgb(135,0,135)
	FgDarkViolet   = "38;5;91" // rgb(135,0,175)
	FgDarkMagenta  = FgPurple
	FgDarkOrchid   = "38;5;128" // rgb(175,0,215)
	FgMediumOrchid = "38;5;134" // rgb(175,95,215)
	FgMediumPurple = "38;5;141" // rgb(175,135,255)
	FgViolet       = "38;5;177" // rgb(215,135,255)
	FgPlum         = "38;5;183" // rgb(215,175,255)
	FgThistle      = "38;5;225" // rgb(255,215,255)
	FgOrchid       = "38;5;213" // rgb(255,135,255)
	FgMagenta      = "38;5;201" // rgb(255,0,255)
	FgFuchsia      = FgMagenta
)

const (
	BgBlack = "48;5;16"  // rgb(0,0,0)
	BgWhite = "48;5;231" // rgb(255,255,255)

	BgNavy        = "48;5;17"  // rgb(0,0,95)
	BgDarkBlue    = "48;5;18"  // rgb(0,0,135)
	BgMediumBlue  = "48;5;19"  // rgb(0,0,175)
	BgBlue        = "48;5;21"  // rgb(0,0,255)
	BgDodgerBlue  = "48;5;27"  // rgb(0,95,255)
	BgAzure       = "48;5;33"  // rgb(0,135,255)
	BgDeepSkyBlue = "48;5;39"  // rgb(0,175,255)
	BgSkyBlue     = "48;5;117" // rgb(135,215,255)
	BgRoyalBlue   = "48;5;63"  // rgb(95,95,255)
	BgCornflower  = "48;5;69"  // rgb(95,135,255)
	BgSlateBlue   = "48;5;99"  // rgb(135,95,255)
	BgLightBlue   = "48;5;153" // rgb(175,215,255)
	BgLightSteel  = "48;5;152" // rgb(175,215,215)

	BgTeal       = "48;5;30" // rgb(0,135,135)
	BgDarkCyan   = BgTeal
	BgCyan       = "48;5;51" // rgb(0,255,255)
	BgAqua       = BgCyan
	BgLightSea   = "48;5;37" // rgb(0,175,175)
	BgTurquoise  = "48;5;44" // rgb(0,215,215)
	BgSteelBlue  = "48;5;67" // rgb(95,135,175)
	BgCadetBlue  = "48;5;66" // rgb(95,135,135)
	BgPowderBlue = BgLightSteel
	BgPaleCyan   = "48;5;195" // rgb(215,255,255)

	BgDarkGreen    = "48;5;22" // rgb(0,95,0)
	BgForestGreen  = "48;5;28" // rgb(0,135,0)
	BgGreen        = "48;5;46" // rgb(0,255,0)
	BgLime         = BgGreen
	BgSpringGreen  = "48;5;47"  // rgb(0,255,95)
	BgMediumSpring = "48;5;48"  // rgb(0,255,135)
	BgGreenYellow  = "48;5;118" // rgb(135,255,0)
	BgChartreuse   = BgGreenYellow
	BgLightGreen   = "48;5;120" // rgb(135,255,135)
	BgPaleGreen    = "48;5;121" // rgb(135,255,175)
	BgSeaGreen     = "48;5;29"  // rgb(0,135,95)
	BgMediumSea    = "48;5;35"  // rgb(0,175,95)
	BgLimeGreen    = "48;5;40"  // rgb(0,215,0)

	BgOliveDrab = "48;5;100" // rgb(135,135,0)
	BgOlive     = "48;5;100" // rgb(135,135,0)

	BgDarkKhaki     = "48;5;143" // rgb(175,175,95)
	BgKhaki         = "48;5;222" // rgb(255,215,135)
	BgYellow        = "48;5;226" // rgb(255,255,0)
	BgPaleYellow    = "48;5;229" // rgb(255,255,175)
	BgLightYellow   = "48;5;230" // rgb(255,255,215)
	BgDarkGoldenrod = "48;5;136" // rgb(175,135,0)
	BgGoldenrod     = "48;5;178" // rgb(215,175,0)

	BgOrangeRed  = "48;5;202" // rgb(255,95,0)
	BgDarkOrange = "48;5;208" // rgb(255,135,0)
	BgOrange     = "48;5;214" // rgb(255,175,0)
	BgGold       = "48;5;220" // rgb(255,215,0)
	BgChocolate  = "48;5;166" // rgb(215,95,0)
	BgSaddle     = "48;5;130" // rgb(175,95,0)
	BgPeru       = "48;5;173" // rgb(215,135,95)
	BgTan        = "48;5;179" // rgb(215,175,95)
	BgSandy      = "48;5;215" // rgb(255,175,95)
	BgPeach      = "48;5;223" // rgb(255,215,175)

	BgDarkRed   = "48;5;88"  // rgb(135,0,0)
	BgFirebrick = "48;5;124" // rgb(175,0,0)
	BgRed       = "48;5;196" // rgb(255,0,0)
	BgCrimson   = "48;5;161" // rgb(215,0,95)
	BgTomato    = "48;5;203" // rgb(255,95,95)
	BgCoral     = "48;5;209" // rgb(255,135,95)
	BgSalmon    = "48;5;210" // rgb(255,135,135)
	BgDeepPink  = "48;5;198" // rgb(255,0,135)
	BgHotPink   = "48;5;205" // rgb(255,95,175)
	BgPink      = "48;5;212" // rgb(255,135,215)
	BgLightPink = "48;5;217" // rgb(255,175,175)

	BgIndigo       = "48;5;54" // rgb(95,0,135)
	BgBlueViolet   = "48;5;56" // rgb(95,0,215)
	BgPurple       = "48;5;90" // rgb(135,0,135)
	BgDarkViolet   = "48;5;91" // rgb(135,0,175)
	BgDarkMagenta  = BgPurple
	BgDarkOrchid   = "48;5;128" // rgb(175,0,215)
	BgMediumOrchid = "48;5;134" // rgb(175,95,215)
	BgMediumPurple = "48;5;141" // rgb(175,135,255)
	BgViolet       = "48;5;177" // rgb(215,135,255)
	BgPlum         = "48;5;183" // rgb(215,175,255)
	BgThistle      = "48;5;225" // rgb(255,215,255)
	BgOrchid       = "48;5;213" // rgb(255,135,255)
	BgMagenta      = "48;5;201" // rgb(255,0,255)
	BgFuchsia      = BgMagenta
)

// writeSGR256 пишет SGR-последовательность для 256-цветной палитры.
// fg == true — foreground, false — background.
func writeSGR256(bb *builder.Builder, n int, fg bool) {
	switch {
	case n < 0:
		n = 0
	case n > 255:
		n = 255
	case n < 8:
		base := 30
		if !fg {
			base = 40
		}
		bb.WriteInt(base + n)
		return
	case n < 16:
		base := 90
		if !fg {
			base = 100
		}
		bb.WriteInt(base + n - 8)
		return
	}
	if fg {
		bb.WriteString("38;5;")
	} else {
		bb.WriteString("48;5;")
	}
	bb.WriteInt(n)
}

// writeRGB пишет SGR-последовательность для True Color.
// fg == true — foreground, false — background.
func writeRGB(bb *builder.Builder, r, g, b int, fg bool) {
	if fg {
		bb.WriteString("38;2;")
	} else {
		bb.WriteString("48;2;")
	}
	bb.WriteInt(r)
	bb.WriteByte(';')
	bb.WriteInt(g)
	bb.WriteByte(';')
	bb.WriteInt(b)
}

// Fg256 возвращает foreground-строку цвета для 256-цветной палитры.
// Значения вне 0..255 приводятся к ближайшей границе.
func Fg256(n int) string {
	var bb builder.Builder
	bb.Grow(16)
	writeSGR256(&bb, n, true)
	return bb.String()
}

// Bg256 возвращает background-строку цвета для 256-цветной палитры.
// Значения вне 0..255 приводятся к ближайшей границе.
func Bg256(n int) string {
	var bb builder.Builder
	bb.Grow(16)
	writeSGR256(&bb, n, false)
	return bb.String()
}

// FgRGB возвращает foreground-строку цвета для True Color.
func FgRGB(r, g, b int) string {
	var bb builder.Builder
	bb.Grow(16)
	writeRGB(&bb, r, g, b, true)
	return bb.String()
}

// BgRGB возвращает background-строку цвета для True Color.
func BgRGB(r, g, b int) string {
	var bb builder.Builder
	bb.Grow(16)
	writeRGB(&bb, r, g, b, false)
	return bb.String()
}
