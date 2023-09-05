package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"FileQueueHTTP/prometheus"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func helloworld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><center><h1>FileQueueHTTP</h1>")
	fmt.Fprintf(w, "<center><a href=\"/metrics\">prometheus</a>")
}

func main() {

	pitcherStore := PitcherStore{pitcherQueue: make(chan string, 10000000)}
	go pitcherStore.run()

	http.HandleFunc("/", helloworld)
	http.HandleFunc("/getfile", pitcherStore.getFile)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type PitcherStore struct {
	pitcherQueue chan string
}

func IsHiddenFile(filename string) bool {
	return filepath.Base(filename)[0] == '.'
}

func (pitcherStore PitcherStore) shareThread(path string) {
	fmt.Println(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Share %s does not exist", path)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if IsHiddenFile(file.Name()) {
			continue
		}
		fmt.Println(file.Name())
		pitcherStore.pitcherQueue <- path + "/" + file.Name()
		prometheus.FileInQueue.Inc()
		prometheus.FileAddedToChannel.Inc()

	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) && !IsHiddenFile(event.Name) {
					pitcherStore.pitcherQueue <- event.Name
					fmt.Println("New file: ", event.Name)
					prometheus.FileInQueue.Inc()
					prometheus.FileAddedToChannel.Inc()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func (pitcherStore PitcherStore) run() {

	shares := os.Getenv("SHARES")

	if shares == "" {
		log.Printf("No shares defined")
		os.Exit(1)
	}

	shareslist := strings.Split(shares, ",")
	for _, share := range shareslist {
		go pitcherStore.shareThread(share)
	}
}

func (pitcherStore PitcherStore) getFile(w http.ResponseWriter, r *http.Request) {

	select {
	case filename, ok := <-pitcherStore.pitcherQueue:
		{
			if ok {
				w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filename))
				http.ServeFile(w, r, filename)

				// remove file
				err := os.Remove(filename)
				if err != nil {
					log.Printf("Error removing file %s", filename)
				}
				prometheus.FileInQueue.Dec()

			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	case <-time.After(50000 * time.Microsecond):
	}
}
