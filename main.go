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

	"github.com/fsnotify/fsnotify"
)

func init() {
	//   http.HandleFunc("/", handler)
	// get env variable
	shares := os.Getenv("SHARES")

	if shares == "" {
		log.Printf("No shares defined")
		os.Exit(1)
	}

	if shares != "" {
		//split shares by comma
		shareslist := strings.Split(shares, ",")
		//loop through shares
		for _, share := range shareslist {
			// check share esists
			if _, err := os.Stat(share); os.IsNotExist(err) {
				log.Printf("Share %s does not exist", share)
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {

	pitcherStore := PitcherStore{pitcherQueue: make(chan string, 10000000)}
	go pitcherStore.run()

	http.HandleFunc("/", pitcherStore.getFile)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type PitcherStore struct {
	counters     map[string]int
	pitcherQueue chan string
}

func IsHiddenFile(filename string) bool {
	return filename[0] == '.'
}

func (pitcherStore PitcherStore) shareThread(path string) {
	fmt.Println(path)
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
				log.Println("event:", event)
				if event.Has(fsnotify.Rename) {
					log.Println("Renamed file:", event.Name)
					pitcherStore.pitcherQueue <- event.Name
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
				// remove path from filename

				// set header
				w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filename))
				// serve file
				http.ServeFile(w, r, filename)

			} else {
				log.Fatalf("Panic! ok was not true")
			}
		}
	case <-time.After(50000 * time.Microsecond):
	}
}
