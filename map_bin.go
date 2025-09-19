package github.com/montcao/goblib

import (
	"os/exec"
	"fmt"
	"strings"
)

// So far I think binary only relies on static analysis, just because I don't trust executing and tracing yet
// For tracing, need to setup a virtual env to trace in. container in container?

// TODO - THIS CLASS IS UNFINISHED

// look for hardcoded executable paths inside the binary
// Need to update so that if it sees an exec() , etc
func StaticBinaryCheck(path string) ([]string, error) {
	out, err := exec.Command("strings", path).Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	deps := make(map[string]struct{})
	for _, line := range lines {
		line = strings.TrimSpace(line)
		//fmt.Println(line)
		// naive pattern: any string containing /bin/ or /usr/bin/
		if strings.HasPrefix(line, "/bin/") || strings.HasPrefix(line, "/usr/bin/") {
			fmt.Println(line)
			deps[line] = struct{}{}
		}
	}
	if len(deps) == 0 {
		//fmt.Println("empty")
		return nil, nil
	}
	result := []string{}
	for dep := range deps {
		result = append(result, dep)
	}
	return result, nil
}