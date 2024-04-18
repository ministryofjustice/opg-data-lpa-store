package shared

type Lang string

var (
	LangCy = Lang("cy")
	LangEn = Lang("en")
)

func (l Lang) IsValid() bool {
	return l == LangCy || l == LangEn
}
