package translate

import (
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
)

func TestLinknamePolicy(t *testing.T) {
	for _, tc := range []struct {
		name, pkg, function, directive, want string
		body, reject                         bool
		replacements                         map[string]string
	}{
		{name: "translated standard directive", pkg: "source", function: "read", directive: "//go:linknamestd read path/to/pkg.read", replacements: map[string]string{"path/to/pkg": "translated/path/to/pkg"}, want: "//go:linkname read translated/path/to/pkg.read"},
		{name: "translated regular directive", pkg: "source", function: "read", directive: "//go:linkname read path/to/pkg.read", replacements: map[string]string{"path/to/pkg": "translated/path/to/pkg"}, want: "//go:linkname read translated/path/to/pkg.read"},
		{name: "unapproved host target", pkg: "source", function: "read", directive: "//go:linkname read runtime.unreviewed", reject: true},
		{name: "unknown declaration", pkg: "source", function: "read", directive: "//go:linkname read", reject: true},
		{name: "malformed target", pkg: "source", function: "read", directive: "//go:linkname read malformed", reject: true},
		{name: "malformed directive", pkg: "source", function: "read", directive: "//go:linkname", reject: true},
		{name: "implemented export", pkg: "source", function: "read", directive: "//go:linknamestd read", body: true, want: "//go:linkname read"},
		{name: "translated counterpart", pkg: "crypto/internal/fips140hash", function: "sha3Unwrap", directive: "//go:linkname sha3Unwrap", replacements: map[string]string{"crypto/sha3": "translated/crypto/sha3"}, want: "//go:linkname sha3Unwrap"},
		{name: "missing counterpart", pkg: "crypto/internal/fips140hash", function: "sha3Unwrap", directive: "//go:linkname sha3Unwrap", reject: true},
		{name: "outbound test implementation", pkg: "net/http", function: "protocolSetHTTP3", directive: "//go:linkname protocolSetHTTP3 golang.org/x/net/internal/http3_test.protocolSetHTTP3", body: true, want: "// gosim removed outbound test alias:"},
		{name: "outbound test declaration", pkg: "net/http", function: "protocolSetHTTP3", directive: "//go:linkname protocolSetHTTP3 golang.org/x/net/internal/http3_test.protocolSetHTTP3", reject: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			missingBodies.Lock()
			previous := len(missingBodies.list)
			missingBodies.Unlock()
			t.Cleanup(func() {
				missingBodies.Lock()
				missingBodies.list = missingBodies.list[:previous]
				missingBodies.Unlock()
			})
			decl := &dst.FuncDecl{Name: dst.NewIdent(tc.function), Type: &dst.FuncType{Params: &dst.FieldList{}}}
			decl.Decs.Start.Append(tc.directive)
			if tc.body {
				decl.Body = &dst.BlockStmt{}
			}
			translator := &packageTranslator{pkgPath: tc.pkg, replacedPkgs: tc.replacements}
			dstutil.Apply(decl, func(c *dstutil.Cursor) bool { translator.rewriteStdlibEmptyAndLinkname(c); return true }, nil)
			missingBodies.Lock()
			rejected := len(missingBodies.list) > previous
			missingBodies.Unlock()
			if rejected != tc.reject {
				t.Fatalf("rejected=%v, want %v", rejected, tc.reject)
			}
			if tc.want != "" && !strings.HasPrefix(decl.Decs.Start[0], tc.want) {
				t.Fatalf("directive %q, want %q", decl.Decs.Start[0], tc.want)
			}
			if tc.body && decl.Body == nil {
				t.Fatal("implemented function lost its body")
			}
		})
	}
}

func TestComparableHashUsesLogicalSeed(t *testing.T) {
	decl := &dst.FuncDecl{Name: dst.NewIdent("comparableHash"), Type: &dst.FuncType{Params: &dst.FieldList{}}, Body: &dst.BlockStmt{}}
	translator := &packageTranslator{pkgPath: "hash/maphash", replacedPkgs: map[string]string{hooksGo123Package: "translated/hooks"}}
	dstutil.Apply(decl, func(c *dstutil.Cursor) bool { translator.rewriteStdlibEmptyAndLinkname(c); return true }, nil)
	result := decl.Body.List[0].(*dst.ReturnStmt).Results[0].(*dst.CallExpr)
	fn := result.Fun.(*dst.Ident)
	if fn.Path != "translated/hooks" || fn.Name != "HashMaphash_Comparable" {
		t.Fatalf("unexpected comparable hook: %+v", fn)
	}
	if result.Args[0].(*dst.Ident).Name != "v" {
		t.Fatal("hash must receive value")
	}
	seed := result.Args[1].(*dst.SelectorExpr)
	if seed.X.(*dst.Ident).Name != "seed" || seed.Sel.Name != "s" {
		t.Fatal("hash must receive logical seed")
	}
}

func TestReplacementResolvesTranslatedDestination(t *testing.T) {
	for _, translated := range []bool{false, true} {
		translator := &packageTranslator{replacedPkgs: map[string]string{}}
		want := hooksGo123Package
		if translated {
			want = "translated/hooks"
			translator.replacedPkgs[hooksGo123Package] = want
		}
		ident := &dst.Ident{Path: "internal/runtime/maps", Name: "Use64BitHash"}
		translator.rewriteIdentImpl(ident)
		if ident.Path != want || ident.Name != "InternalRuntimeMaps_Use64BitHash" {
			t.Fatalf("translated=%v: replacement=%s.%s, want %s.InternalRuntimeMaps_Use64BitHash", translated, ident.Path, ident.Name, want)
		}
	}
}
