package dictmatch_tries

import (
	"fmt"
	"sort"
	"sync"
)

type KeyWordKV map[int64]string
type CharBeginKV map[string][]*KeyWordTreeNode
type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].V > p[j].V }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Pair struct {
	K int64
	V int64
}

//字典树的每个节点
type KeyWordTreeNode struct {
	// 1, 2, 3,这是啥？没理解
	KeyWordIds map[int64]bool
	// 百，度， 一， 下
	//节点的值
	Char string
	// 父节点
	ParentKeyWordTreeNode *KeyWordTreeNode
	// 子节点集合
	SubKeyWordTreeNodes map[string]*KeyWordTreeNode
}

//创建一个字典树的头节点
func NewKeyWordTreeNode() *KeyWordTreeNode {
	return &KeyWordTreeNode{
		KeyWordIds:            make(map[int64]bool, 0),
		Char:                  "",
		ParentKeyWordTreeNode: nil,
		SubKeyWordTreeNodes:   make(map[string]*KeyWordTreeNode, 0),
	}
}

func NewKeyWordTreeNodeWithParams(ch string, parent *KeyWordTreeNode) *KeyWordTreeNode {
	return &KeyWordTreeNode{
		KeyWordIds:            make(map[int64]bool, 0),
		Char:                  ch,
		ParentKeyWordTreeNode: parent,
		SubKeyWordTreeNodes:   make(map[string]*KeyWordTreeNode, 0),
	}
}

type KeyWordServer struct {
	root          *KeyWordTreeNode
	kv            KeyWordKV
	char_begin_kv CharBeginKV
	rw            sync.RWMutex
}

func NewKeyWordServer() *KeyWordServer {
	return &KeyWordServer{
		root:          NewKeyWordTreeNode(),
		kv:            KeyWordKV{},
		char_begin_kv: CharBeginKV{},
	}
}

func (s *KeyWordServer) Put(id int64, keyword string) {
	s.rw.Lock()
	s.kv[id] = keyword
	key_word_tmp_pt := s.root
	for _, v := range keyword {
		ch := string(v)
		if key_word_tmp_pt.SubKeyWordTreeNodes[ch] == nil {
			// root：{百:{}}
			node := NewKeyWordTreeNodeWithParams(ch, key_word_tmp_pt)
			key_word_tmp_pt.SubKeyWordTreeNodes[ch] = node
			// 记住不同位置的"百"的起点
			s.char_begin_kv[ch] = append(s.char_begin_kv[ch], node)

		}
		// {百:{度:{}}}
		key_word_tree_node := key_word_tmp_pt.SubKeyWordTreeNodes[ch]
		key_word_tree_node.KeyWordIds[id] = true
		// 更新指针
		key_word_tmp_pt = key_word_tmp_pt.SubKeyWordTreeNodes[ch]
	}
	s.rw.Unlock()
}

func (s *KeyWordServer) Sugg(keyword string, limit int) []string {
	s.rw.RLock()
	key_word_tmp_pt := s.root
	is_end := true
	for _, v := range keyword {
		ch := string(v)
		if key_word_tmp_pt.SubKeyWordTreeNodes[ch] == nil {
			is_end = false
			break
		}
		// 更新指针
		key_word_tmp_pt = key_word_tmp_pt.SubKeyWordTreeNodes[ch]
	}
	if is_end {
		// 前缀
		ret := make([]string, 0)
		ids := key_word_tmp_pt.KeyWordIds
		for id, _ := range ids {
			ret = append(ret, s.kv[id])
			limit --
			if limit == 0 {
				break
			}
		}
		return ret
	}
	s.rw.RUnlock()
	return make([]string, 0)
}

func (s *KeyWordServer) Search(keyword string, limit int) []string {
	s.rw.RLock()
	ids := make(map[int64]int64, 0)
	for pos, v := range keyword {
		ch := string(v)
		begins := s.char_begin_kv[ch]
		for _, begin := range begins {
			key_word_tmp_pt := begin
			next_pos := pos + 1
			for len(key_word_tmp_pt.SubKeyWordTreeNodes) > 0 && next_pos < len(keyword) {
				// 最大匹配
				next_ch := string(keyword[next_pos])
				if key_word_tmp_pt.SubKeyWordTreeNodes[next_ch] == nil {
					break
				}
				key_word_tmp_pt = key_word_tmp_pt.SubKeyWordTreeNodes[next_ch]
				next_pos++
			}
			// 保存结果
			for id, _ := range key_word_tmp_pt.KeyWordIds {
				ids[id] = ids[id] + 1
			}
		}
	}
	// 排序输出果
	list := PairList{}
	for id, count := range ids {
		list = append(list, Pair{
			K: id,
			V: count,
		})
	}
	if !sort.IsSorted(list) {
		sort.Sort(list)
	}
	if len(list) > limit {
		list = list[:limit]
	}
	ret := make([]string, 0)
	for _, item := range list {
		ret = append(ret, s.kv[item.K])
	}
	s.rw.RUnlock()
	return ret
}

func (s *KeyWordServer) DebugPrint() {
	fmt.Println("s.kv =", s.kv)
	key_word_tmp_pt := s.root
	dfs(key_word_tmp_pt)
}

func dfs(root *KeyWordTreeNode) {
	if root == nil {
		return
	} else {
		fmt.Println("s.root =", root.Char)
		fmt.Println("s.KeyWordIds =", root.KeyWordIds)
		for _, v := range root.SubKeyWordTreeNodes {
			dfs(v)
		}
	}
}
