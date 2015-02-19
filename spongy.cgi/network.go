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
	name string
	currentLog string
	lineno int64

	basePath string
	seq int
}

type Update struct {
	Lines []string
	LastEventId string
}

func NewNetwork(basePath string) (*Network) {
	return &Network{
		basePath: basePath,
	}
}

func (nw *Network) LastEventId() string {
	return fmt.Sprintf("%s:%s:%d", nw.name, nw.currentLog, nw.lineno)
}

func (nw *Network) SetPosition(filename string, lineno int64) {
	nw.currentLog = filename
	nw.lineno = lineno
}

func (nw *Network) Tail(out chan<- *Update) error {
	if nw.currentLog == "" {
		var err error
		
		currentfn := path.Join(nw.basePath, "log", "current")
		nw.currentLog, err = os.Readlink(currentfn)
		if err != nil {
			return err
		}
	}
	
	filepath := path.Join(nw.basePath, "log", nw.currentLog)
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	
	watcher.Add(filepath)
	
	for {
		lines := make([]string, 0)
		bf := bufio.NewScanner(f)
		for bf.Scan() {
			t := bf.Text()
			nw.lineno += 1
			
			parts := strings.Split(t, " ")
			if (len(parts) >= 4) && (parts[2] == "NEXTLOG") {
				watcher.Remove(filepath)
				filename := parts[3]
				filepath = path.Join(NetworkDir, filename)
				f.Close()
				f, err = os.Open(filepath)
				if err != nil {
					return err
				}
				watcher.Add(filepath)
				nw.lineno = 0
			}
			lines = append(lines, t)
		}
		if len(lines) > 0 {
			update := Update{
				Lines: lines,
				LastEventId: nw.LastEventId(),
			}
			out <- &update
		}
		
		select {
		case _ = <-watcher.Events:
			// Somethin' happened!
		case err := <-watcher.Errors:
			return err
		}
	}
	
	return nil
}

func (nw *Network) Write(data []byte) {
	epoch := time.Now().Unix()
	pid := os.Getpid()
	filename := fmt.Sprintf("%d-%d-%d.txt", epoch, pid, nw.seq)
	
	filepath := path.Join(nw.basePath, "outq", filename)
	ioutil.WriteFile(filepath, data, 0750)
	nw.seq += 1
}
