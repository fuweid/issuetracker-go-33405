// +build linux

package main

import (
	"fmt"
	"os/exec"
)

func main() {
	// print the file descriptors for self
	cmd := exec.Command("ls", "-al", "/proc/self/fd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}
