package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	cache "github.com/patrickmn/go-cache"
)

const (
	VenvCmd = "source scraper/venv/bin/activate"
	PyCmd   = "python scraper/scraper.py"
)

var (
	c = cache.New(5*time.Minute, 10*time.Minute)
)

type client struct {
	RegNo         string
	Password      string
	Attendance    map[string]*json.RawMessage
	Internalmarks map[string]*json.RawMessage
	Externalmarks map[string]*json.RawMessage
	Timetable     map[string]*json.RawMessage
	Marks         map[string]*json.RawMessage
	Json          string
}

func api(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)

	if r.FormValue("regno") == "" || r.FormValue("password") == "" {
		fmt.Println("Blank")
		return
	}

	curr := &client{}
	curr.RegNo = r.FormValue("regno")
	curr.Password = r.FormValue("password")
	Json, found := c.Get(curr.RegNo)
	if found {
		fmt.Fprintf(w, Json.(string))
		return
	} else {
		curr.SetData()
		c.Set(curr.RegNo, curr.Json, cache.DefaultExpiration)
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
	http.HandleFunc("/", api)
	http.ListenAndServe(":9090", nil)
}
