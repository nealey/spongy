package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/nealey/spongy/logfile"
	"io"
	"log"
	"net"
	"os"
	"os/user"
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

	Nick string
	
	basePath string

	conn io.ReadWriteCloser
	logq chan Message
	inq  chan string
	outq chan string
}

func NewNetwork(basePath string) *Network {
	nw := Network{
		running: true,
		basePath: basePath,
		logq: make(chan Message, 20),
	}
	
	go nw.LogLoop()
	
	return &nw
}

func (nw *Network) Close() {
	nw.conn.Close()
	close(nw.logq)
	close(nw.inq)
	close(nw.outq)
}

func (nw *Network) WatchOutqDirectory() {
	outqDirname := path.Join(nw.basePath, "outq")

	dir, err := os.Open(outqDirname)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()
	
	// XXX: Do this with fsnotify
	for nw.running {
		entities, _ := dir.Readdirnames(0)
		for _, fn := range entities {
			pathname := path.Join(outqDirname, fn)
			nw.HandleInfile(pathname)
		}
		_, _ = dir.Seek(0, 0)
		time.Sleep(500 * time.Millisecond)
	}
}

func (nw *Network) HandleInfile(fn string) {
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
		nw.outq <- txt
	}
}

func (nw *Network) LogLoop() {
	logf := logfile.NewLogfile(int(maxlogsize))
	defer logf.Close()
	
	for m := range nw.logq {
		logf.Log(m.String())
	}
}

func (nw *Network) ServerWriteLoop() {
	for v := range nw.outq {
		m, _ := NewMessage(v)
		nw.logq <- m
		fmt.Fprintln(nw.conn, v)
	}
}

func (nw *Network) ServerReadLoop() {
	scanner := bufio.NewScanner(nw.conn)
	for scanner.Scan() {
		nw.inq <- scanner.Text()
	}
	close(nw.inq)
}

func (nw *Network) NextNick() {
	nicks, err := ReadLines(path.Join(nw.basePath, "nick"))
	if err != nil {
		log.Print(err)
		return
	}
	
	// Make up some alternates if they weren't provided
	if len(nicks) == 1 {
		nicks = append(nicks, nicks[0] + "_")
		nicks = append(nicks, nicks[0] + "__")
		nicks = append(nicks, nicks[0] + "___")
	}
	
	nextidx := 0
	for idx, n := range nicks {
		if n == nw.Nick {
			nextidx = idx + 1
		}
	}
	
	nw.Nick = nicks[nextidx % len(nicks)]
	nw.outq <- "NICK " + nw.Nick
}

func (nw *Network) JoinChannels() {
	chans, err := ReadLines(path.Join(nw.basePath, "channels"))
	if err != nil {
		log.Print(err)
		return
	}
	
	for _, ch := range chans {
		nw.outq <- "JOIN " + ch
	}
}

func (nw *Network) MessageDispatch() {
	for line := range nw.inq {
		m, err := NewMessage(line)
		if err != nil {
			log.Print(err)
			continue
		}
		
		nw.logq <- m
		// XXX: Add in a handler subprocess call
		
		switch m.Command {
		case "PING":
			nw.outq <- "PONG: " + m.Text
		case "001":
			nw.JoinChannels()
		case "433":
			nw.NextNick()
		}
	}
}

func (nw *Network) ConnectToServer(server string) bool {
	var err error
	var name string

	names, err := ReadLines(path.Join(nw.basePath, "name"))
	if err != nil {
		me, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		name = me.Name
	} else {
		name = names[0]
	}

	switch (server[0]) {
	case '|':
		parts := strings.Split(server[1:], " ")
		nw.conn, err = StartStdioProcess(parts[0], parts[1:])
	case '^':
		nw.conn, err = net.Dial("tcp", server[1:])
	default:
		log.Print("Not validating server certificate!")
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		nw.conn, err = tls.Dial("tcp", server, config)
	}
	
	if err != nil {
		log.Print(err)
		time.Sleep(2 * time.Second)
		return false
	}
	
	fmt.Fprintf(nw.conn, "USER g g g :%s\n", name)
	nw.NextNick()
	
	return true
}
	

func (nw *Network) Connect(){
	serverIndex := 0
	for nw.running {
		servers, err := ReadLines(path.Join(nw.basePath, "servers"))
		if err != nil {
			serverIndex = 0
			log.Print(err)
			time.Sleep(8)
			continue
		}
		
		if serverIndex > len(servers) {
			serverIndex = 0
		}
		server := servers[serverIndex]
		serverIndex += 1
		
		if ! nw.ConnectToServer(server) {
			continue
		}
		
		nw.inq = make(chan string, 20)
		nw.outq = make(chan string, 20)
		
		go nw.ServerWriteLoop()
		go nw.MessageDispatch()
		nw.ServerReadLoop()
		
		close(nw.outq)
	}
}

