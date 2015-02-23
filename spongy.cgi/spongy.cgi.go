package main

import (
	"fmt"
	"github.com/go-fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"os"
	"bufio"
	"strconv"
	"strings"
	"net/http"
	"net/http/cgi"
	"time"
	"path"
)

type Handler struct {
	cgi.Handler
}

var NetworkDir string

func ReadString(fn string) string {
	octets, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(octets))
}

func tail(w http.ResponseWriter, filename string, pos int64) {
	var err error

	currentfn := path.Join(NetworkDir, "current")
	if filename == "" {
		filename, err = os.Readlink(currentfn)
		if err != nil {
			log.Fatal(err)
		}
	}
	
	filepath := path.Join(NetworkDir, filename)

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	watcher.Add(filepath)
	
	for {
		printid := false
		
		newpos, err := f.Seek(pos, 0)
		if err != nil {
			log.Fatal(err)
		}
		
		if newpos != pos {
			log.Fatal("Lost my position in the log, somehow (log truncated?)")
		}
		
		bf := bufio.NewScanner(f)
		for bf.Scan() {
			t := bf.Text()
			pos += int64(len(t)) + 1 // XXX: this breaks if we ever see \r\n
			
			parts := strings.Split(t, " ")
			if (len(parts) >= 4) && (parts[2] == "NEXTLOG") {
				watcher.Remove(filepath)
				filename = parts[3]
				filepath = path.Join(NetworkDir, filename)
				f.Close()
				f, err = os.Open(filepath)
				if err != nil {
					log.Fatal(err)
				}
				watcher.Add(filepath)
				pos = 0
			}
			fmt.Fprintf(w, "data: %s\n", t)
			printid = true
		}
		if printid {
			_, err = fmt.Fprintf(w, "id: %s/%d\n\n", filename, pos)
		}
		if err != nil {
			break
		}
		w.(http.Flusher).Flush()
		
		select {
		case _ = <-watcher.Events:
			// Somethin' happened!
		case err := <-watcher.Errors:
			log.Fatal(err)
		}
	}
}

func handleCommand(w http.ResponseWriter, text string, target string) {
	fn := path.Join(NetworkDir, fmt.Sprintf("outq/cgi.%d", time.Now().Unix()))
	f, err := os.Create(fn)
	if err != nil {
		fmt.Fprintln(w, "NO: Cannot create outq file")
		fmt.Fprintln(w, err)
		return
	}
	defer f.Close()
	
	switch {
	case strings.HasPrefix(text, "/quote "):
		fmt.Fprintln(f, text[7:])
	case strings.HasPrefix(text, "/me "):
		fmt.Fprintf(f, "PRIVMSG %s :\001ACTION %s\001\n", target, text[4:])
	default:
		fmt.Fprintf(f, "PRIVMSG %s :%s\n", target, text)
	}

	fmt.Fprintln(w, "OK")
}


func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	BaseDir := "networks"
	DefaultDir := path.Join(BaseDir, "default")
	NetworkDir = path.Join(BaseDir, r.FormValue("network"))
	
	if path.Dir(DefaultDir) != path.Dir(NetworkDir) {
		NetworkDir = DefaultDir
	}
	
	authtok := ReadString(path.Join(NetworkDir, "authtok"))
	if r.FormValue("auth") != authtok {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "NO: Invalid authtok")
		return
	}
	switch r.FormValue("type") {
	case "command":
		w.Header().Set("Content-Type", "text/plain")
		handleCommand(w, r.Form.Get("text"), r.FormValue("target"))
	default:
		w.Header().Set("Content-Type", "text/event-stream")
		parts := strings.Split(os.Getenv("HTTP_LAST_EVENT_ID"), "/")
		if len(parts) == 2 {
			filename := path.Base(parts[0])
			pos, _ := strconv.ParseInt(parts[1], 0, 64)
			tail(w, filename, pos)
		} else {
			tail(w, "", 0)
		}
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	log.SetPrefix("Status: 500 CGI Go Boom\nContent-type: text/plain\n\nERROR: ")
	h := Handler{}
	if err := cgi.Serve(h); err != nil {
		log.Fatal(err)
	}
}

