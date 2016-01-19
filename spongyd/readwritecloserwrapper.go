package main

import (
	"io"
	"os/exec"
	"log"
)

type ReadWriteCloserWrapper struct {
	Reader io.ReadCloser
	Writer io.WriteCloser
	cmd *exec.Cmd
}

func NewReadWriteCloseWrapper(r io.ReadCloser, w io.WriteCloser) *ReadWriteCloserWrapper {
	return &ReadWriteCloserWrapper{r, w, nil}
}

func (w *ReadWriteCloserWrapper) Close() (error) {
	err1 := w.Reader.Close()
	err2 := w.Writer.Close()
	if w.cmd != nil{
		w.cmd.Wait()
	}
	
	switch {
	case err1 != nil:
		return err1
	case err2 != nil:
		return err2
	}
	return nil
}

func (w *ReadWriteCloserWrapper) Read(p []byte) (n int, err error) {
	n, err = w.Reader.Read(p)
	return
}

func (w *ReadWriteCloserWrapper) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	return
}

func StartStdioProcess(name string, args []string) (w *ReadWriteCloserWrapper, err error) {
	w = new(ReadWriteCloserWrapper)

	cmd := exec.Command(name, args...)
	
	if cmd == nil {
		log.Fatalf("Can't run command: %v %v", name, args)
	}
	
	w.Reader, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	
	w.Writer, err = cmd.StdinPipe()
	if err != nil {
		w.Reader.Close()
		return nil, err
	}
	
	if err = cmd.Start(); err != nil {
		w.Reader.Close()
		w.Writer.Close()
		return nil, err
	}
	
	w.cmd = cmd
	
	return 
}
