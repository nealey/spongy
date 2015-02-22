package main

import (
	"bufio"
	"fmt"
	"github.com/go-fsnotify/fsnotify"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type Network struct {
	running bool

	name string
	currentLog string
	lineno int64

	basePath string
	seq int
}

func NewNetwork(basePath string) (*Network) {
	return &Network{
		running: true,
		name: path.Base(basePath),
		basePath: basePath,
	}
}

func (nw *Network) Close() {
	nw.running = false
}

func (nw *Network) LastEventId() string {
	return fmt.Sprintf("%s/%s/%d", nw.name, nw.currentLog, nw.lineno)
}

func (nw *Network) SetPosition(filename string, lineno int64) {
	nw.currentLog = filename
	nw.lineno = lineno
}

func (nw *Network) errmsg(err error) []string {
	s := fmt.Sprintf("ERROR: %s", err.Error())
	return []string{s}
}

func (nw *Network) Tail(out chan<- []string) {
	if nw.currentLog == "" {
		var err error
		
		currentfn := path.Join(nw.basePath, "log", "current")
		nw.currentLog, err = os.Readlink(currentfn)
		if err != nil {
			out <- nw.errmsg(err)
			return
		}
	}
	
	filepath := path.Join(nw.basePath, "log", nw.currentLog)
	f, err := os.Open(filepath)
	if err != nil {
		out <- nw.errmsg(err)
		return
	}
	defer f.Close()
	
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		out <- nw.errmsg(err)
		return
	}
	defer watcher.Close()
	
	watcher.Add(filepath)
	lineno := int64(0)
	
	// XXX: some way to stop this?
	for nw.running {
		bf := bufio.NewScanner(f)
		lines := make([]string, 0)
		for bf.Scan() {
			lineno += 1
			if lineno <= nw.lineno {
				continue
			} else {
				nw.lineno = lineno
			}
			
			t := bf.Text()
			
			// XXX: Consider omitting PING and PONG
			parts := strings.Split(t, " ")
			if (len(parts) >= 4) && (parts[2] == "NEXTLOG") {
				watcher.Remove(filepath)
				filename := parts[3]
				filepath = path.Join(nw.basePath, "log", filename)
				f.Close()
				f, err = os.Open(filepath)
				if err != nil {
					out <- nw.errmsg(err)
					return
				}
				watcher.Add(filepath)
				lineno = 0
				nw.lineno = 0
			}
			lines = append(lines, t)
		}
		if len(lines) > 0 {
			out <- lines
		}
		
		select {
		case _ = <-watcher.Events:
			// Somethin' happened!
		case err := <-watcher.Errors:
			out <- nw.errmsg(err)
			return
		}
	}
}

func (nw *Network) Write(data []byte) {
	epoch := time.Now().Unix()
	pid := os.Getpid()
	filename := fmt.Sprintf("%d-%d-%d.txt", epoch, pid, nw.seq)
	
	filepath := path.Join(nw.basePath, "outq", filename)
	ioutil.WriteFile(filepath, data, 0750)
	nw.seq += 1
}


func Networks(basePath string) (found []*Network) {

	dir, err := os.Open(basePath)
	if err != nil {
		return
	}
	defer dir.Close()
	
	
	entities, _ := dir.Readdirnames(0)
	for _, fn := range entities {
		netdir := path.Join(basePath, fn)
		
		_, err = os.Stat(path.Join(netdir, "nick"))
		if err != nil {
			continue
		}
		
		nw := NewNetwork(netdir)
		found = append(found, nw)
	}
	
	return
}
	