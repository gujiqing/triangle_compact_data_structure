package main
import (
	"fmt"
	xlsx "github.com/tealeg/xlsx-master"
	"math"
	"strconv"
	"testing"
)

type float64t float64

func TestUnionVHLL(t *testing.T) {
	vhll, _ := NewVHLL(16, 8) //初始 18, 12

	var (
		unionFlow [100][100]int
		expected = make(map[uint]uint)
		Err [100][100]float64
	)
	//模拟100条大小不同的流，从1000到100,000，每1000递增
	for k := uint(0); k < 100; k++ {
		id := []byte(strconv.Itoa(int(k)))
		for i := uint(0); i < k*1000; i++ {
			vhll.Insert([]byte(id), []byte(strconv.Itoa(int(i))))
		}
		expected[k] = 1000*(k+1)
	}
	var sum float64 = 0
	//100条流两两合并
	for i := uint(0); i < 100; i++ {
		for j :=  uint(i + 1); j < 100; j++ {
			id1 := []byte(strconv.Itoa(int(i)))
			id2 := []byte(strconv.Itoa(int(j)))
			unionCard, _ := vhll.EstUnionCard(id1, id2)
			unionFlow [i][j] = int(unionCard)
			err:= math.Abs(float64(unionCard) -float64(expected[j])) / float64(expected[j])
			Err[i][j] = err
			sum += err
		}
	}
	fmt.Printf("相对误差为%f %% \n",  100*sum/(100*(100-1)/2))

//union flow 存入表格
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
	result := unionFlow
	file = xlsx.NewFile()
	sheet, _ = file.AddSheet("Sheet1")
	for i, rowcoent := range result {
		row = sheet.AddRow()
		for j, _ := range rowcoent {
			cell = row.AddCell()
			cell.Value = strconv.Itoa(result[i][j])
		}
	}
	err = file.Save("estimated1.xlsx")
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
	for _, rowcoent1 := range result  {
		//fmt.Print(rowcoent)
		row1 = sheet1.AddRow()
		for j, _ := range rowcoent1 {
			cell1 = row1.AddCell()
			cell1.Value = strconv.Itoa((j + 1) * 1000)
		}
	}
	err1 = file1.Save("expected1.xlsx")
	if err1 != nil {
		fmt.Printf(err1.Error())
	}
}
