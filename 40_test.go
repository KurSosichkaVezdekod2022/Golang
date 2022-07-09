package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func add(serverURL, sync, duration string, t *testing.T) time.Duration {
	postBody, _ := json.Marshal(map[string]string{
		"name":  "Toby",
		"email": "Toby@example.com",
	})
	responseBody := bytes.NewBuffer(postBody)

	start := time.Now()
	_, err := http.Post(fmt.Sprintf("%s/add?timeDuration=%s&%s=true", serverURL, duration, sync), "application/json", responseBody)
	if err != nil {
		t.Fatal(err)
	}
	return time.Now().Sub(start)
}

func getBodyByURL(url string, t *testing.T) string {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalln(err)
	}
	return string(body)
}

func getSchedule(serverURL string, t *testing.T) string {
	return getBodyByURL(serverURL+"/schedule", t)
}

func getTime(serverURL string, t *testing.T) time.Duration {
	bodyString := getBodyByURL(serverURL+"/time", t)
	duration, err := time.ParseDuration(bodyString)
	if err != nil {
		t.Fatal(err)
	}
	return duration
}

func Test1(t *testing.T) {
}
