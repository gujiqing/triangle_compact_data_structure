package vhll

import (
	"math"
	"metro"
)

func alpha(m float64) float64 {
	switch m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	}
	return 0.7213 / (1 + 1.079/m)
}

func zeros(registers []uint8) (z float64) {
	for _, val := range registers {
		if val == 0 {
			z++
		}
	}
	return z
}

func beta(ez float64) float64 {
	zl := math.Log(ez + 1)
	return -0.370393911*ez +
		0.070471823*zl +
		0.17393686*math.Pow(zl, 2) +
		0.16339839*math.Pow(zl, 3) +
		-0.09237745*math.Pow(zl, 4) +
		0.03738027*math.Pow(zl, 5) +
		-0.005384159*math.Pow(zl, 6) +
		0.00042419*math.Pow(zl, 7)
}

// Calculate the position of the leftmost 1-bit.
func rho(val uint64) (r uint8) {
	for val&0x8000000000000000 == 0 {
		val <<= 1
		r++
	}
	return r + 1
}

func hash(e []byte) uint64 {
	return metro.Hash64(e, 1337)
}

func sumAndZeros(register []uint8) (float64, float64) {
	ez := 0.0
	sum := 0.0
	for _, val := range register {
		sum += 1.0 / math.Pow(2.0, float64(val))
		if val == 0 {
			ez++
		}
	}
	return sum, ez
}

// VHLL ...
type VHLL struct {
	M      []uint8
	m      uint64
	s      uint64
	log2s  uint64
	mAlpha float64
	sAlpha float64
}

func (v *VHLL) hashi(i uint64, f []byte) uint64 {
	return metro.Hash64(f, i) % v.m
}

// NewVHLL ...
func NewVHLL(precision, vPrecision uint8) (*VHLL, error) {
	//if precision < 9 {
	//	return nil, errors.New("precision needs to be >= 9")
	//}
	//if vPrecision < 8 || vPrecision > 12 {
	//	return nil, errors.New("virtual precision needs to be >= 8 and <= 10")
	//}
	//if precision < vPrecision {
	//	return nil, errors.New("virtual precision needs to be > precision")
	//}
	m := uint64(math.Pow(2, float64(precision)))
	s := uint64(math.Pow(2, float64(vPrecision)))
	return &VHLL{
		M:      make([]uint8, m, m),
		m:      m,
		s:      s,
		log2s:  uint64(vPrecision),
		mAlpha: alpha(float64(m)),
		sAlpha: alpha(float64(s)),
	}, nil
}

// Insert ...
func (v *VHLL) Insert(f []byte, e []byte) {
	he := hash(e)
	p := he % v.s
	q := he << v.log2s
	r := rho(q)
	index := metro.Hash64(f, p) % v.m
	if r > v.M[index] {
		v.M[index] = r
	}
}

// Estimate ...
//func (v *VHLL) Estimate(f []byte) uint64 {
func (v *VHLL) Estimate(f []byte) uint64 {
	M := make([]uint8, v.s, v.s)
	for i := range M {
		index := metro.Hash64(f, uint64(i)) % v.m
		M[i] = v.M[index]
	}

	sum, ez := sumAndZeros(M)
	s := float64(v.s)
	beta := beta(ez)
	ns := v.sAlpha * s * (s - ez) / (beta + sum)

	// estimate error
	m := float64(v.m)
	n := float64(v.totalCardinality())
	//e := ns - (s * n / m)
	e := m * s / (m - s) *(ns/s -n/m)
	//如果估计值较小的情况下，e < 2.5*s，则当成bitmap来计算
	if e < 8{
		zeroCount :=0
		for i := range M {
			if M[i] >0 {
				M[i] = 1
			} else {
				M[i] = 0
				zeroCount ++
			}
		}
		tmp := float64(zeroCount)
		lenm:= float64(len(M))
		v1 := tmp/lenm
		ns = - s * math.Log2(v1)
		e = m * s / (m - s) *(ns/s -n/m)
	}
	// rounding
	return uint64(e + 0.5)
}

