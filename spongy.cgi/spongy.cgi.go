package main

import (
	"fmt"
	"log"
	"strings"
	"net/http"
	"net/http/cgi"
	"path"
)

type Handler struct {
	cgi.Handler
}

func (h Handler) handleCommand(cfg *Config, w http.ResponseWriter, r *http.Request) {
	network := r.FormValue("network")
	text := r.FormValue("text")
	target := r.FormValue("target")
	
	nw := NewNetwork(path.Join(cfg.BaseDir, network))
	
	var out string
	switch {
	case strings.HasPrefix(text, "/quote "):
		out = text[7:]
	case strings.HasPrefix(text, "/me "):
		out = fmt.Sprintf("PRIVMSG %s :\001ACTION %s\001", target, text[4:])
	default:
		out = fmt.Sprintf("PRIVMSG %s :%s", target, text)
	}
	nw.Write([]byte(out))

	fmt.Fprintln(w, "OK")
}

func (h Handler) handleTail(cfg *Config, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	nws := Networks(cfg.BaseDir)
	
	lastEventId := r.FormValue("HTTP_LAST_EVENT_ID")
	updates := make(chan []string, 100)
	
	for _, nw := range nws {
		nw.ReadLastEventId(lastEventId)
		go nw.Tail(updates)
		defer nw.Close()
	}
		
	for lines := range updates {
		for _, line := range lines {
			fmt.Fprintf(w, "data: %s\n", line)
		}
		
		ids := make([]string, 0)
		for _, nw := range nws {
			ids = append(ids, nw.LastEventId())
		}
		idstring := strings.Join(ids, " ")
		_, err := fmt.Fprintf(w, "id: %s\n\n", idstring)
		if err != nil {
			// Can't write anymore, guess they hung up.
			return
		}
		w.(http.Flusher).Flush()
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cfg, err := ReadConfig(h.Dir)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	
	// Validate authtok
	authtok, err := cfg.Get("authtok")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if r.FormValue("auth") != authtok {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "NO: Invalid authtok")
		return
	}
	
	// Switch based on type
	switch r.FormValue("type") {
	case "command":
		h.handleCommand(cfg, w, r)
	default:
		h.handleTail(cfg, w, r)
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

