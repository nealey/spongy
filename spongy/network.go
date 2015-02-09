package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

// This gets called a lot.
// So it's easy to fix stuff while running.

func ReadLines(fn string) ([]string, error) {
	lines := make([]string, 0)

	f, err := os.Open(fn)
	if err != nil {
		return lines, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "":
		case line[0] == '#':
		default:
			lines = append(lines, line)
		}
	}

	return lines, nil
}

type Network struct {
	basePath string

	conn io.ReadWriteCloser
	logq chan Message
	inq  chan string
	outq chan string
}

func NewNetwork(basePath string) (*Network, error) {
	nicks, err := ReadLines(path.Join(basePath, "nicks"))
	if err != nil {
		return nil, err
	}

	gecoses, err := ReadLines(path.Join(basePath, "gecos"))
	if err != nil {
		return nil, err
	}

	return &Network{
		basePath: basePath,

		servers: servers,
		nicks:   nicks,
		gecos:   gecoses[0],

		logq: make(chan Message),
		inq:  make(chan string),
		outq: make(chan string),
	}, err
}

func (n *Network) Close() {
	close(n.logq)
	close(n.inq)
	close(n.outq)
}

func (n *Network) WatchOutqDirectory() {
	outqDirname := path.Join(n.basePath, "outq")

	dir, err := os.Open(outqDirname)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()
	
	// XXX: Do this with fsnotify
	for running {
		entities, _ := dir.Readdirnames(0)
		for _, fn := range entities {
			pathname := path.Join(outqDirname, fn)
			n.HandleInfile(pathname)
		}
		_, _ = dir.Seek(0, 0)
		time.Sleep(500 * time.Millisecond)
	}
}

func (n *Network) HandleInfile(fn string) {
	f, err := os.Open(fn)
	if err != nil {
		return
	}
	defer f.Close()
	
	// Do this after Open attempt.
	// If Open fails, the file will stick around.
	// Hopefully this is helpful for debugging.
	os.Remove(fn)

	inf := bufio.NewScanner(f)
	for inf.Scan() {
		txt := inf.Text()
		n.outq <- txt
	}
}

func (n *Network) WriteLoop() {
	for v := range n.outq {
		m, _ := Parse(v)
		n.logq <- m
		fmt.Fprintln(n.conn, v)
	}
}

func (n *Network) ReadLoop() {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		n.inq <- scanner.Text()
	}
	close(n.inq)
}

func 

func (n *Network) Connect(){
	serverIndex := 0
	for running {
		servers, err := ReadLines(path.Join(basePath, "servers"))
		if err != nil {
			log.Print(err)
			serverIndex := 0
			time.sleep(2 * time.Second)
			continue
		}
		
		if serverIndex > len(servers) {
			serverIndex = 0
		}
		server := servers[serverIndex]
		
		switch (server[0]) {
		case '|':
			
		
	if dotls {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		return tls.Dial("tcp", host, config)
	} else {
		return net.Dial("tcp", host)
	}
}
