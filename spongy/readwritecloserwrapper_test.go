package main

import (
	"bytes"
	"testing"
)

func TestRWCWCat(t *testing.T) {
	proc, err := StartStdioProcess("cat", []string{})
	if err != nil {
		t.Error(err)
	}
	
	out := []byte("Hello, World\n")
	p := make([]byte, 0, 50)
	
	n, err := proc.Write(out)
	if err != nil {
		t.Error(err)
	}
	if n != len(out) {
		t.Errorf("Wrong number of bytes in Write: wanted %d, got %d", len(out), n)
	}
	
	n, err = proc.Read(p)
	if err != nil {
		t.Error(err)
	}
	if n != len(out) {
		t.Errorf("Wrong number of bytes in Read: wanted %d, got %d", len(out), n)
	}
	if 0 != bytes.Compare(p, out) {
		t.Errorf("Mangled read")
	}
	
	proc.Close()
}