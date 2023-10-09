package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"FileQueueHTTP/prometheus"

	"github.com/shirou/gopsutil/disk"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func helloworld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><title>FileQueueHTTP(v.RELEASE_DATE)</title><center><h1 style=\"color: red;\">FileQueueHTTP(v.RELEASE_DATE)</h1>")
	fmt.Fprintf(w, "<center><a href=\"/metrics\">Prometheus</a>")
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

var ctx = context.Background()

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

			usageStat, err := disk.UsageWithContext(ctx, path)
			if err != nil {
				fmt.Println("Panic! Disk usage not working err:", err)
			}

			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) && !IsHiddenFile(event.Name) {

					if usageStat.UsedPercent > DiskUsageAllowed {
						fmt.Println("Panic! UsedPercent:", usageStat.UsedPercent, " > DiskUsageAllowed:", DiskUsageAllowed, " Auto removing file:", event.Name)
						err = os.Remove(event.Name)
						if err != nil {
							log.Printf("Panic! Error removing file %s", event.Name)
						}
						prometheus.FilesPruned.Inc()
						continue
					}

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

	for {
		select {
		case filename, ok := <-pitcherStore.pitcherQueue:
			{
				if ok {
					// check if file exists
					if _, err := os.Stat(filename); os.IsNotExist(err) {
						fmt.Println("File does not exist. Removing from queue.", filename)
						prometheus.FilesDoNotExist.Inc()
						continue

					}
					w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filename))
					http.ServeFile(w, r, filename)

					// remove file
					err := os.Remove(filename)
					if err != nil {
						log.Printf("Error removing file %s", filename)
					}
					prometheus.FileInQueue.Dec()
					return

				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}
		case <-time.After(50000 * time.Microsecond):
		}
	}
}

func getDiskUsageAllowed() float64 {
	allowedDiskEnv, err := strconv.ParseFloat(os.Getenv("DISK_USAGE_ALLOWED"), 64)
	if err != nil {
		allowedDiskEnv = 75
	}
	if allowedDiskEnv < 5 || allowedDiskEnv > 99 {
		fmt.Println("Panic! DISK_USAGE_ALLOWED not accepted. Only 5-99 is allowed")
		allowedDiskEnv = 75
	}
	fmt.Println("DISK_USAGE_ALLOWED:", allowedDiskEnv, "%")
	return allowedDiskEnv
}

var DiskUsageAllowed = getDiskUsageAllowed()