func (v *VHLL) totalCardinality() uint64 {
	sum, ez := sumAndZeros(v.M)
	m := float64(len(v.M))
	beta := beta(ez)
	return uint64(v.mAlpha * m * (m - ez) / (beta + sum))
}

//计算合并后的基数，以及 standard error
func EstUnionCard(f1,f2 []byte, v []*VHLL) (uint64, float64){
	v1 := v[0]
	v2 := v[1]
	M := getUnionArray2(f1, f2, v)
	sum, ez := sumAndZeros(M)
	s := float64(v1.s)
	beta := beta(ez)
	ns := v1.sAlpha * s * (s - ez) / (beta + sum)

	// estimate error
	m := float64(v1.m)
	n := float64(v1.totalCardinality() + v2.totalCardinality())
	//e := ns - (s * n / m)
	e := (m - 1) * s / (m - 1 - s) *(ns/s -n/(m-1))
	//if e < 128{
	//	zeroCount :=0
	//	for i := range M {
	//		if M[i] >0 {
	//			M[i] = 1
	//		} else {
	//			M[i] = 0
	//			zeroCount ++
	//		}
	//	}
	//	tmp := float64(zeroCount)
	//	lenm:= float64(len(M))
	//	v := tmp/lenm
	//	ns := - s * math.Log2(v)
	//	e = (m - 1) * s / (m - 1 - s) *(ns/s -n/(m-1))
	//}

	// 求 standard error
	x1 := math.Pow((m-1)/(m-1-s), 2)
	x2 := math.Pow(1.04,2)/ s * math.Pow(e + (n - e)*s/(m- 1),2)
	x3 := (math.Pow(1.04, 2)/s + 1) * (n - e) * (s /(m - 1)) * (1 - (s /(m - 1)))
	x4 := math.Pow(s/(m- 1), 2) * math.Pow(1.04*n,2)/ m
	Var := x1 * x2 + x3 + x4
	strErr := math.Sqrt(Var)
	// rounding
	return uint64(e + 0.5), strErr
}

