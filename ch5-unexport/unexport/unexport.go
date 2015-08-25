package unexport

import (
	"fmt"
	"go/build"
	"errors"
	"sync"
	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"strings"
)

// Unexporter configuration
type Config struct {
	// Name of package that should be scanned
	Pkg string
	// Flag to include workspace packages in $GOPATH
	Workspace bool
	// Flag to include core packages
	Core bool
	// Debugging info
	Debug bool
}

// Create new default unexport configuration
func NewConfig(pkg_name string) *Config {
	return &Config{
		Pkg:       pkg_name,
		Workspace: true,
		Core:      true,
		Debug:     false,
	}
}

// Create unexporter
func NewUnexporter(o *Config) (*unexporter, error) {

	var conf loader.Config
	for _, package_name := range buildutil.AllPackages(&build.Default) {
		conf.ImportWithTests(package_name)
	}
	program, err := conf.Load()
	if err != nil {
		return nil, err
	}

	if program.Package(o.Pkg) == nil {
		return nil, errors.New(fmt.Sprintf("'%s' is not a valid package", o.Pkg))
	}

	if o.Debug {
		pkg := program.Package(o.Pkg).Pkg
		fmt.Printf("finding unused identifiers for %s (%s)\n", pkg.Name(), pkg.Path())
	}

	unexporter := &unexporter{
		Program:     program,
		MainPackage: program.Package(o.Pkg),
		Verbose:     o.Debug,
	}

	unexporter.run()

	return unexporter, nil

}

type unexporter struct {
	Program     *loader.Program
	MainPackage *loader.PackageInfo
	UsageStats  *UsageStats
	Verbose     bool
}

// Run function will collect usage data and store them in unexporter
func (unexp *unexporter) run() {

	// By default all are unused
	exported := getExportedIdentifiers(unexp.MainPackage)
	usages := identToUsages(exported, unexp.MainPackage)
	stats := &UsageStats{
		PackageInfo: unexp.MainPackage,
		Identifiers: usages,
	}

	// Run in goroutines
	var wg sync.WaitGroup
	for pkg, info := range unexp.Program.AllPackages {
		if pkg != unexp.MainPackage.Pkg {

			wg.Add(1)

			go func(info *loader.PackageInfo) {
				if unexp.Verbose {
					fmt.Println(".. scanninng ", info.Pkg.Path())
				}
				defer wg.Done()

				findUsesInPackage(info, stats)

			}(info)
		}
	}
	wg.Wait()

	if unexp.Verbose {
		for _, stat := range usages {
			fmt.Printf("identifier '%s' (%d)\n", stat.Ident.Name, len(stat.Uses))
			for ident, _ := range stat.Uses {
				file := unexp.Program.Fset.File(ident.Pos())
				fmt.Printf("\t%s in %s:%d\n", ident.Name, file.Name(), file.Line(ident.Pos()))
			}
		}
	}

	unexp.UsageStats = stats

}

// Get all identifiers that are unused
func (unexp *unexporter) GetUnusedNames() []string {

	unused := make([]string, 0)

	for ident, stats := range unexp.UsageStats.Identifiers {
		if !stats.IsUsed() {
			unused = append(unused, ident.Name)
		}
	}

	return unused

}

// Generate gorename tool commands to unexport unused exported identifiers
func (unexp *unexporter) GenerateRenameCommands() []string {
	commands := make([]string, 0)

	for _, identifier := range unexp.UsageStats.Identifiers {
		if !identifier.IsUsed() {
			pkg_path := unexp.UsageStats.PackageInfo.Pkg.Path()
			commands = append(commands, fmt.Sprintf("gorename -from '\"%s\".%s' -to %s", pkg_path, unexp.GetRenameIdentifier(identifier), identifier.GetUnexportedName()))
		}
	}

	return commands
}

// Get full identifier name. For struct it means this string will include
// struct name in identifier. I.e. MyStruct.MyMethod
func (unexp *unexporter) GetRenameIdentifier(iu *IdentifierUsage) string {

	var name string
	pkg_path := unexp.UsageStats.PackageInfo.Pkg.Path()

	switch t := unexp.MainPackage.Defs[iu.Ident].(type) {
	case *types.TypeName:
		name = t.Type().(*types.Named).String()
	case *types.Func:
		if sig := t.Type().(*types.Signature); sig != nil {
			if recv := sig.Recv(); recv != nil {
				// TODO This is hacky :S
				name = strings.TrimLeft(recv.Type().String(), "*")
				name = fmt.Sprintf("%s.%s", name, t.Name())
			}
		}
	case *types.Var:
		if t.IsField() {
			// TODO: How to find struct where it belongs
			name = fmt.Sprintf("%s.%s", pkg_path, t.Name())
//			for expr, typeandval := range unexp.MainPackage.Types {
//				if typeandval.Type != nil {
//					fmt.Println(" --> ", typeandval.Type)
//					switch ot := typeandval.Type.(type) {
//					case *types.Struct:
//						for i := 0; i < ot.NumFields(); i++ {
//							if ot.Field(i) == t {
//								fmt.Println(t)
//								fmt.Println(ot)
//								fmt.Println(expr)
//								fmt.Println(typeandval)
//							}
//						}
//					}
//				}
//			}
//			for _, def := range unexp.MainPackage.Defs {
//				if def != nil {
////					switch dt := def.Type().(type) {
////					case *types.TypeName:
////						fmt.Println(dt)
////						fmt.Println(dt.Name())
////					}
//				}
//			}
		} else {
			name = fmt.Sprintf("%s.%s", pkg_path, t.Name())
		}
	default:
		name = fmt.Sprintf("%s.%s", pkg_path, iu.Ident.Name)
	}

	return name

}