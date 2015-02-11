package main

import (
	"os"
	"path"
	"testing"
)

func setupRunSpongy(t *testing.T, baseChan chan<- string) {
	base, _ := createNetwork(t)
	baseChan <- base
	close(baseChan)
	runsvdir(base)
	os.RemoveAll(base)
}

func TestRunsvdir(t *testing.T) {
	baseChan := make(chan string)
	go setupRunSpongy(t, baseChan)
	base := <- baseChan
	expect(t, path.Join(base, "log", "current"), " 001 ")
	os.RemoveAll(base)
}
