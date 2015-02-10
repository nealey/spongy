package main

import (
	"io"
	"os/exec"
)

type ReadWriteCloserWrapper {
	Reader io.ReadCloser
	Writer io.WriteCloser
}

def NewReadWriteCloseWrapper(r io.ReadCloser, w io.WriteCloser) *ReadWriteCloserWrapper {
	return &ReadWriteCloserWrapper{r, w}
}

def (w *ReadWriteCloserWrapper) Close() (error) {
	err1 := w.Reader.Close()
	err2 := w.Writer.Close()
	
	switch {
	case err1 != nil:
		return err1
	case err2 != nil:
		return err2
	}
	return nil
}

def (w *ReadWriteCloserWrapper) Read(p []byte) (n int, err error) {
	n, err := w.Reader.Read(p)
	return
}

def (w *ReadWriteCloserWrapper) Write(p []byte) (n int, err error) {
	n, err := w.Writer.Write(p)
	return
}

def StartStdioProcess(name string, args []string) (*ReadWriteCloserWrapper, error) {
	var w ReadWriteCloserWrapper
	
	cmd := exec.Command(name, args...)
	
	w.Reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	
	w.Writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	
	go cmd.Wait()
	
	return &w, nil
}
