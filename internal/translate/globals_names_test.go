package translate

import (
	"go/ast"
	"go/types"
	"testing"
)

func TestGlobalsAccessorAvoidsShadowing(t *testing.T) {
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
	original := globalsAccessorName("crypto/internal/fips140/mlkem", false, info)
	info.Defs[ast.NewIdent("G")] = nil
	if got := globalsAccessorName("crypto/internal/fips140/mlkem", false, info); got != original {
		t.Fatalf("ordinary local changed stable name: %s", got)
	}
	info.Defs[ast.NewIdent(original)] = nil
	info.Uses[ast.NewIdent(original+"_")] = nil
	if got := globalsAccessorName("crypto/internal/fips140/mlkem", false, info); got != original+"__" {
		t.Fatalf("accessor collided with source identifier: %s", got)
	}
	if got := globalsAccessorName("crypto/internal/fips140/mlkem", true, info); got == original {
		t.Fatal("test globals must have separate accessor")
	}
	if got := globalsAccessorName("other/package", false, info); got == original {
		t.Fatal("package accessors must be distinct")
	}
}
