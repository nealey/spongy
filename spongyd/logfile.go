package main

import (
	"fmt"
	"os"
	"log"
	"path"
	"time"
)

type Logfile struct {
	baseDir string
	file *os.File
	name string
	nlines int
	maxlines int
	outq chan string
}

func timestamp(s string) string {
	ret := fmt.Sprintf("%d %s", time.Now().Unix(), s)
	return ret
}

func NewLogfile(baseDir string, maxlines int) (*Logfile) {
	lf := Logfile{baseDir, nil, "", 0, maxlines, make(chan string, 50)}
	go lf.processQueue();
	return &lf
}

func (lf *Logfile) Close() {
	if lf.file != nil {
		lf.Log("EXIT")
		close(lf.outq)
	}
}

func (lf *Logfile) Log(s string) error {
	lf.outq <- timestamp(s)
	return nil
}

//
//

func (lf *Logfile) processQueue() {
	for line := range lf.outq {
		if (lf.file == nil) || (lf.nlines >= lf.maxlines) {
			if err := lf.rotate(); err != nil {
				// Just keep trying, I guess.
				log.Print(err)
				continue
			}
			lf.nlines = 0
		}

		if _, err := fmt.Fprintln(lf.file, line); err != nil {
			log.Print(err)
			continue
		}
		lf.nlines += 1
	}

	lf.file.Close()
}

func (lf *Logfile) rotate() error {
	fn := fmt.Sprintf("%s.log", time.Now().UTC().Format(time.RFC3339))
	pathn := path.Join(lf.baseDir, "log", fn)
	newf, err := os.OpenFile(pathn, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	
	currentPath := path.Join(lf.baseDir, "log", "current")
	
	if lf.file == nil {
		// Open "current" to append a NEXTLOG line.
		// If there's no "current", that's okay
		lf.file, _ = os.OpenFile(currentPath, os.O_WRONLY|os.O_APPEND, 0666)
	}
	
	if lf.file != nil {
		// Note location of new log
		logmsg := fmt.Sprintf("NEXTLOG %s", fn)
		fmt.Fprintln(lf.file, timestamp(logmsg))
		
		// All done with the current log
		lf.file.Close()
	}
	
	// Point to new log file
	lf.file = newf
		
	// Record symlink to new log
	os.Remove(currentPath)
	os.Symlink(fn, currentPath)
	
	logmsg := fmt.Sprintf("PREVLOG %s", lf.name)
	fmt.Fprintln(lf.file, timestamp(logmsg))
	
	lf.name = fn
	
	return nil
}
