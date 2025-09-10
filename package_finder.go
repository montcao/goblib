package goblib

import (
	"os/exec"
	"strings"
	"fmt"
	"os"
)

type PackageFinder interface {
	FindPackage(lib string) (string, error)
}


type DebianFinder struct{}

func (d DebianFinder) FindPackage(lib string) (string, error) {
	// Try dpkg -S first
	out, err := exec.Command("dpkg", "-S", lib).CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	// Fallback: apt-file search
	out, err = exec.Command("apt-file", "search", lib).CombinedOutput()
	if err == nil && len(out) > 0 {
		return strings.SplitN(strings.TrimSpace(string(out)), ":", 2)[0], nil
	}

	return "", fmt.Errorf("package not found for %s", lib)
}

type FedoraFinder struct{}

func (r FedoraFinder) FindPackage(lib string) (string, error) {
	// rpm -qf
	out, err := exec.Command("rpm", "-qf", lib).CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	// fallback: dnf provides
	out, err = exec.Command("dnf", "provides", lib).CombinedOutput()
	if err == nil && len(out) > 0 {
		return strings.SplitN(strings.TrimSpace(string(out)), ":", 2)[0], nil
	}
	return "", fmt.Errorf("package not found for %s", lib)
}

type ApkFinder struct{}

func (a ApkFinder) FindPackage(lib string) (string, error) {
	// apk info -W
	out, err := exec.Command("apk", "info", "-W", lib).CombinedOutput()
	if err == nil && len(out) > 0 {
		return strings.TrimSpace(string(out)), nil
	}
	return "", fmt.Errorf("package not found for %s", lib)
}

func DetectFinder() PackageFinder {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return nil
	}
	s := string(data)
	switch {
	case strings.Contains(s, "Ubuntu"), strings.Contains(s, "Debian"):
		return DebianFinder{}
	case strings.Contains(s, "Fedora"), strings.Contains(s, "Red Hat"), strings.Contains(s, "CentOS"):
		return FedoraFinder{}
	case strings.Contains(s, "Alpine"):
		return ApkFinder{}
	default:
		return nil
	}
}