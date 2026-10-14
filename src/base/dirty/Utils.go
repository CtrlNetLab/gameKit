package dirty

import (
	"regexp"
	"strings"
	"unicode"
)

func toStringArray(content string) []string {
	content = strings.ReplaceAll(content, " ", "")
	result := make([]string, 0)
	var runes = []rune(content)
	for _, r := range runes {
		result = append(result, string(r))
	}
	return result
}

func IsWordLang(lang string) bool {
	// lang = strings.ToLower(lang)
	switch lang {
	case "zh":
		return false
	case "tc":
		return false
	case "en":
		return true
	case "jp":
		return false
	}
	return true
}

func TransLang(lang string) string {
	lang = strings.ToLower(lang)
	if strings.HasPrefix(lang, "zh") {
		if lang == "zh-cn" {
			return "zh"
		}
		return "tc"
	}

	if strings.HasPrefix(lang, "en") {
		return "en"
	}

	if strings.HasPrefix(lang, "ja") {
		return "jp"
	}
	if strings.HasPrefix(lang, "cn") {
		return "zh"
	}

	return "en"
}

func SplitSpaceAdd(content string, isWord bool) []string {
	if !isWord {
		return toStringArray(content)
	}

	content = strings.ToLower(content)
	list := strings.Split(content, " ")
	var result []string
	for _, item := range list {
		item = strings.Trim(item, " ")
		if len(item) > 0 {
			result = append(result, item)
		}
	}
	return result
}

func SplitSpace(content string, isWord bool) []string {
	if !isWord {
		return toStringArray(content)
	}
	return strings.Split(content, " ")
}

func StringJoin(list []string, low, up int, isWord bool) string {
	var sep string = " "
	if !isWord {
		sep = ""
	}

	var result string = ""
	for i := low; i < up; i++ {
		if i == low {
			result = list[i]
		} else {
			result = result + sep + list[i]
		}
	}
	return result
}

func extractWords(s string) []string {
	// 去除空格
	s = strings.Replace(s, " ", "", -1)
	// 正则表达式匹配一个或多个英文单词
	regex := regexp.MustCompile(`[a-z]+`)
	// FindAllStringSubmatch返回所有匹配的子字符串切片
	matches := regex.FindAllString(s, -1)
	// 将[][]string转换为[]string
	words := make([]string, len(matches))
	for i, m := range matches {
		words[i] = m
	}

	splitWords := SplitContentWords(s)
	words = append(words, splitWords...)

	return words
}

// 根据内容拆分词组
func SplitContentWords(content string) []string {
	content = strings.ToLower(content)
	list := strings.Split(content, "")

	var result []string
	for i := 0; i < len(list); i++ {
		flag := false
		cnWord := ""
		for j := i; j < len(list); j++ {
			for _, r := range list[j] {
				if unicode.Is(unicode.Han, r) {
					//中文词拼接
					cnWord = cnWord + list[j]
					result = append(result, cnWord)
				} else {
					flag = true
				}

				break
			}
			if flag {
				break
			}
		}
	}

	return result
}

// 创建正则表达式模式，用于匹配含有随机空格的敏感词
func createSensitivityPattern(word string) string {
	var sb strings.Builder
	for i, char := range word {
		if i > 0 {
			sb.WriteString(`[ \s]*`) // 在每个字符之间插入可选空格
		}
		sb.WriteRune(char)
	}
	return sb.String()
}
