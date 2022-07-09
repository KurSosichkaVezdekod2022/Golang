package main

import (
	"bufio"
	"log"
	"os"
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

func main() {
	doTasks(getDurations("../test.txt"))
}
