package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type Task struct {
	duration time.Duration
	channels []chan bool
}

var mutex sync.Mutex
var taskDurations []time.Duration
var guard chan bool
var lastStart time.Time
var taskChannel chan Task

const MAX_TASKS = 100000

func runTasks() {
	for {
		task := <-taskChannel

		mutex.Lock()
		lastStart = time.Now()
		mutex.Unlock()

		go func() {
			time.Sleep(task.duration)
			for _, channel := range task.channels {
				channel <- true
			}
		}()

		<-guard

		mutex.Lock()
		taskDurations = taskDurations[1:]
		mutex.Unlock()
	}
}

func handleAddSync(w http.ResponseWriter, r *http.Request, duration time.Duration) {
	waiter := make(chan bool)
	mutex.Lock()
	taskDurations = append(taskDurations, duration)
	mutex.Unlock()
	taskChannel <- Task{
		duration: duration,
		channels: []chan bool{guard, waiter},
	}
	<-waiter
	log.Print("sync added")
}

func handleAddAsync(w http.ResponseWriter, r *http.Request, duration time.Duration) {
	mutex.Lock()
	taskDurations = append(taskDurations, duration)
	mutex.Unlock()
	taskChannel <- Task{
		duration: duration,
		channels: []chan bool{guard},
	}
	log.Print("async added")
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	timeDuration, err := time.ParseDuration(r.FormValue("timeDuration"))
	if err != nil {
		log.Print("incorrect duration", err)
		return
	}
	if r.FormValue("sync") != "" {
		handleAddSync(w, r, timeDuration)
	} else if r.FormValue("async") != "" {
		handleAddAsync(w, r, timeDuration)
	} else {
		log.Print("sync/async argument not found")
	}
}

func handleSchedule(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	w.Header().Set("Content-Type", "application/json")
	taskDurationStrings := []string{}
	for _, duration := range taskDurations {
		taskDurationStrings = append(taskDurationStrings, duration.String())
	}
	json.NewEncoder(w).Encode(taskDurationStrings)
	mutex.Unlock()
}

func handleTime(w http.ResponseWriter, r *http.Request) {
	var totalTime time.Duration
	mutex.Lock()
	for _, duration := range taskDurations {
		totalTime += duration
	}
	if totalTime > 0 {
		totalTime -= time.Now().Sub(lastStart)
	}
	mutex.Unlock()
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(totalTime.String()))
}

func init() {
	guard = make(chan bool)
	taskChannel = make(chan Task, MAX_TASKS)
	go runTasks()
}

func main() {
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/schedule", handleSchedule)
	http.HandleFunc("/time", handleTime)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
