package dirty

import (
	"gamekit/pkg/configs"
	"strings"
)

var (
	FilterMgr *filterMgr = &filterMgr{}
	languages []string   = []string{"zh", "tc", "en", "jp"}
)

type filterMgr struct {
	// map[lang] -> *Filter
	filters []map[string]*Filter
	index   int
}

func (this *filterMgr) Init(start bool) {
	// 加载脏字库
	if start {
		FilterMgr = &filterMgr{
			filters: make([]map[string]*Filter, 2),
			index:   1,
		}
		for i := 0; i < 2; i++ {
			FilterMgr.filters[i] = make(map[string]*Filter)
		}
	}

	filterMap := make(map[string]*Filter)
	dictRes := configs.WordsConfigResMgr.GetAllData()
	for _, item := range dictRes.Items {
		if item == nil {
			continue
		}

		var (
			filter *Filter
			ok     bool
		)

		load := func(data, lang string) {
			data = strings.Trim(data, " ")
			isWord := IsWordLang(lang)
			if filter, ok = filterMap[lang]; !ok || filter == nil {
				filter = NewFilter(lang, isWord)
				filterMap[lang] = filter
			}
			if filter != nil {
				filter.trie.Add(data)
			}
		}

		load(item.En, "en")
	}

	index := 1 - FilterMgr.index
	FilterMgr.filters[index] = filterMap
	FilterMgr.index = index

}

/***
 * @brief 判断单词是否合法
 *
 * @param lang: 玩家当前语言环境，客户端上传
 * @param Sentence: 要检测的单词
 *
 * @retval bool: true 合法; false 不合法
 *         string: 匹配到的敏感词
 *
 */
func Validate(lang, Sentence string) (bool, string) {
	return FilterMgr.validateSentence(lang, Sentence)
}

func (this *filterMgr) validateSentence(lang, sentence string) (bool, string) {
	lang = TransLang(lang)
	filter, ok := this.filters[this.index][lang]
	if !ok || filter == nil {
		return false, "[unkonwn lang]"
	}
	return filter.Validate(sentence)
}

/***
 * @brief 判断单词是否合法并屏蔽不合法词为：*
 *
 * @param lang: 玩家当前语言环境，客户端上传
 * @param sentence: 要替换的句子
 *
 * @retval bool: true 存在不合法; false 合法
 *         string: 替换之后的句子
 *         []string: 匹配中的敏感词列表
 *
 */
func Replace(lang, sentence string) (bool, string, []string) {
	return FilterMgr.filterSentence(lang, sentence)
}

func (this *filterMgr) filterSentence(lang, sentence string) (bool, string, []string) {
	lang = TransLang(lang)
	filter, ok := this.filters[this.index][lang]
	if !ok || filter == nil {
		return true, "******", []string{"unknown lang"}
	}
	return filter.Replace(sentence, "*")
}
