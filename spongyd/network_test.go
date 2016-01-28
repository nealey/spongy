package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func writeFile(fn string, data string) {
	ioutil.WriteFile(fn, []byte(data), os.ModePerm)
}

func createNetwork (t *testing.T, parent string) (base string, current string) {
	base, err := ioutil.TempDir(parent, "spongy-test")
	if err != nil {
		t.Fatal(err)
	}
	
	writeFile(path.Join(base, "nick"), "SpongyTest")
	writeFile(path.Join(base, "server"), "moo.slashnet.org:6697")
	writeFile(path.Join(base, "channels"), "#SpongyTest")
	os.Mkdir(path.Join(base, "outq"), os.ModePerm)
	os.Mkdir(path.Join(base, "log"), os.ModePerm)
	
	current = path.Join(base, "log", "current")
	
	return
}

func expect(t *testing.T, fpath string, needle string) {
	for i := 0; i < 8; i += 1 {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}

		fpBytes, err := ioutil.ReadFile(fpath)
		if err != nil {
			t.Log(err)
			time.Sleep(1 * time.Second)
			continue
		}
		fpString := string(fpBytes)
		if strings.Contains(fpString, needle) {
			return
		}
	}
	t.Errorf("Could not find %#v in %s", needle, fpath)
}

func TestCreateNetwork(t *testing.T) {
	base, _ := createNetwork(t, "")
	
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
	base, current := createNetwork(t, "")
	defer os.RemoveAll(base)

	n := NewNetwork(base)
	go n.Connect()
	
	time.Sleep(5 * time.Second)
	
	expect(t, current, " 001 ")
	expect(t, current, " JOIN " + n.Nick + " #SpongyTest")
	
	ioutil.WriteFile(path.Join(base, "outq", "merf"), []byte("PART #SpongyTest\n"), os.ModePerm)
	expect(t, current, " PART ")
	
	n.Close()
	return
}
