package unexport

import (
	"go/ast"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"sync"
	"strings"
)

// Represents usage stats for exported identifiers found in package
type UsageStats struct {
	// Package details
	PackageInfo *loader.PackageInfo
	// Keyed by exported identifier and value is usage data
	Identifiers map[*ast.Ident]*IdentifierUsage
}

// Represents usage map of exported identifiers
type IdentifierUsage struct {
	Ident *ast.Ident
	Uses  map[*ast.Ident]types.Object
	mutex sync.Mutex
}

func (iu *IdentifierUsage) AddUse(ident *ast.Ident, object types.Object) {
	iu.mutex.Lock()
	defer iu.mutex.Unlock()
	iu.Uses[ident] = object
}

// Check if identifier is used
func (iu *IdentifierUsage) IsUsed() bool {
	return len(iu.Uses) > 0
}

// Return identifier unexported name (first lowercase)
func (iu *IdentifierUsage) GetUnexportedName() string {
	lower := []byte(iu.Ident.Name)
	lower[0] = strings.ToLower(string(lower[0]))[0]
	return string(lower)
}

// Filters exposed package identifiers
func getExportedIdentifiers(info *loader.PackageInfo) []*ast.Ident {
	exported := make([]*ast.Ident, 0)
	for identifier, _ := range info.Defs {
		if identifier.IsExported() {
			exported = append(exported, identifier)
		}
	}
	return exported
}

// Convert identifiers to stats
func identToUsages(idents []*ast.Ident, info *loader.PackageInfo) map[*ast.Ident]*IdentifierUsage {
	usages := make(map[*ast.Ident]*IdentifierUsage, 0)
	for _, ident := range idents {
		usages[ident] = &IdentifierUsage{
			Ident: ident,
			Uses:  make(map[*ast.Ident]types.Object, 0),
		}
	}
	return usages
}

// Find usage of identifies in given package
func findUsesInPackage(info *loader.PackageInfo, stats *UsageStats) {
	// For each known exported identifier for searched package
	for _, ident_stats := range stats.Identifiers {
		// Check if its found in scanned package Uses
		for ident, use := range info.Uses {
			if use.Pkg() != nil && use.Pkg().Path() == stats.PackageInfo.Pkg.Path() {
				if use.Pos() == ident_stats.Ident.Pos() {
					// Record identifier use in identifier stats
					ident_stats.AddUse(ident, use)
				}
			}
		}
	}
}
