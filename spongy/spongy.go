package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/nealey/spongy/logfile"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var running bool = true
var nick string
var gecos string
var maxlogsize uint
var logq chan Message

func isChannel(s string) bool {
	if s == "" {
		return false
	}

	switch s[0] {
	case '#', '&', '!', '+', '.', '-':
		return true
	default:
		return false
	}
}

func (m Message) String() string {
	args := strings.Join(m.Args, " ")
	return fmt.Sprintf("%s %s %s %s %s :%s", m.FullSender, m.Command, m.Sender, m.Forum, args, m.Text)
}

func logLoop() {
	logf := logfile.NewLogfile(int(maxlogsize))
	defer logf.Close()

	for m := range logq {
		logf.Log(m.String())
	}
}

func nuhost(s string) (string, string, string) {
	var parts []string

	parts = strings.SplitN(s, "!", 2)
	if len(parts) == 1 {
		return s, "", ""
	}
	n := parts[0]
	parts = strings.SplitN(parts[1], "@", 2)
	if len(parts) == 1 {
		return s, "", ""
	}
	return n, parts[0], parts[1]
}


func dispatch(outq chan<- string, m Message) {
	logq <- m
	switch m.Command {
	case "PING":
		outq <- "PONG :" + m.Text
	case "433":
		nick = nick + "_"
		outq <- fmt.Sprintf("NICK %s", nick)
	}
}

func handleInfile(path string, outq chan<- string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	os.Remove(path)
	inf := bufio.NewScanner(f)
	for inf.Scan() {
		txt := inf.Text()
		outq <- txt
	}
}

func monitorDirectory(dirname string, dir *os.File, outq chan<- string) {
	latest := time.Unix(0, 0)
	for running {
		fi, err := dir.Stat()
		if err != nil {
			break
		}
		current := fi.ModTime()
		if current.After(latest) {
			latest = current
			dn, _ := dir.Readdirnames(0)
			for _, fn := range dn {
				path := dirname + string(os.PathSeparator) + fn
				handleInfile(path, outq)
			}
			_, _ = dir.Seek(0, 0)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] HOST:PORT\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	dotls := flag.Bool("notls", true, "Disable TLS security")
	outqdir := flag.String("outq", "outq", "Output queue directory")
	flag.UintVar(&maxlogsize, "logsize", 1000, "Log entries before rotating")
	flag.StringVar(&gecos, "gecos", "Bob The Merry Slug", "Gecos entry (full name)")

	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "Error: must specify nickname and host")
		os.Exit(69)
	}

	dir, err := os.Open(*outqdir)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	nick := flag.Arg(0)
	host := flag.Arg(1)

	conn, err := connect(host, *dotls)
	if err != nil {
		log.Fatal(err)
	}

	inq := make(chan string)
	outq := make(chan string)
	logq = make(chan Message)
	go logLoop()
	go readLoop(conn, inq)
	go writeLoop(conn, outq)
	go monitorDirectory(*outqdir, dir, outq)

	outq <- fmt.Sprintf("NICK %s", nick)
	outq <- fmt.Sprintf("USER %s %s %s: %s", nick, nick, nick, gecos)
	for v := range inq {
		p, err := Parse(v)
		if err != nil {
			continue
		}
		dispatch(outq, p)
	}

	running = false

	close(outq)
	close(logq)
}
