package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

func getDurations(filePath string) []time.Duration {
	file, err := os.Open(filePath)
	durations := []time.Duration{}
	if err != nil {
		log.Print(err)
		return durations
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		duration, err := time.ParseDuration(scanner.Text())
		if err != nil {
			log.Print(err)
			continue
		}
		durations = append(durations, duration)
	}
	return durations
}

func doTasks(durations []time.Duration) {
	for i, duration := range durations {
		log.Print("starting task ", i)
		time.Sleep(duration)
		log.Print("task ", i, " completed")
	}
}

func doTasksParallel(durations []time.Duration) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(durations))
	for i, duration := range durations {
		go func(i int, duration time.Duration) {
			log.Print("starting task ", i)
			time.Sleep(duration)
			log.Print("task ", i, " completed")
			waitGroup.Done()
		}(i, duration)
	}
	waitGroup.Wait()
}

func doParallelTask(duration time.Duration, id int, guard chan bool, waitGroup *sync.WaitGroup) {
	log.Print("starting task ", id)
	time.Sleep(duration)
	log.Print("task ", id, " completed")
	waitGroup.Done()
	<-guard
}

func doTasksParrallelLimited(durations []time.Duration, processNumber int) {
	guard := make(chan bool, processNumber)
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(durations))

	for i, duration := range durations {
		guard <- false
		go doParallelTask(duration, i, guard, &waitGroup)
	}

	waitGroup.Wait()
}

func main() {
	processNumber := 0
	_, err := fmt.Scanf("%d", &processNumber)
	if err != nil {
		log.Print("could not read process number from stdin(")
		return
	}

	doTasksParrallelLimited(getDurations("../test.txt"), processNumber)
}
