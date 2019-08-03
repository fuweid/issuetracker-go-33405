// +build linux

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	dirNum    = 100
	fileNum   = 100
	deepLevel = 100

	rootDir       string
	tmpPrefixName = "fdleak"

	childCmd string

	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(&rootDir, "root", "/tmp", "root path to contain the temporary data")
	flag.StringVar(&childCmd, "childCmd", "bin/child", "child process to output content of /proc/self/fd")
}

func main() {
	flag.Parse()

	fmt.Printf("rootdir = %s\n", rootDir)
	fmt.Printf("childcmd = %s\n", childCmd)
	time.Sleep(3 * time.Second)

	startCh := make(chan struct{})
	go func() {
		close(startCh)
		for {
			root, err := generateData(rootDir)
			os.RemoveAll(root)
			if err != nil {
				panic(err)
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	<-startCh
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		for j := 0; j < 100; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cmd := exec.Command(childCmd)
				output, err := cmd.CombinedOutput()
				if err != nil {
					panic(err)
				}

				fmt.Println(string(output))
				if strings.Contains(string(output), filepath.Join(rootDir, tmpPrefixName)) {
					panic("fd leak to child process")
				}
			}()
		}
		wg.Wait()
	}
}

func generateData(rootdir string) (string, error) {
	root, err := ioutil.TempDir(rootdir, tmpPrefixName)
	if err != nil {
		return "", err
	}
	for i := 0; i < dirNum; i++ {
		dirpaths := make([]string, 0, deepLevel+1)

		dirpaths = append(dirpaths, root)
		for j := 0; j < deepLevel; j++ {
			dirpaths = append(dirpaths, randString(4))
		}

		dirpath := filepath.Join(dirpaths...)
		if err := os.MkdirAll(dirpath, 644); err != nil {
			return "", err
		}

		for i := 0; i < fileNum; i++ {
			fpath := filepath.Join(dirpath, strconv.Itoa(i))
			if err := ioutil.WriteFile(fpath, []byte(randString(i)), 644); err != nil {
				return "", err
			}
		}
	}
	return root, nil
}

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
