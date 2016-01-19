package main

import (
	"bufio"
	"fmt"
	"flag"
	"log"
	"os"
	"path/filepath"
)

var playback int
var running bool = true

func inputLoop(nw *Network) {
	bf := bufio.NewScanner(os.Stdin)
	for bf.Scan() {
		line := bf.Bytes()
		nw.Write(line)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] NETDIR\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "NETDIR is the path to your IRC directory (see README)\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.IntVar(&playback, "playback", 0, "Number of lines to play back on startup")
	
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	netDir, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	nw := NewNetwork(netDir)
	defer nw.Close()
	go inputLoop(nw)

 	outq := make(chan string, 50) // to stdout
	go nw.Tail(outq)
	for line := range outq {
		fmt.Println(line)
	}
}
