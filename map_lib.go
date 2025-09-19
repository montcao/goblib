package github.com/montcao/goblib

import (
	"debug/elf" // https://pkg.go.dev/debug/elf#pkg-overview
	"fmt"
	"path/filepath"
	"strings"
	"os/exec"
	"bytes"
)

var ldconfigLines []string 

func init() {
	cmd := exec.Command("ldconfig", "-p")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}

	lines := strings.Split(string(out), "\n")
	ldconfigLines = lines

}

// resolve library names to real file paths using ldconfig -p
func ResolveLib(name string) (string, error) {
	for _, line := range ldconfigLines {
		if strings.Contains(line, name) {
			parts := strings.Fields(line)
			// parts = [libc.so.6 (libc6,AArch64, OS ABI: Linux 3.7.0) => /lib/aarch64-linux-gnu/libc.so.6]
			if len(parts) > 0 {
				path := parts[len(parts)-1]
				//fmt.Println(path)
				// ex: /lib/aarch64-linux-gnu/ld-linux-aarch64.so.1
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("could not resolve %s", name)
}

func EmptyTree()(map[string]*Node){
	return make(map[string]*Node)
}

func BuildTree(bin_path string, visited map[string]*Node) *Node {
	// Get the symbolic links from the binary
	loc, err := filepath.EvalSymlinks(bin_path)
	if err != nil {
		fmt.Println("Error")
	}
	// Convert to absolute path
	loc, err = filepath.Abs(loc)
	if err != nil {
		fmt.Println("Error")
	}

	// Return if already visited
	if node, ok := visited[loc]; ok {
		return node
	}

	// New, create BinaryMetadata to add
	bm := &BinaryMetadata{
		AbsPath: loc,
	}

	// if not, create a node from the path of the lib
	node := &Node{Path: loc, Metadata: bm}
	// attach the node to the map
	visited[loc] = node


	// Break this up to handle dynamic loaders and interpreters too

	f, err := elf.Open(loc)
	// If we can't check the binary, return the node
	if err != nil {
		fmt.Println(err)
		return node
	}
	defer f.Close() // release function when exit

	// architecture
	var arch string
	switch f.Machine {
	case elf.EM_X86_64:
		arch = "x86-64"
	case elf.EM_AARCH64:
		arch = "ARM64"
	default:
		arch = fmt.Sprintf("unknown (%d)", f.Machine)
	}
	// Update metadata
	node.Metadata.Architecture = arch


	// Dynamic loader / interpreter
	// Had to add this to handle cross platform binaries
	var interp string
	for _, prog := range f.Progs {
		if prog.Type == elf.PT_INTERP {
			data := make([]byte, prog.Filesz)
			_, err := prog.ReadAt(data, 0)
			if err != nil {
				fmt.Errorf("Failed to read PT_INTERP: %v", err)
			}
			interp = string(bytes.Trim(data, "\x00"))
			break
		}
	}
	node.Metadata.DynamicLoader = interp

	// Shared libraries
	slibs, err :=  f.ImportedLibraries() 

	if err != nil {
		fmt.Printf("No shared libraries found or error: %v", err)
	}
	node.Metadata.SharedLibraries = slibs

	// do we need ELF Type, f.Type later? revisit just in case

	// This is the recursion part, so we can get all the dependencies of dependencies...
	for _, lib := range slibs {
		resolved, err := ResolveLib(lib)
		if err != nil {
			fmt.Println(err)
			// keep unresolved name in tree for visibility
			node.Deps = append(node.Deps, &Node{Path: lib})
			continue
		}
		node.Deps = append(node.Deps, BuildTree(resolved, visited))
	}
	return node
}

func PrintFullTree(node *Node, indent string, seen map[string]bool) {
	if seen[node.Path] {
		fmt.Printf("%s↳ %s (x - already visited)\n", indent, node.Path)
		return
	}
	fmt.Printf("%s↳ %s\n", indent, node.Path)
	seen[node.Path] = true
	for i, dep := range node.Deps {
		last := i == len(node.Deps)-1
		nextIndent := indent
		if last {
			nextIndent += "    "
		} else {
			nextIndent += "│   "
		}
		PrintFullTree(dep, nextIndent, seen)
	}
}

func (node *Node) GetNodes() ([]*Node) {
	nodes := []*Node{}
    visited := make(map[string]struct{})

    var collect func(n *Node)
    collect = func(n *Node) {
        if n == nil {
            return
        }
        if _, exists := visited[n.Path]; exists {
            return
        }
        visited[n.Path] = struct{}{}
        nodes = append(nodes, n)
        for _, dep := range n.Deps {
            collect(dep)
        }
    }
    collect(node)
    return nodes
}


// Common recursive collect function for traversing the tree
func collectNodes(node *Node, getKey func(*Node) string) map[string]struct{} {
	// Make a new map of structs
	// Efficient in Go because empty structs are zero bytes
    unique := make(map[string]struct{})
    var collect func(n *Node)
    collect = func(n *Node) {
        if n == nil {
            return
        }
        key := getKey(n)
        if key == "" {
            return
        }
        if _, exists := unique[key]; exists {
            return
        }
        unique[key] = struct{}{}
        for _, dep := range n.Deps {
            collect(dep)
        }
    }
    collect(node)
    return unique
}

// Returns all the DynamicLoaders found in the tree
func (node *Node) GetDynamicLoaders() ([]string) {
	unique := collectNodes(node, func(n *Node) string { return n.Metadata.DynamicLoader })
	dynamic_loaders := make([]string, 0, len(unique)) // Create an empty array for the keys
	for path := range unique {
		dynamic_loaders = append(dynamic_loaders, path)
	}
	//fmt.Println(len(libs)) // Debug to check stability
	return dynamic_loaders

}

// Returns all the shared libraries for a binary (that was passed in as the root)
func (node *Node) GetUniqueDependencies() ([]string, error) {

	unique := collectNodes(node, func(n *Node) string { return n.Path })
	bin := node.Path
	delete(unique, bin) // Remove the binary from the list of dependencies
	//fmt.Println(bin)
	libs := make([]string, 0, len(unique)) // Create an empty array for the keys
	for path := range unique {
		libs = append(libs, path)
	}
	//fmt.Println(len(libs)) // Debug to check stability
	return libs, nil
}

// Provide the option to update all metadata with the package that the lib is in
func (node *Node) UpdateMetaWithPackage(finder PackageFinder) {
	visited := make(map[string]struct{})
	node.updateMetaWithPackageHelper(finder, visited)
}

func (node *Node) updateMetaWithPackageHelper(finder PackageFinder, visited map[string]struct{}) {
	if node == nil {
		return
	}
	if _, ok := visited[node.Path]; ok {
		return
	}
	visited[node.Path] = struct{}{}

	pkg, err := finder.FindPackage(node.Path)
	if err != nil {
		fmt.Printf("Package  %s not found\n", node.Path)
	} else {
		parse_pkg := strings.Fields(pkg)
		pkg = parse_pkg[0]
		pkg = pkg[:len(pkg)-1]
		//fmt.Println(pkg)
		node.Metadata.Package = pkg
	}
	// walk the tree
	for _, dep := range node.Deps {
		dep.updateMetaWithPackageHelper(finder, visited)
	}
}
