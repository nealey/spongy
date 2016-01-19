package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func setupRunSpongy(t *testing.T, parent string, baseChan chan<- string) {
	base, _ := createNetwork(t, parent)
	baseChan <- base
	close(baseChan)
	runsvdir(parent)
	os.RemoveAll(base)
}

func TestRunsvdir(t *testing.T) {
	parent, err := ioutil.TempDir("", "spongy-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(parent)

	baseChan := make(chan string)
	go setupRunSpongy(t, parent, baseChan)
	
	base := <- baseChan
	expect(t, path.Join(base, "log", "current"), " 001 ")
	os.RemoveAll(base)
}
