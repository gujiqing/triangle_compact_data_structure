package main

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

/*  read data from .txt， store the data by 2-dim slice */
func FileToArray(filePath string) ([][]string, []string, error) {
	var keysResult [][]string

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return keysResult, nil, err
	}
	s := string(b)
	for _, lineStr := range strings.Split(s, "\n") {
		lineStr = strings.TrimSpace(lineStr)
		//fmt.Println(lineStr, reflect.TypeOf(lineStr), len(lineStr))
		if lineStr == "" {
			continue
		}
		//keys := strings.Split(result[i], "	")
		var keys []string
		for _, key := range strings.Split(lineStr, "	") {//这一行很重要，有的两个顶点之间是空格隔开，有的是tab键
			keys = append(keys, key)
		}
		keysResult = append(keysResult, keys)
	}
	uniqEdgeArray := UniqueEdgeArray(keysResult)
	uniqueNode := UniqueNode(keysResult)
	return uniqEdgeArray, uniqueNode, err
}

// 通过map主键唯一的特性过滤重复元素
//返回值是数据流中所有元素的集合，用字符串数组表示
func UniqueNode(slc [][]string) []string {
	var nodesInGraph []string
	tempMap := map[string]byte{}// 存放不重复主键,这里为了节省内存，使用map[int]byte。 因为map的value并没有用到，所以什么类型都可以。
	for i := 0; i < len(slc); i++ {
		for j := 0; j < 2; j++ {
			e := slc[i][j]
			l := len(tempMap)
			tempMap[e] = 0
			if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
				nodesInGraph = append(nodesInGraph, e)
			}
		}

	}
	return nodesInGraph
}

func UniqueEdgeArray(slc [][]string) [][]string {
	var uniqueEdge [][]string
	uniqueNode := UniqueNode(slc) //对分割后的数据集去重，得到一个一维切片，即图数据流中顶点集合
	nodeNum := len(uniqueNode)
	var intNode []int
	//transfer to type int
	for _, node := range uniqueNode {
		inode, _:= strconv.Atoi(node)
		intNode = append(intNode, inode)
	}
	sort.Ints(intNode)
	maxNode := intNode[len(intNode)-1]

	var neiborMatrix = make([][]int, maxNode+1,maxNode+1)
	for i := 0; i < nodeNum; i++ {
		neiborMatrixLine := make([]int, maxNode+1, maxNode+1)
		neiborMatrix[i] = neiborMatrixLine
	}


	for i := 0; i < len(slc); i++ {
			e1, _ := strconv.Atoi(slc[i][0])
			e2, _ := strconv.Atoi(slc[i][1])
			neiborMatrix[e1][e2] = 1
			neiborMatrix[e2][e1] = 1
	}
	for i := 0; i < len(neiborMatrix); i++ {
		for j := i + 1; j < len(neiborMatrix); j++ {
			var trmstr []string
			if neiborMatrix[i][j] == 1 {
				e1 := strconv.Itoa(i)
				e2 := strconv.Itoa(j)
				trmstr = append(trmstr, e1, e2)
				uniqueEdge = append(uniqueEdge, trmstr)
			}
		}
	}
	return uniqueEdge
}