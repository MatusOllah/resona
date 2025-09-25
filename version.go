package resona

import (
	"fmt"
	"runtime/debug"
)

const root string = "github.com/MatusOllah/resona"

// Version returns the version of Resona and its checksum. The returned
// values are only valid in binaries built with module support.
//
// If a replace directive exists in the Resona go.mod, the replace will
// be reported in the version in the following format:
//
//	"version=>[replace-path] [replace-version]"
//
// and the replace sum will be returned in place of the original sum.
func Version() (version, sum string) {
	b, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}
	for _, m := range b.Deps {
		if m.Path == root {
			if m.Replace != nil {
				switch {
				case m.Replace.Version != "" && m.Replace.Path != "":
					return fmt.Sprintf("%s=>%s %s", m.Version, m.Replace.Path, m.Replace.Version), m.Replace.Sum
				case m.Replace.Version != "":
					return fmt.Sprintf("%s=>%s", m.Version, m.Replace.Version), m.Replace.Sum
				case m.Replace.Path != "":
					return fmt.Sprintf("%s=>%s", m.Version, m.Replace.Path), m.Replace.Sum
				default:
					return m.Version + "*", m.Sum + "*"
				}
			}
			return m.Version, m.Sum
		}
	}
	return "", ""
}
