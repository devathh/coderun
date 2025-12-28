package xcutrcontainer

type lang int

const (
	GO lang = iota
	PYTHON
)

type Lang struct {
	lang lang
}

func NewLang(id lang) Lang {
	return Lang{
		lang: id,
	}
}

func (l Lang) Value() int {
	return int(l.lang)
}

func (l Lang) String() string {
	switch l.lang {
	case GO:
		return "go"
	case PYTHON:
		return "python"
	}

	return ""
}
