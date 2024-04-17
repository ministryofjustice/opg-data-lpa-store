package shared

type Lang string

var (
	LangNotSet = Lang("")
	LangCy     = Lang("cy")
	LangEn     = Lang("en")
)

func (l Lang) IsValid() bool {
	return l == LangCy || l == LangEn
}
