package main

import (
	"bufio"
	"fmt"
	"github.com/nealey/spongy/logfile"
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
	running bool

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
		running: true,

		basePath: basePath,

		servers: servers,
		nicks:   nicks,
		gecos:   gecoses[0],

		logq: make(chan Message, 20),
	}, err
	
	go n.LogLoop()
}

func (n *Network) Close() {
	n.conn.Close()
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
	for n.running {
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

func (n *Network) LogLoop() {
	logf := logfile.NewLogFile(int(maxlogsize))
	defer logf.Close()
	
	for m := range logq {
		logf.Log(m.String())
	}
}

func (n *Network) ServerWriteLoop() {
	for v := range n.outq {
		m, _ := Parse(v)
		n.logq <- m
		fmt.Fprintln(n.conn, v)
	}
}

func (n *Network) ServerReadLoop() {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		n.inq <- scanner.Text()
	}
	close(n.inq)
}

func (n *Network) MessageDispatch() {
	for line := n.inq {
		m, err := NewMessage(line)
		if err != nil {
			log.Print(err)
			continue
		}
		
		n.logq <- m
		// XXX: Add in a handler subprocess call
		
		switch m.Command {
		case "PING":
			n.outq <- "PONG: " + m.Text
		case "433":
			nick = nick + "_"
			outq <- fmt.Sprintf("NICK %s", nick)
		}
	}
}

func (n *Network) ConnectToServer(server string) bool {
	var err error

	switch (server[0]) {
	case '|':
		parts := strings.Split(server[1:], " ")
		n.conn, err = StartStdioProcess(parts[0], parts[1:])
	case '^':
		n.conn, err = net.Dial("tcp", server[1:])
	default:
		log.Print("Not validating server certificate!")
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		n.conn, err = tls.Dial("tcp", host, config)
	}
	
	if err != nil {
		log.Print(err)
		time.sleep(2 * time.Second)
		return false
	}
	
	return true
}
	

func (n *Network) Connect(){
	serverIndex := 0
	for n.running {
		servers, err := ReadLines(path.Join(basePath, "servers"))
		if err != nil {
			serverIndex := 0
			log.Print(err)
			time.sleep(8)
			continue
		}
		
		if serverIndex > len(servers) {
			serverIndex = 0
		}
		server := servers[serverIndex]
		serverIndex += 1
		
		if ! n.ConnectToServer(server) {
			continue
		}
		
		n.inq = make(chan string, 20)
		n.outq = make(chan string, 20)
		
		go n.ServerWriteLoop()
		go n.MessageDispatch()
		n.ServerReadLoop()
		
		close(n.outq)
	}
}