func EsUnCa(f1,f2 []byte, M []uint8, v [] *VHLL) (uint64, float64){
	// 一条边的两个顶点，存储在多个vhll，多个vhll合并, M 为合并后的矩阵
	v1 := v[0]
	sum, ez := sumAndZeros(M)
	s := float64(v1.s)
	beta := beta(ez)
	ns := v1.sAlpha * s * (s - ez) / (beta + sum)

	// estimate error
	m := float64(v1.m)
	var n  float64 = 0
	for i := range v {
		n = n + float64(v[i].totalCardinality())
	}
	e := (m - 1) * s / (m - 1 - s) *(ns/s -n/(m-1))

	// 求 standard error
	x1 := math.Pow((m-1)/(m-1-s), 2)
	x2 := math.Pow(1.04,2)/ s * math.Pow(e + (n - e)*s/(m- 1),2)
	x3 := (math.Pow(1.04, 2)/s + 1) * (n - e) * (s /(m - 1)) * (1 - (s /(m - 1)))
	x4 := math.Pow(s/(m- 1), 2) * math.Pow(1.04*n,2)/ m
	Var := x1 * x2 + x3 + x4
	strErr := math.Sqrt(Var)
	// rounding
	return uint64(e + 0.5), strErr
}
//两个矩阵合并后的矩阵
func getUnionArray2(f1,f2 []byte, v []*VHLL) []uint8 {
	//两个vhll种的矩阵合并

	v1 := v[0]
	//fmt.Printf("v1 的类型 %T\n", v1)
	v2 := v[1]
	M1 := v1.GetArray(f1)
	M2 := v2.GetArray(f2)
	//var MaxM uint8 = math.Max(M1, M2)
	//_, M1 := v.Estimate(f1)
	//_, M2 := v.Estimate(f2)
	MaxM := make([]uint8, len(M1), cap(M1))
	for i := range M1 {
		//对应位置取大值
		if M1[i] >= M2[i]{
			MaxM[i] = M1[i]
		} else{
			MaxM[i] = M2[i]
	}
	//另一种思路
	//MaxM[i] = M1[i] + M2[i]
	}
	return MaxM
}
func unionArray(f1,f2 []byte, v []*VHLL) []uint8 {
	//两个vhll种的矩阵合并

	v1 := v[0]
	//fmt.Printf("v1 的类型 %T\n", v1)
	v2 := v[1]
	M1 := v1.GetArray(f1)
	M2 := v2.GetArray(f2)
	//var MaxM uint8 = math.Max(M1, M2)
	//_, M1 := v.Estimate(f1)
	//_, M2 := v.Estimate(f2)
	MaxM := make([]uint8, len(M1), cap(M1))
	for i := range M1 {
		//对应位置取大值
		if M1[i] >= M2[i]{
			MaxM[i] = M1[i]
		} else{
			MaxM[i] = M2[i]
		}
		//另一种思路
		//MaxM[i] = M1[i] + M2[i]
	}
	return MaxM
}
func (v *VHLL) getUnionArray1(f1,f2 []byte) []uint8 {
	// 一个vhll 中)的两个矩阵合并

	M1 := v.GetArray(f1)
	M2 := v.GetArray(f2)
	//var MaxM uint8 = math.Max(M1, M2)
	//_, M1 := v.Estimate(f1)
	//_, M2 := v.Estimate(f2)
	MaxM := make([]uint8, len(M1), cap(M1))
	for i := range M1 {
		if M1[i] >= M2[i]{
			MaxM[i] = M1[i]
		} else{
			MaxM[i] = M2[i]
		}
	}
	return MaxM
}
//得到一个流的vhll矩阵
func (v *VHLL) GetArray(f []byte) []uint8 {
	//var i *int
	//i = new(int)
	//*i = 1
	//v = new(VHLL)
	M := make([]uint8, v.s, v.s)
	for i := range M {
		index := metro.Hash64(f, uint64(i)) % v.m
		M[i] = v.M[index]
	}
	return M
}

func UnionArray(M1,M2 []uint8) []uint8 {
	// 一个vhll 中)的两个矩阵合并

	//M1 := v.GetArray(f1)
	//M2 := v.GetArray(f2)
	//var MaxM uint8 = math.Max(M1, M2)
	//_, M1 := v.Estimate(f1)
	//_, M2 := v.Estimate(f2)
	MaxM := make([]uint8, len(M1), cap(M1))
	for i := range M1 {
		if M1[i] >= M2[i]{
			MaxM[i] = M1[i]
		} else{
			MaxM[i] = M2[i]
		}
	}
	return MaxM
}

func UniArray(M[][]uint8) []uint8 {
	// 一个vhll 中)的两个矩阵合并

	//M1 := v.GetArray(f1)
	//M2 := v.GetArray(f2)
	//var MaxM uint8 = math.Max(M1, M2)
	//_, M1 := v.Estimate(f1)
	//_, M2 := v.Estimate(f2)
	MaxM := M[0]
	for i := 1; i<len(M); i++ {
		MaxM = UnionArray(MaxM,M[i])
	}
	return MaxM
}



