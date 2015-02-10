package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

var running bool = true
var maxlogsize uint

func exists(filename string) bool {
	_, err := os.Stat(filename); if err != nil {
		return false
	}
	return true
}

func runsvdir(dirname string) {
	services := make(map[string]*Network)
	
	dir, err := os.Open(dirname)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()
	
	for running {
		dn, err := dir.Readdirnames(0); if err != nil {
			log.Fatal(err)
		}
		
		found := make(map[string]bool)
		for _, fn := range dn {
			fpath := path.Join(dirname, fn)
			if exists(path.Join(fpath, "down")) {
				continue
			}
			if _, ok := services[fpath]; ! ok {
				if ! exists(path.Join(fpath, "servers")) {
					continue
				}
				
				newnet := NewNetwork(fpath)
				services[fpath] = newnet
				go newnet.Connect()
			}
			found[fpath] = true
		}
		
		// If anything vanished, disconnect it
		for fpath, nw := range services {
			if _, ok := found[fpath]; ! ok {
				nw.Close()
			}
		}
		
		_, _ = dir.Seek(0, 0)
		time.Sleep(20 * time.Second)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] HOST:PORT\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.UintVar(&maxlogsize, "logsize", 1000, "Log entries before rotating")
	flag.Parse()
	running = false
}
