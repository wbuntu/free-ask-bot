package utils

type LanguageCode string

const (
	// 英文
	LanguageCodeEn LanguageCode = "en"
	// 简体中文
	LanguageCodeZhHans LanguageCode = "zh-hans"
	// 繁體中文
	LanguageCodeZhHant LanguageCode = "zh-hant"
)

func (lc LanguageCode) Description() string {
	m := map[LanguageCode]string{
		LanguageCodeEn:     "English",
		LanguageCodeZhHans: "简体中文",
		LanguageCodeZhHant: "繁體中文",
	}
	if v, ok := m[lc]; ok {
		return v
	}
	return ""
}
