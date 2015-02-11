package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func writeFile(fn string, data string) {
	ioutil.WriteFile(fn, []byte(data), os.ModePerm)
}

func createNetwork(t *testing.T) (base string) {
	base, err := ioutil.TempDir("", "spongy-test")
	if err != nil {
		t.Fatal(err)
	}
	
	writeFile(path.Join(base, "nick"), "spongy_test")
	writeFile(path.Join(base, "server"), "moo.slashnet.org:6697")
	os.Mkdir(path.Join(base, "outq"), os.ModePerm)
	os.Mkdir(path.Join(base, "log"), os.ModePerm)
	
	return
}

func TestCreateNetwork(t *testing.T) {
	base := createNetwork(t)
	
	if fi, err := os.Stat(path.Join(base, "nick")); err != nil {
		t.Error(err)
	} else if fi.IsDir() {
		t.Error("%s is not a regular file", path.Join(base, "nick"))
	}
	
	os.RemoveAll(base)
	if _, err := os.Stat(path.Join(base, "outq")); err == nil {
		t.Error("Didn't unlink outq")
	}
}

func TestConnect(t *testing.T) {
	base := createNetwork(t)
	defer os.RemoveAll(base)

	n := NewNetwork(base)
	go n.Connect()
	
	time.Sleep(5 * time.Second)
	
	logBytes, err := ioutil.ReadFile(path.Join(base, "log", "current"))
	if err != nil {
		n.Close()
		t.Fatal(err)
	}
	t.Log("logBytes: ", logBytes)
	
	n.Close()
	return
}
