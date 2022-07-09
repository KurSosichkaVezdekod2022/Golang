package main

import (
	"bufio"
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

func main() {
	doTasksParallel(getDurations("../test.txt"))
}
