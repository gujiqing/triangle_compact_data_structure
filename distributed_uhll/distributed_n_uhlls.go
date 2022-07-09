package main

import (
	"fmt"
	."centralized_uhll/vhll"
	"math"
)

func main() {
	//simple graph
	filePath := "G:/Work/Datasets/streaming-dataset/email-Eu-core.txt"
	//输入.txt文件地址，将txt文件转化为n行2列的去除重边后的数组，n为边数目 ，每一行为一条边的两个顶点元素，
	uniqEdgeArray, _, _ := FileToArray(filePath)
	precision := 19
	vPrecision := 5 // precision 是m的值; 	vPrecision是s的值
	totalVhlls := 4 //设置的分布式vhll的个数
	Vhlls := make([]*VHLL, totalVhlls)
	for i := 0; i < totalVhlls; i++ {
		Vhlls[i], _ = NewVHLL(uint8(precision), uint8(vPrecision))
	}
	////开辟两块vhll空间
	//FirstVhll, _ := NewVHLL(uint8(precision), uint8(vPrecision))
	//SecondVhll, _ := NewVHLL(uint8(precision), uint8(vPrecision))

	nodeFlag := make(map[string][]int)
	//  设置一个map, Key为顶点的时候，value是vhll编号，查询的涉及value数组去重
	for i, edge := range uniqEdgeArray {
		//轮询的方式存入vhll
		arrangeNum := i % totalVhlls
		Vhlls[arrangeNum].Insert([]byte(edge[0]), []byte(edge[1]))
		Vhlls[arrangeNum].Insert([]byte(edge[1]), []byte(edge[0]))

		value0, _ := nodeFlag[edge[0]]
		value0 = append(value0, arrangeNum)
		nodeFlag[edge[0]] = value0

		value1, _ := nodeFlag[edge[1]]      //取出map中的value值
		value1 = append(value1, arrangeNum) //添加0到value里面
		nodeFlag[edge[1]] = value1          //将value赋值

	}

	//
	var triCount uint64 = 0

	//markVhll := [] *VHLL {FirstVhll, SecondVhll}
	for i, edge := range uniqEdgeArray {
		id := edge //
		Card := make([]uint64, 2)
		var Array [][]uint8
		var Array1 []uint8
		var allNum []int
		for j := 0; j < 2; j++ {
			// 计算一条边的两个顶点的邻接节点的基数
			storeVhllNum := removeDup(nodeFlag[id[j]]) // 第j个顶点的邻接节点被存储的vhll编号
			allNum = append(allNum, storeVhllNum...)
			if len(storeVhllNum) == 1 {
				tmpvhll := Vhlls[storeVhllNum[0]]
				Card[j] = Estim([]byte(id[j]), tmpvhll)
				Array1 = tmpvhll.GetArray([]byte(id[j]))
			} else {
				MarkVhll := make([]*VHLL, len(storeVhllNum))
				var arraySet [][]uint8
				for k, VhllNum := range storeVhllNum {
					MarkVhll[k] = Vhlls[VhllNum] // 将一个顶点的邻接节点存储的vhll地址，放到一个数组里
					tmpArray := MarkVhll[k].GetArray([]byte(id[j]))
					arraySet = append(arraySet, tmpArray)
				}
				Array1 = UniArray(arraySet)
				//Card[j], _ = EsUnCa([]byte(id[j]), []byte(id[j]), Array1, Vhlls)
				Card[j], _ = EsUnCa([]byte(id[j]), []byte(id[j]), Array1, MarkVhll)
			}
			Array = append(Array, Array1)
		}

		rowTri := make([]uint64, len(uniqEdgeArray)) //初始化rowTri
		uniArray := UnionArray(Array[0], Array[1])

		//	算两个顶点对应的多个集合的并集
		uniqAllNum := removeDup(allNum)
		uniqueVhll := make([]*VHLL, len(uniqAllNum))
		for h, Num := range uniqAllNum {
			uniqueVhll[h] = Vhlls[Num] // 将一个顶点的邻接节点存储的vhll地址，放到一个数组里
		}
		unionCard, _ := EsUnCa([]byte(id[0]), []byte(id[1]), uniArray, uniqueVhll)
		//unionCard, _ := EsUnCa([]byte(id[0]), []byte(id[1]), uniArray, Vhlls)
		if unionCard > (Card[0] + Card[1]) {
			rowTri[i] = 0
		} else {
			rowTri[i] = Card[0] + Card[1] - unionCard
		}
		triCount += rowTri[i]
	}

	globTriCount := triCount / 3
	//groundGlobTriCount := 727044
	groundGlobTriCount := 105461 //1112
	err := math.Abs(float64(globTriCount)-float64(groundGlobTriCount)) / float64(groundGlobTriCount)
	fmt.Printf("Global triangle count：%d  \n", globTriCount)

	fmt.Printf("erroe of global triangle count：%f %% \n", err*100)
}

//
func removeDup(arr []int) (newArr []int) {
	newArr = make([]int, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}