func (v *VHLL) EstUniCard(f1,f2 []byte) (uint64, float64){
	M := v.getUnionArray1(f1, f2)
	sum, ez := sumAndZeros(M)
	s := float64(v.s)
	beta := beta(ez)
	ns := v.sAlpha * s * (s - ez) / (beta + sum)

	// estimate error
	m := float64(v.m)
	n := float64(v.totalCardinality())
	//e := ns - (s * n / m)
	e := (m - 1) * s / (m - 1 - s) *(ns/s -n/(m-1))
	if e < 64{
		zeroCount :=0
		for i := range M {
			if M[i] >0 {
				M[i] = 1
			} else {
				M[i] = 0
				zeroCount ++
			}
		}
		tmp := float64(zeroCount)
		lenm:= float64(len(M))
		v := tmp/lenm
		ns := - s * math.Log2(v)
		e = (m - 1) * s / (m - 1 - s) *(ns/s -n/(m-1))
	}

	// 求 standard error
	x1 := math.Pow((m-1)/(m-1-s), 2)
	x2 := math.Pow(1.04,2)/ s * math.Pow(e + (n - e)*s/(m- 1),2)
	x3 := (math.Pow(1.04, 2)/s + 1) * (n - e) * (s /(m - 1)) * (1 - (s /(m - 1)))
	x4 := math.Pow(s/(m- 1), 2) * math.Pow(1.04*n,2)/ m
	Var := x1 * x2 + x3 + x4
	strErr := math.Sqrt(Var)
	// rounding
	return uint64(e + 0.5), strErr
}
func (v *VHLL) Simple_estUniCard(f1,f2 []byte) (uint64, float64){
	M := v.getUnionArray1(f1, f2)
	sum, ez := sumAndZeros(M)
	s := float64(v.s)
	beta := beta(ez)
	ns := v.sAlpha * s * (s - ez) / (beta + sum)

	// estimate error
	m := float64(v.m)
	n := float64(v.totalCardinality())
	//e := ns - (s * n / m)
	e := (m - 1) * s / (m - 1 - s) *(ns/s -n/(m-1))
	if e < 64{
		zeroCount :=0
		for i := range M {
			if M[i] >0 {
				M[i] = 1
			} else {
				M[i] = 0
				zeroCount ++
			}
		}
		tmp := float64(zeroCount)
		lenm:= float64(len(M))
		v := tmp/lenm
		ns := - s * math.Log2(v)
		e = (m - 1) * s / (m - 1 - s) *(ns/s -n/(m-1))
	}

	// 求 standard error
	x1 := math.Pow((m-1)/(m-1-s), 2)
	x2 := math.Pow(1.04,2)/ s * math.Pow(e + (n - e)*s/(m- 1),2)
	x3 := (math.Pow(1.04, 2)/s + 1) * (n - e) * (s /(m - 1)) * (1 - (s /(m - 1)))
	x4 := math.Pow(s/(m- 1), 2) * math.Pow(1.04*n,2)/ m
	Var := x1 * x2 + x3 + x4
	strErr := math.Sqrt(Var)
	// rounding
	return uint64(e + 0.5), strErr
}

func Estim(f []byte, v *VHLL) uint64 {
	M := make([]uint8, v.s, v.s)
	for i := range M {
		index := metro.Hash64(f, uint64(i)) % v.m
		M[i] = v.M[index]
	}

	sum, ez := sumAndZeros(M)
	s := float64(v.s)
	beta := beta(ez)
	ns := v.sAlpha * s * (s - ez) / (beta + sum)

	// estimate error
	m := float64(v.m)
	n := float64(v.totalCardinality())
	//e := ns - (s * n / m)
	e := m * s / (m - s) *(ns/s -n/m)
	//如果估计值较小的情况下，e < 2.5*s，则当成bitmap来计算
	if e < 64 {
		zeroCount :=0
		for i := range M {
			if M[i] >0 {
				M[i] = 1
			} else {
				M[i] = 0
				zeroCount ++
			}
		}
		tmp := float64(zeroCount)
		lenm:= float64(len(M))
		v := tmp/lenm
		ns := - s * math.Log2(v)
		e = m * s / (m - s) *(ns/s -n/m)
	}
	// rounding
	return uint64(e + 0.5)
}

func (v *VHLL) EstUC(f1,f2 []byte) (uint64, float64){
	//未移除噪音的时的并集
	M := v.getUnionArray1(f1, f2)
	sum, ez := sumAndZeros(M)
	s := float64(v.s)
	beta := beta(ez)
	ns := v.sAlpha * s * (s - ez) / (beta + sum)
	return uint64(ns), 0
}