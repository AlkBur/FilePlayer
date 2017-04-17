package main

import (
	"log"
	"net/http"
	"strings"
	"fmt"
)

const (
	cfgFile = "config.json"
)

var (
	defaultPathMP3 = ""
	debug = false
	gzip_enabled = true
)

func main() {
	cfg := LoadConfiguration(cfgFile)
	defaultPathMP3 = cfg.Database.Path
	debug = cfg.Debug
	httpAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	service := http.NewServeMux()

	log.Println("Open DB")
	db := newDB(cfg.Database.File)
	if db == nil {
		log.Fatalln("Error open DB")
	}
	log.Println("Run DB")
	go db.run()

	fs := http.FileServer(http.Dir("public"))
	service.Handle("/", fs)
	service.Handle("/ws", db)
	service.HandleFunc("/mp3/", mp3Handler(db))

	log.Println("Start web server on", httpAddr)
	log.Fatalf("Error running webserver: %v", http.ListenAndServe(httpAddr, logging(service)))
}

func mp3Handler(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, "/mp3/"); len(p) < len(r.URL.Path) {
			file, ok := mainPlayList.Get(p)
			if ok {
				http.ServeFile(w, r, file.path)
				return
			}
		}
		http.NotFound(w, r)
	}
}
