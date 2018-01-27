package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
)

const (
	seperator = ":"
	port      = "8080"
	VenvCmd   = "source scraper/venv/bin/activate"
	PyCmd     = "python scraper/scraper.py"
	TimedOut  = "{ error : 'Request timed out' }"
)

var (
	c = cache.New(5*time.Minute, 10*time.Minute)
)

type client struct {
	RegNo    string
	Password string
	ID       int64
	Json     string
}

func Handler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		log.Printf("Recieved %s request. Invalid", r.Method)
		fmt.Fprintf(w, "Invalid request")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")

	if r.FormValue("regno") == "" || r.FormValue("password") == "" {
		log.Printf("Recieved blank param values")
		return
	}

	fmt.Println("\n--- New Request ---")
	defer fmt.Println("\n-------------------")

	unique_id := time.Now().Unix()

	curr := &client{} // Current client
	curr.RegNo = r.FormValue("regno")
	curr.Password = r.FormValue("password")
	curr.ID = unique_id

	Json, found := c.Get(curr.RegNo) // Check cache
	if found {
		log.Printf("Found JSON in cache for user %s", curr.RegNo)
		fmt.Fprintf(w, Json.(string)) // If response found in cache, print it
		return
	} else {
		log.Printf("Couldn't find user in cache, running scraper")
		err := curr.Run() // Set the clients data (Invoke scraper)
		if err != nil {
			log.Printf("Encountered error while scraping %s", err.Error())
			curr.Json = err.Error()
			fmt.Fprintf(w, curr.Json)
			return
		}
		if strings.Compare(curr.Json, TimedOut) != 0 { // Don't cache failed responses (Wrong credentials)
			c.Set(curr.RegNo, curr.Json, cache.DefaultExpiration) // Set cache
		}
		log.Println("Successfully printed json")
		fmt.Fprintf(w, curr.Json) // Print json
	}
}

func (c *client) Run() error {

	filePath := "results/" + strconv.FormatInt(c.ID, 10) + ".json"

	defer func() {
		_, err := os.Stat(filePath)
		if err != nil {
			log.Println("Json file doesn't exist")
		} else {
			os.Remove(filePath)
		}
	}()

	command := fmt.Sprintf("%s; %s %s %s %s;", VenvCmd, PyCmd, c.RegNo, c.Password, filePath)
	cmd := exec.Command("sh", "-c", command)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err // To do : Send custom error along with stderr log
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	in := bufio.NewScanner(stderr)
	for in.Scan() {
		log.Printf(in.Text())
	}
	if err := in.Err(); err != nil {
		log.Printf("error : %s", err)
	}

	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	c.Json = string(dat)

	return nil
}

func main() {
	addr := seperator + port
	http.HandleFunc("/", Handler)
	log.Fatal(http.ListenAndServe(addr, nil))
}
