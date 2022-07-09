package main

import (
	"fmt"
	"github.com/tealeg/xlsx-master"
	"strconv"
)

//
func main() {
	uniqEdge, uniqueNode, _ := FileToArray("E:/Work/Datasets/streaming-dataset/p2p-Gnutella05.txt")
	vhll, _ := NewVHLL(18, 7)  //the parameter can be revised
	nodeSet := make(map[string][]string)
	for i := 0; i < len(uniqEdge); i++ {
		e1 := uniqEdge[i][0]
		e2 := uniqEdge[i][1]
		nodeSet[e1] = append(nodeSet[e1], e2) //e1 is a key, value is a string array
		nodeSet[e2] = append(nodeSet[e2], e1)
		id1 := []byte(uniqEdge[i][0])
		id2 := []byte(uniqEdge[i][1])
		vhll.Insert(id1, id2)
		vhll.Insert(id2, id1)
	}
	fmt.Println("插入vhll完成")
	var estimatedUnionNum []uint64
	var unionNumGroundTruth []uint64
	for i := 0; i < len(uniqueNode); i++ {
		e1 := uniqueNode[i]
		for j := i + 1; j < len(uniqueNode); j++ {
			e2 := uniqueNode[j]
			uniResult := UnionArray(nodeSet[e1], nodeSet[e2])
			unionNumGroundTruth = append(unionNumGroundTruth, uint64(len(uniResult))) //union set cardinality of any two adjacency set
			tmp, _ := vhll.EstUnionCard([]byte(e1), []byte(e2))
			estimatedUnionNum = append(estimatedUnionNum, tmp)
		}
		if i % 10 == 0 {
			fmt.Printf("%d\n", i)
		}
	}
	fmt.Println("union cardinality estimation finished")
	//save union cardinality to excel
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf(err.Error())
	}
	file = xlsx.NewFile()
	sheet, _ = file.AddSheet("Sheet1")
	for i := range estimatedUnionNum {
		row = sheet.AddRow()
		cell = row.AddCell()
		cell.Value = strconv.Itoa(int(estimatedUnionNum[i]))
	}
	err = file.Save("EstimatedUnionNum.xlsx")
	if err != nil {
		fmt.Printf(err.Error())
	}

	//	ground truth 存入表格
	var file1 *xlsx.File
	var sheet1 *xlsx.Sheet
	var row1 *xlsx.Row
	var cell1 *xlsx.Cell
	var err1 error

	file1 = xlsx.NewFile()
	sheet1, err1 = file1.AddSheet("Sheet1")
	if err1 != nil {
		fmt.Printf(err1.Error())
	}
	file1 = xlsx.NewFile()
	sheet1, _ = file1.AddSheet("Sheet1")

	for i := range unionNumGroundTruth {
		row1 = sheet1.AddRow()
		cell1 = row1.AddCell()
		cell1.Value = strconv.Itoa(int(unionNumGroundTruth[i]))
	}

	err1 = file1.Save("unionNumGroundTruth.xlsx")
	if err1 != nil {
		fmt.Printf(err1.Error())
	}
	fmt.Println("save to excel finished")
}

func UnionArray(a []string, b []string) []string {
	a = append(a, b...)
	var result []string
	m := make(map[string]bool)
	for _, v := range a {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}
