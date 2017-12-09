package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
)

const (
	seperator = ":"
	port      = "8080"
	VenvCmd   = "source scraper/venv/bin/activate"
	PyCmd     = "python scraper/scraper.py"
	TimedOut  = string('"') + "{ error : 'Request timed out' }" + string('"') + "\n"
)

var (
	c = cache.New(5*time.Minute, 10*time.Minute)
)

type client struct {
	RegNo    string
	Password string

	Timetable     map[string]*json.RawMessage
	Attendance    map[string]*json.RawMessage
	Internalmarks map[string]*json.RawMessage
	Externalmarks map[string]*json.RawMessage
	Marks         map[string]*json.RawMessage

	Json string
}

func Handler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")

	if r.FormValue("regno") == "" || r.FormValue("password") == "" {
		return
	}

	curr := &client{} // Current client
	curr.RegNo = r.FormValue("regno")
	curr.Password = r.FormValue("password")

	Json, found := c.Get(curr.RegNo) // Check cache
	if found {
		fmt.Fprintf(w, Json.(string)) // If response found in cache, print it
		return
	} else {
		err := curr.Run() // Set the clients data (Invoke scraper)
		if err != nil {
			curr.Json = err.Error()
			log.Fatal(err) // Change this
		}
		if strings.Compare(curr.Json, TimedOut) != 0 { // Don't cache failed responses (Wrong credentials)
			c.Set(curr.RegNo, curr.Json, cache.DefaultExpiration) // Set cache
		}
		fmt.Fprintf(w, curr.Json) // Print json
	}
}

func (c *client) Run() error {
	command := fmt.Sprintf("%s; %s %s %s;", VenvCmd, PyCmd, c.RegNo, c.Password)
	cmd := exec.Command("sh", "-c", command)
	stdout, err := cmd.Output()
	if err != nil {
		return err // To do : Send custom error along with stderr log
	}
	c.Json = string(stdout)
	return nil
}

func main() {
	addr := seperator + port
	http.HandleFunc("/", Handler)
	http.ListenAndServe(addr, nil)
}
