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
	VenvCmd  = "source scraper/venv/bin/activate"
	PyCmd    = "python scraper/scraper.py"
	TimedOut = string('"') + "{ error : 'Request timed out' }" + string('"') + "\n"
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

	fmt.Println(r.Method)

	if r.FormValue("regno") == "" || r.FormValue("password") == "" {
		fmt.Println("Blank")
		return
	}

	curr := &client{} // Current client
	curr.RegNo = r.FormValue("regno")
	curr.Password = r.FormValue("password")

	Json, found := c.Get(curr.RegNo) // Check cache
	if found {
		fmt.Fprintf(w, Json.(string))
	} else {
		curr.SetData()
		if strings.Compare(curr.Json, TimedOut) != 0 { // Don't cache failed responses
			c.Set(curr.RegNo, curr.Json, cache.DefaultExpiration)
		}
		fmt.Fprintf(w, curr.Json)
	}
}

func (c *client) Run() (string, error) {
	command := fmt.Sprintf("%s; %s %s %s;", VenvCmd, PyCmd, c.RegNo, c.Password)
	fmt.Println(command)
	cmd := exec.Command("sh", "-c", command)
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

func (c *client) SetData() {
	response, err := c.Run()
	if err != nil {
		log.Fatal("Error executing command : ", err)
	}
	c.Json = response
	/*
		json.Unmarshal([]byte(response), c.Json)
		json.Unmarshal(*c.Json["Attendance"], c.Attendance)
		json.Unmarshal(*c.Json["Timetable"], c.Timetable)
		json.Unmarshal(*c.Json["Marks"], c.Marks)
		json.Unmarshal(*c.Marks["External Marks"], c.Externalmarks)
		json.Unmarshal(*c.Marks["Internal Marks"], c.Internalmarks)
	*/
}

func main() {
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":9090", nil)
}
