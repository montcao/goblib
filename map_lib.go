package goblib

import (
	"debug/elf"
	"fmt"
	"path/filepath"
	"strings"
	"os/exec"
)

var ldconfigLines []string 
type Node struct {
	Path string  `json:"path"`
	Deps []*Node `json:"deps,omitempty"`
}

type Forest []*Node // multiple bin support TODO

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

	// if not, create a node from the path of the lib
	node := &Node{Path: loc}
	// attach the node to the map
	visited[loc] = node

	f, err := elf.Open(loc)
	// If we can't check the binary, return the node
	if err != nil {
		fmt.Println(err)
		return node
	}
	defer f.Close() // release function when exit

	// Get all the libs that the bin depends on
	libs, err := f.ImportedLibraries()
	if err != nil {
		fmt.Println(err)
		return node
	}

	// This is the recursion part, so we can get all the dependencies of dependencies...
	// Supply chain baby
	for _, lib := range libs {
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


func (node *Node) GetUniqueDependencies() ([]string, error) {
	// Make a new map of structs
	// Efficient in Go because empty structs are zero bytes
	unique := make(map[string]struct{}) 
	// Declare a recursive helper function
	var collect func(n *Node)

	collect = func(n *Node) {
		if n == nil {
			return
		}
		if _, exists := unique[n.Path]; exists {
			return
		}
		unique[n.Path] = struct{}{} // populate the keys
		for _, dep := range n.Deps {
			collect(dep)
		}
	}
	collect(node)
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



