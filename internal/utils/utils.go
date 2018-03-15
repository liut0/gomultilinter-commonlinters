package utils

import (
	"go/token"

	"github.com/liut0/gomultilinter/api"
)

// GetPkgPos tries to return the *token.Position of the pkg.
// Since only ast level info is present this only works if at least one
// go file is present in the pkgs dir, otherwise nil is returned
func GetPkgPos(pkg *api.Package) *token.Position {
	if len(pkg.PkgInfo.Files) == 0 {
		return nil
	}

	pos := pkg.FSet.Position(pkg.PkgInfo.Files[0].Pos())
	return &pos
}

// StringArrayToMap creates a map and sets true for every element
// of the array
func StringArrayToMap(vs []string) map[string]bool {
	m := make(map[string]bool, len(vs))
	for _, v := range vs {
		m[v] = true
	}
	return m
}
