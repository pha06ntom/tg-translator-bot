package translate

type Direction string
type Mode string

const (
	RU_EN Direction = "ru_en"
	EN_RU Direction = "en_ru"
)

const (
	ModeDefault Mode = "default"
	ModeDrawing Mode = "drawing"
)

func (d Direction) FromTo() (from, to string) {
	switch d {
	case EN_RU:
		return "English", "Russian"
	default:
		return "Russian", "English"
	}
}
