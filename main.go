package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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

	pitcherStore := PitcherStore{pitcherQueue: make(chan string, 1000)}
	go pitcherStore.run()

	http.HandleFunc("/", pitcherStore.getFile)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type PitcherStore struct {
	counters     map[string]int
	pitcherQueue chan string
}

func (pitcherStore PitcherStore) run() {

	shares := os.Getenv("SHARES")
	shareslist := strings.Split(shares, ",")
	for _, share := range shareslist {
		files, err := ioutil.ReadDir(share)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			fmt.Println(file.Name())
			pitcherStore.pitcherQueue <- file.Name()
		}
	}
}

func (pitcherStore PitcherStore) getFile(w http.ResponseWriter, r *http.Request) {

	select {
	case filename, ok := <-pitcherStore.pitcherQueue:
		{
			if ok {
				fmt.Print(filename)
				fmt.Fprint(w, filename)
			} else {
				log.Fatalf("Panic! ok was not true")
			}
		}
	case <-time.After(50000 * time.Microsecond):
	}

}
