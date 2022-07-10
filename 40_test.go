package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

const timeEps = time.Millisecond * 10

func add(serverURL, sync, duration string, t *testing.T) time.Duration {
	postBody, _ := json.Marshal(map[string]string{
		sync:           "true",
		"timeDuration": duration,
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
		t.Fatal(err)
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

func abs(t time.Duration) time.Duration {
	if t < 0 {
		return -t
	}
	return t
}

func evaluateGetTime(t *testing.T, serverURL string, trueAns, maxDifference time.Duration) {
	ans := getTime(serverURL, t)
	if abs(ans-trueAns) > maxDifference {
		t.Fatal("Incorrect answer: ", ans, " != ", trueAns)
	}
}

func evaluateGetSchedule(t *testing.T, serverURL, trueAns string) {
	ans := getSchedule(serverURL, t)
	if strings.TrimSpace(ans) != trueAns {
		t.Fatal("Incorrect answer: ", ans, " != ", trueAns)
	}
}

func evaluateAdd(t *testing.T, serverURL, sync, duration string, trueRequestDuration time.Duration) {
	requestDuration := add(serverURL, sync, duration, t)
	if abs(requestDuration-trueRequestDuration) > timeEps {
		t.Fatal("Incorrect request duration: ", requestDuration, " != ", trueRequestDuration)
	}
}

func Test1(t *testing.T) {
	serverURL := "http://localhost:8080"
	evaluateAdd(t, serverURL, "sync", "1s", time.Second)
	go evaluateAdd(t, serverURL, "sync", "2s", time.Second*2)
	time.Sleep(time.Millisecond)
	go evaluateAdd(t, serverURL, "sync", "3s", time.Second*5)
	time.Sleep(time.Millisecond)
	go evaluateAdd(t, serverURL, "async", "2s", 0)
	time.Sleep(time.Millisecond)
	go evaluateGetSchedule(t, serverURL, "[\"2s\",\"3s\",\"2s\"]")
	time.Sleep(time.Millisecond)
	go evaluateGetTime(t, serverURL, time.Second*7, timeEps)
	time.Sleep(time.Second * 8)
}

func Test2(t *testing.T) {
	serverURL := "http://localhost:8080"
	go evaluateGetSchedule(t, serverURL, "[]")
	time.Sleep(time.Millisecond)
	go evaluateAdd(t, serverURL, "async", "400ms", 0)
	go evaluateAdd(t, serverURL, "async", "0.4s", 0)
	time.Sleep(time.Millisecond)
	go evaluateAdd(t, serverURL, "sync", "0.5s", time.Millisecond*1300)
	time.Sleep(time.Millisecond)
	go evaluateGetSchedule(t, serverURL, "[\"400ms\",\"400ms\",\"500ms\"]")
	time.Sleep(time.Second * 2)
}

func TestMany(t *testing.T) {
	serverURL := "http://localhost:8080"
	for i := 0; i < 100; i++ {
		go evaluateAdd(t, serverURL, "async", "5s", 0)
		time.Sleep(time.Millisecond)
	}
	evaluateGetTime(t, serverURL, time.Second*500+1, time.Second*3)
}
