package dirty

import (
	// "fmt"

	"fmt"
	"regexp"
	"strings"
)

// Trie 短语组成的Trie树.
type Trie struct {
	lang   string
	isWord bool
	Root   *Node
}

// Node Trie树上的一个节点.
type Node struct {
	isRootNode bool
	isPathEnd  bool
	// Character  rune
	// Children   map[rune]*Node
	Character string
	Children  map[string]*Node
}

// NewTrie 新建一棵Trie
func NewTrie(lang string, isWord bool) *Trie {
	return &Trie{
		lang:   lang,
		isWord: isWord,
		// Root: NewRootNode(0),
		Root: NewRootNode(""),
	}
}

// Add 添加若干个词
func (tree *Trie) Add(words ...string) {
	for _, word := range words {
		tree.add(word)
	}
}

func (tree *Trie) add(word string) {
	var current = tree.Root
	// var runes = []rune(word)
	var runes = SplitSpaceAdd(word, tree.isWord)
	for position := 0; position < len(runes); position++ {
		r := runes[position]
		if next, ok := current.Children[r]; ok {
			current = next
		} else {
			newNode := NewNode(r)
			current.Children[r] = newNode
			current = newNode
		}
		if position == len(runes)-1 {
			current.isPathEnd = true
		}
	}
}

func (tree *Trie) Del(words ...string) {
	for _, word := range words {
		tree.del(word)
	}
}

func (tree *Trie) del(word string) {
	var current = tree.Root
	// var runes = []rune(word)
	// var runes = SplitSpace(word, tree.isWord)
	var runes = SplitSpaceAdd(word, tree.isWord)
	for position := 0; position < len(runes); position++ {
		r := runes[position]
		if next, ok := current.Children[r]; !ok {
			return
		} else {
			current = next
		}

		if position == len(runes)-1 {
			current.SoftDel()
		}
	}
}

func (tree *Trie) ReplaceAgain(text string, character string, status bool, matchs []string) (bool, string, []string) {
	var (
		parent = tree.Root
	)

	runes := extractWords(text)

	for i := 0; i < len(runes); i++ {
		_, found := parent.Children[runes[i]]
		if found {
			status = true
			r := []rune(runes[i])
			iLen := len(r)
			replaceSymbol := ""
			for j := 0; j < iLen; j++ {
				replaceSymbol = replaceSymbol + character
			}
			text = strings.ReplaceAll(text, runes[i], replaceSymbol)

			// 使用正则表达式，替换含空格字符
			pattern := createSensitivityPattern(runes[i])
			re := regexp.MustCompile("(?i)" + pattern)
			text = re.ReplaceAllString(text, replaceSymbol)

			matchs = append(matchs, runes[i])
		}
	}
	return status, text, matchs
}

// Replace 词语替换
// func (tree *Trie) Replace(text string, character rune) string {
func (tree *Trie) Replace(text string, character string) (bool, string, []string) {
	text = strings.ReplaceAll(text, "\n", " ")

	var (
		status  bool
		parent  = tree.Root
		current *Node
		// runes   = []rune(text)
		// runes  = SplitSpace(text, tree.isWord)
		runes  = SplitSpaceAdd(text, tree.isWord)
		length = len(runes)
		left   = 0
		found  bool
		matchs []string = make([]string, 0)
		sep    string   = " "
	)

	if !tree.isWord {
		sep = ""
	}

	for position := 0; position < len(runes); position++ {
		current, found = parent.Children[runes[position]]

		if !found || (!current.IsPathEnd() && position == length-1) {
			parent = tree.Root
			position = left
			left++
			continue
		}

		// println(string(current.Character), current.IsPathEnd(), left)
		if current.IsPathEnd() && left <= position {
			status = true
			var match string
			for i := left; i <= position; i++ {
				r := []rune(runes[i])
				iLen := len(r)
				runes[i] = strings.Repeat(character, iLen)
				if len(match) == 0 {
					match = fmt.Sprintf("%s", string(r))
				} else {
					match = fmt.Sprintf("%s%s%s", match, sep, string(r))
				}
			}
			matchs = append(matchs, match)
		}

		parent = current
	}

	filterText := strings.Join(runes, sep)
	//二次过滤
	status, filterText, matchs = tree.ReplaceAgain(filterText, character, status, matchs)

	//logger.Log.Debug("nicolelog Replace filterText:%d, matchs:%d", filterText, matchs)

	return status, filterText, matchs
}

