package go123

import (
	"github.com/chucklehead-dev/gosim/gosimruntime"
)

func Maps_clone(m any) any {
	return gosimruntime.CloneMap(m)
}
