package vhll


import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"
)


func TestTriCount(t *testing.T) {
	uniqEdge, uniqueNode, _ := FileToArray("E:/Work/Datasets/streaming-dataset/p2p-Gnutella05.txt")
	/* online recording */
	vhll, _ := NewVHLL(23, 8)
	nodeDegree := make([]uint64,len(uniqueNode))
	start := time.Now() // get current time
	for i := 0; i < len(uniqEdge); i++ {
		id1 := []byte(uniqEdge[i][0])
		id2 := []byte(uniqEdge[i][1])
		vhll.Insert([]byte(id1), []byte(id2))
		vhll.Insert([]byte(id2), []byte(id1))
		idx1, _ := strconv.ParseUint(uniqEdge[i][0], 0, 64)
		nodeDegree[idx1]++
		idx2, _ := strconv.ParseUint(uniqEdge[i][1], 0, 64)
		nodeDegree[idx2]++
	}

	var triCount uint64 = 0

	/////// edge_tri_list := make(map[string]uint64)
	for i := 0; i <len(uniqEdge); i++ {
		id1 := []byte(uniqEdge[i][0])
		//w1, _ := strconv.Atoi(uniqEdge[i][0])
		card1 := vhll.Estimate(id1)
		rowTri := make([]uint64, len(uniqEdge), len(uniqEdge))   //initial rowTri

		id2 := []byte(uniqEdge[i][1])
		//w2, _ := strconv.Atoi(uniqEdge[i][1])
		card2 := vhll.Estimate(id2)
		//tmpIdx2, _ := strconv.ParseUint(uniqEdge[i][1], 0, 64)

		unionCard, _ := vhll.EstUniCard(id1, id2)
		//p3 := float64(3 * unionCard / 100)
		//float64(unionCard) > float64(card1 + card2)-p3 || float64(unionCard) < float64(card1 + card2)+p3
		if unionCard > (card1 + card2){
				rowTri[i] = 0
			} else {
				rowTri[i] = card1 + card2 - unionCard
			}
			triCount += rowTri[i]
	}

	globTriCount := triCount/3
	elapsed := time.Now().Sub(start)
	fmt.Println("time cost：", elapsed)
	//groundGlobTriCount := 727044
	groundGlobTriCount :=   1112 //105461  727044
	err := math.Abs(float64(globTriCount) - float64(groundGlobTriCount))/float64(groundGlobTriCount)
	println("global triangle count：", globTriCount)
	fmt.Printf("relative error of global triangle count：%f %% \n", err*100)
}
