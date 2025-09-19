# goblib (beta)


<div align=center>
<img width="333" height="333" alt="goblib_mascot" src="https://github.com/user-attachments/assets/000fce85-c306-4a30-a736-008d3f6c5f32" />
</div>
goblib is a safe binary dependency analysis toolkit in Go. 

It can be used in other Go programs to take in a binary and output the dependencies that the binary relies on, including the dependencies of its dependencies. In this way, a user can see full visibility of the binaries supply chain.

goblib uses Go's built in debug/elf package to parse ELF headers, so there's no chance of accidental binary execution like ldd (which was the initial starting point on how this project started....).

### Why is this needed?

There was once a NVDIA CUDA container that my buddy, notthatguy, and I were deploying to run for hashcracking. We were using the GPU's for a single binary (hashcat) and the container was overkill at [8GB](https://hub.docker.com/r/nvidia/cuda). To narrow it down, notthatguy used ldd to trace the system calls and get everything we needed to run on a minimal CUDA image. This was the [result](https://hub.docker.com/r/cerog/hashtopolis-nvidia-agent-lite12.0).

8GB -> 300MB ain't too bad. 

This library is meant to help with tasks like that, and as a helper to a distroless tool we're currently building in stealth mode (coming soon!).


## Installation

TODO

## Usage

goblib builds a tree from a given binary path. In the initial beta it has only been tested with `/bin/ls`, `unzip`, `hashcat`, and `vim-tiny`.

### Example usage

The `main()` example below takes in a binary path as the argument.

```
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <binary>\n", os.Args[0])
		os.Exit(2)
	}

	rootPath := flag.Args()[0]

	tree := BuildTree(rootPath, EmptyTree())

	fmt.Println("\n**********PRINT TREE************\n")
	PrintFullTree(tree, "", map[string]bool{})

	fmt.Println("\n**********PRINT UNIQUE LIB DEPENDENCIES **************\n")
	lib_dependencies, err := tree.GetUniqueDependencies()
	if err != nil {
		// do nothing
	}
	fmt.Println(lib_dependencies)

	fmt.Println("\n**********GET NODES************\n")
	nodes := tree.GetNodes()
	fmt.Println(nodes)

	fmt.Println("\n**********GET INDIVIDUAL NODES************\n")
	for _, node := range nodes {
		if node.Metadata != nil {
			fmt.Printf("Path: %s\n", node.Path)
			fmt.Printf("Architecture: %s\n", node.Metadata.Architecture)
			fmt.Printf("DynamicLoader: %s\n", node.Metadata.DynamicLoader)
			fmt.Printf("SharedLibraries: %v\n", node.Metadata.SharedLibraries)
			fmt.Printf("RuntimeBinaries: %v\n", node.Metadata.RuntimeBinaries)
			fmt.Printf("Package: %v\n", node.Metadata.Package)
			fmt.Println()
		}
	}

	fmt.Println("\n**********GET DYNAMIC LOADER LIBS************\n")
	dls := tree.GetDynamicLoaders()
	fmt.Println(dls)

	fmt.Println("\n**********STATIC BINARY CHECK************\n")
	bin_found, _ := StaticBinaryCheck(rootPath)
	if bin_found != nil {
		print(bin_found)
	}
	fmt.Println("\n*********************\n")
	
	df := DebianFinder{}
	tree.UpdateMetaWithPackage(df)
	fmt.Println("\n**********GET INDIVIDUAL NODES AFTER FINDING PACKAGES************\n")
	for _, node := range nodes {
		if node.Metadata != nil {
			fmt.Printf("Path: %s\n", node.Path)
			fmt.Printf("Architecture: %s\n", node.Metadata.Architecture)
			fmt.Printf("DynamicLoader: %s\n", node.Metadata.DynamicLoader)
			fmt.Printf("SharedLibraries: %v\n", node.Metadata.SharedLibraries)
			fmt.Printf("RuntimeBinaries: %v\n", node.Metadata.RuntimeBinaries)
			fmt.Printf("Package: %v\n", node.Metadata.Package)
			fmt.Println()
		}
	}
}
```
What that looks like:

<img width="582" height="450" alt="image" src="https://github.com/user-attachments/assets/4dc45bf6-7ad4-4d32-ba19-94dd4e7dbece" />




## Contributions
Always welcome 

## TODO 
- Test with other binaries, i.e hashcat []
- Link trees so that a list of binaries produces a forest of all dependencies needed by that list []
- Error testing and edge cases []