// Filter 直接过滤掉字符串中的敏感词
func (tree *Trie) Filter(text string) string {
	var (
		parent  = tree.Root
		current *Node
		left    = 0
		found   bool
		// runes       = []rune(text)
		// runes  = SplitSpace(text, tree.isWord)
		runes  = SplitSpaceAdd(text, tree.isWord)
		length = len(runes)
		// resultRunes = make([]rune, 0, length)
		resultRunes = make([]string, 0, length)
	)

	for position := 0; position < length; position++ {
		current, found = parent.Children[runes[position]]

		if !found || (!current.IsPathEnd() && position == length-1) {
			resultRunes = append(resultRunes, runes[left])
			parent = tree.Root
			position = left
			left++
			continue
		}

		if current.IsPathEnd() {
			// fmt.Printf("match left: %d, position: %d, str: %s\n", left, position, StringJoin(runes, left, position+1))
			for i := left; i < position+1; i++ {
				// resultRunes = append(resultRunes, "*")
				r := []rune(runes[i])
				iLen := len(r)
				star := strings.Repeat("*", iLen)
				resultRunes = append(resultRunes, star)
				if i != left && i != position {
					resultRunes = append(resultRunes, " ")
				}
			}
			left = position + 1
			parent = tree.Root
		} else {
			parent = current
		}

	}

	resultRunes = append(resultRunes, runes[left:]...)
	// return string(resultRunes)
	return strings.Join(resultRunes, " ")
}

// Validate 验证字符串是否合法，如不合法则返回false和检测到
// 的第一个敏感词
func (tree *Trie) Validate(text string) (bool, string) {
	// fmt.Printf("text: %s\n", text)
	const (
		Empty = ""
	)
	var (
		parent  = tree.Root
		current *Node
		// runes   = []rune(text)
		// runes  = SplitSpace(text, tree.isWord)
		runes  = SplitSpaceAdd(text, tree.isWord)
		length = len(runes)
		left   = 0
		found  bool
	)

	for position := 0; position < len(runes); position++ {
		current, found = parent.Children[runes[position]]
		// fmt.Printf("current: %v, found: %v\n", current, found)
		if !found || (!current.IsPathEnd() && position == length-1) {
			parent = tree.Root
			position = left
			left++
			continue
		}

		if current.IsPathEnd() && left <= position {
			// return false, string(runes[left : position+1])
			return false, StringJoin(runes, left, position+1, tree.isWord)
		}

		parent = current
	}

	//二次校验
	runes = extractWords(text)
	for i := 0; i < len(runes); i++ {
		_, found := parent.Children[runes[i]]
		if found {
			return false, runes[i]
		}
	}

	return true, Empty
}

// FindIn 判断text中是否含有词库中的词
func (tree *Trie) FindIn(text string) (bool, string) {
	validated, first := tree.Validate(text)
	return !validated, first
}

// FindAll 找有所有包含在词库中的词
func (tree *Trie) FindAll(text string) []string {
	var matches []string
	var (
		parent  = tree.Root
		current *Node
		// runes   = []rune(text)
		// runes  = SplitSpace(text, tree.isWord)
		runes  = SplitSpaceAdd(text, tree.isWord)
		length = len(runes)
		left   = 0
		found  bool
	)

	for position := 0; position < length; position++ {
		current, found = parent.Children[runes[position]]

		if !found {
			parent = tree.Root
			position = left
			left++
			continue
		}

		if current.IsPathEnd() && left <= position {
			// matches = append(matches, string(runes[left:position+1]))
			matches = append(matches, StringJoin(runes, left, position+1, tree.isWord))
		}

		if position == length-1 {
			parent = tree.Root
			position = left
			left++
			continue
		}

		parent = current
	}

	var i = 0
	if count := len(matches); count > 0 {
		set := make(map[string]struct{})
		for i < count {
			_, ok := set[matches[i]]
			if !ok {
				set[matches[i]] = struct{}{}
				i++
				continue
			}
			count--
			copy(matches[i:], matches[i+1:])
		}
		return matches[:count]
	}

	return nil
}

// NewNode 新建子节点
// func NewNode(character rune) *Node {
func NewNode(character string) *Node {
	return &Node{
		Character: character,
		Children:  make(map[string]*Node, 0),
	}
}

// NewRootNode 新建根节点
// func NewRootNode(character rune) *Node {
func NewRootNode(character string) *Node {
	return &Node{
		isRootNode: true,
		Character:  character,
		Children:   make(map[string]*Node, 0),
	}
}

// IsLeafNode 判断是否叶子节点
func (node *Node) IsLeafNode() bool {
	return len(node.Children) == 0
}

// IsRootNode 判断是否为根节点
func (node *Node) IsRootNode() bool {
	return node.isRootNode
}

// IsPathEnd 判断是否为某个路径的结束
func (node *Node) IsPathEnd() bool {
	return node.isPathEnd
}

// SoftDel 置软删除状态
func (node *Node) SoftDel() {
	node.isPathEnd = false
}
