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
	w.Header().Set("Content-Type", "text/plain")
	nws := Networks(cfg.BaseDir)
	for _, nw := range nws {
		fmt.Fprintf(w, "%v\n", nw)
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
	h := Handler{}
	if err := cgi.Serve(h); err != nil {
		log.Fatal(err)
	}
}

