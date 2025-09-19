package github.com/montcao/goblib

// Model for Tree Node objects, this is the graph
type Node struct {
	Path string  `json:"path"`
	Deps []*Node `json:"deps,omitempty"`
	Metadata *BinaryMetadata
}

type Forest []*Node // multiple bin support TODO

//

// Model for metadata information from a binary 
type Metadata interface {

}

type BinaryMetadata struct {
	Architecture     string   // e.g., "x86-64", "ARM64"
	DynamicLoader    string   // e.g., "/lib64/ld-linux-x86-64.so.2"
	SharedLibraries  []string // e.g., ["libpthread.so.0", "libdl.so.2", "libm.so.6", "libc.so.6"]
	RuntimeBinaries  []string // e.g., ["tar", "unzip", "sh"] - TODO maybe put in path instead
	AbsPath          string // e.g the binary location /usr/bin/ls
	Package          string // e.g the package where the library can be installed
}


