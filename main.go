package main

import (
	"fmt"
	"github.com/flashfeifei/dictmatch_tries"
)

func main() {

	s := dictmatch_tries.NewKeyWordServer()
	//加入数据
	s.Put(1, "五月天")
	s.Put(2, "澳门皇家赌场")
	s.Put(3, "性感荷官")
	s.Put(4, "在线发牌")
	s.Put(5, "AV")
	s.Put(6, "澳门顶级赌场")
	//s.DebugPrint()
	//fmt.Println(s.Sugg("ba", 2))
	//fmt.Println(s.Search("a", 2))
	fmt.Println(s.Search("荷官", 4))
	//fmt.Println(s.Search("c", 2))
	//fmt.Println(s.Search("ba", 2))
}
