package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

var running bool = true
var verbose bool = false
var maxlogsize uint

func debug(format string, a ...interface{}) {
	if verbose {
		log.Printf(format, a...)
	}
}

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
	
	log.Printf("Starting in %s\n", dirname)
	for running {
		dn, err := dir.Readdirnames(0); if err != nil {
			log.Fatal(err)
		}
		
		found := make(map[string]bool)
		for _, fn := range dn {
			fpath := path.Join(dirname, fn)
			if _, ok := services[fpath]; ! ok {
				if exists(path.Join(fpath, "down")) {
					continue
				}
				if ! exists(path.Join(fpath, "server")) {
					continue
				}
				
				log.Printf("Found new network %s", fpath)
				newnet := NewNetwork(fpath)
				services[fpath] = newnet
				go newnet.Connect()
			}
			found[fpath] = true
		}
		
		// If anything vanished, disconnect it
		for fpath, nw := range services {
			if _, ok := found[fpath]; ! ok {
				log.Printf("Removing vanished network %s", fpath)
				nw.Close()
			}
		}
		
		_, _ = dir.Seek(0, 0)
		time.Sleep(20 * time.Second)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] BASEPATH\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "BASEPATH is the path to your IRC directory (see README)\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.UintVar(&maxlogsize, "logsize", 6000, "Log entries before rotating")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging")
	notime := flag.Bool("notime", false, "Don't timestamp debugging messages")
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	basePath, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	if *notime {
		log.SetFlags(0)
	}
	
	runsvdir(basePath)
	
	running = false
	log.Print("Exiting for some reason!")
}
