package dictmatch_tries

import (
	"fmt"
	"sort"
	"sync"
)

//字典树的整个字符串的存储容器
type KeyWordKV map[int64]string

type CharBeginKV map[string][]*KeyWordTreeNode

//用于结果排序，实现排序接口
type PairList []Pair

// Len方法返回集合中的元素个数
func (p PairList) Len() int {
	return len(p)
}

// Less方法报告索引i的元素是否比索引j的元素小
func (p PairList) Less(i, j int) bool {
	return p[i].V > p[j].V
}

// Swap方法交换索引i和j的两个元素
func (p PairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type Pair struct {
	K int64
	V int64
}

//字典树的每个节点
type KeyWordTreeNode struct {
	//猜测一下是用来保存key的主要，用来直接提取字符串用
	//这个字，都可以匹配到什么字符串,map中的key是string的索引
	//通过key直接去获取KeyWordKV保存的字符串
	//保存的是这个字，可以在哪些字符串中出现!!!!
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

//对一颗字典树的操作的封装
type KeyWordServer struct {
	root          *KeyWordTreeNode //字典树
	kv            KeyWordKV        //这是什么？
	char_begin_kv CharBeginKV      //这是什么？
	rw            sync.RWMutex     //读写锁,为了线程安全呗
}

func NewKeyWordServer() *KeyWordServer {
	return &KeyWordServer{
		root:          NewKeyWordTreeNode(),
		kv:            KeyWordKV{},
		char_begin_kv: CharBeginKV{},
	}
}

//添加一个字符串，进字典树
func (s *KeyWordServer) Put(id int64, keyword string) {
	s.rw.Lock()
	defer s.rw.Unlock()

	//存储这个字符串，作用是为了搜索的时候通过id直接获取值
	s.kv[id] = keyword
	//获取树的根节点
	key_word_tmp_pt := s.root

	//往这棵字典树导入这个字符串的数据
	for _, v := range keyword {
		ch := string(v)

		if key_word_tmp_pt.SubKeyWordTreeNodes[ch] == nil {
			node := NewKeyWordTreeNodeWithParams(ch, key_word_tmp_pt)
			//挂在子节点上去
			key_word_tmp_pt.SubKeyWordTreeNodes[ch] = node
			//记住统一字符的不同位置
			s.char_begin_kv[ch] = append(s.char_begin_kv[ch], node)

		}

		//获取子节点的位置
		key_word_tree_node := key_word_tmp_pt.SubKeyWordTreeNodes[ch]
		//卧槽，这又是干嘛的
		key_word_tree_node.KeyWordIds[id] = true
		// 更新当前指针的位置，让本次字符串继续往下走
		key_word_tmp_pt = key_word_tmp_pt.SubKeyWordTreeNodes[ch]
	}
}

//完整字符串前开始搜索功能，limit是限制可以输出多少值
func (s *KeyWordServer) Sugg(keyword string, limit int) []string {
	s.rw.RLock()
	defer s.rw.RUnlock()

	//从头开始，you know ？
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

	//搜索是否到这个字符串的字典树的尾部
	if is_end {
		//没有到尾部
		ret := make([]string, 0)
		//获取这个值有什么用 ????
		ids := key_word_tmp_pt.KeyWordIds
		for id, _ := range ids {
			//获取值
			ret = append(ret, s.kv[id])
			limit --
			if limit == 0 {
				break
			}
		}
		return ret
	}

	return make([]string, 0)
}

//这个搜索功能又是啥？？
func (s *KeyWordServer) Search(keyword string, limit int) []string {
	s.rw.RLock()
	defer s.rw.RUnlock()

	ids := make(map[int64]int64, 0)

	//执行字符串搜索，（最大匹配搜索）
	for pos, v := range keyword {
		ch := string(v)
		//获取这个字符在字典树出现的位置
		begins := s.char_begin_kv[ch]
		//遍历这个字符开头的子节点
		for _, begin := range begins {
			//从这个字符的头开始
			key_word_tmp_pt := begin
			//这个是什么作用???
			next_pos := pos + 1

			for len(key_word_tmp_pt.SubKeyWordTreeNodes) > 0 && next_pos < len(keyword) {
				// 最大匹配,有点模糊的匹配
				next_ch := string(keyword[next_pos])
				if key_word_tmp_pt.SubKeyWordTreeNodes[next_ch] == nil {
					break
				}
				key_word_tmp_pt = key_word_tmp_pt.SubKeyWordTreeNodes[next_ch]
				next_pos++
			}

			// 保存搜索结果
			for id, _ := range key_word_tmp_pt.KeyWordIds {
				ids[id] = ids[id] + 1
			}
		}
	}

	// 排序输出果
	// 构建排序对象
	list := PairList{}
	for id, count := range ids {
		list = append(list, Pair{
			K: id,
			V: count,
		})
	}

	//排序
	if !sort.IsSorted(list) {
		//排序，对象需要实现golang的排序接口
		sort.Sort(list)
	}

	//限制输出
	if len(list) > limit {
		list = list[:limit]
	}

	//按照排序结构输出值
	ret := make([]string, 0)
	for _, item := range list {
		ret = append(ret, s.kv[item.K])
	}
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
