package go123

import "github.com/chucklehead-dev/gosim/gosimruntime"

func MathRand_runtime_rand() uint64 {
	return gosimruntime.Fastrand64()
}
