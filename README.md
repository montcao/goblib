# goblib (beta)

<div align=center>
<img width="333" height="333" alt="goblib_mascot" src="https://github.com/user-attachments/assets/000fce85-c306-4a30-a736-008d3f6c5f32" />
</div>
goblib is a safe binary dependency analysis toolkit in Go. It can be used in other Go programs to take in a binary and output the dependencies that the binary relies on, including the dependencies of its dependencies. In this way, a user can see full visibility of the binaries supply chain.

goblib uses Go's built in debug/elf package to parse ELF headers, never executing the binary like ldd (which was the initial starting point).

### Why is this needed?

There was once a NVDIA CUDA container that my buddy, notthatguy, and I were deploying to run for hashcracking. We were using the GPU's for a single binary (hashcat) and the container was overkill at [8GB](https://hub.docker.com/r/nvidia/cuda). To narrow it down, notthatguy used ldd to trace the system calls and get everything we needed to run on a minimal CUDA image. This was the [result](https://hub.docker.com/r/cerog/hashtopolis-nvidia-agent-lite12.0).

8GB -> 300MB ain't too bad. 

This library is meant to help with tasks like that, and as a helper to a distroless tool we're currently building in stealth mode (coming soon!).


## Installation

TODO

## Usage

goblib builds a tree from a given binary path. In the initial beta it has only been tested with `/bin/ls`.

### Example usage

The `main()` example below takes in the `/bin/ls` as an argument, but you can pass it in the binary name however you wish. 
```
func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <binary>\n", os.Args[0])
		os.Exit(2)
	}

	rootPath := flag.Args()[0]

	tree := goblib.BuildTree(rootPath, EmptyTree())
	goblib.PrintFullTree(tree, "", map[string]bool{})
    dependencies, err = tree.GetUniqueDependencies()
	fmt.Println(dependencies)
}
```
What that looks like:

<img width="821" height="317" alt="image" src="https://github.com/user-attachments/assets/dbde7338-8304-4338-88bd-ac8d865b5bd4" />



## Contributions
Always welcome 

## TODO 
- Test with other binaries, i.e hashcat []
- Link trees so that a list of binaries produces a forest of all dependencies needed by that list []
- Error testing and edge cases []
