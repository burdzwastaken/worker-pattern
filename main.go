package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	client "github.com/burdzwastaken/worker-pattern/clients"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var (
	version   = "0.0.0"
	buildDate = ""
)

const (
	errorQueue            = "error-queue"
	activeQueue           = "active-queue"
	completedQueue        = "completed-queue"
	completedWorkersQueue = "completed-workers"
)

func main() {
	host := flag.String("host", "redis:6379", "Redis host")
	password := flag.String("password", "", "Redis password")
	workers := flag.Int("workers", 3, "Workers in pool")
	iterations := flag.Int("iterations", 100, "Work to be processed")
	version := flag.Bool("v", false, "prints current version")
	flag.Parse()

	if *version {
		printVersion()
		os.Exit(0)
	}

	ctx := context.Background()

	redis := client.NewClient(ctx, *host, *password)

	err := redis.HealthCheck()
	if err != nil {
		log.Printf("Error when running: %s", err.Error())
	}

	populateRedis(redis.Context, redis, *iterations)
	populateWorkers(redis.Context, redis, *workers)
}

// prints version
func printVersion() {
	v := map[string]interface{}{
		"version":    version,
		"build_date": buildDate,
	}
	encoder := json.NewEncoder(os.Stderr)
	encoder.SetIndent("", "  ")
	encoder.Encode(v)
}

// Generates a random delay for workers between 10-60 ms
func generateRandomDelay() int {
	rand.Seed(time.Now().UnixNano())
	min := 10
	max := 60
	return (rand.Intn(max-min+1) + min)
}

// Add information to Redis to be processed by workers
// This will add a command to `sleep` with a random delay
func populateRedis(ctx context.Context, client *client.Client, iterations int) {
	err := client.DelHashKey(activeQueue)
	if err != nil {
		log.Printf("Error when running: %s", err.Error())
	}

	for i := 0; i < iterations; i++ {
		keyName := fmt.Sprintf("command-%d", i)

		command := map[string]interface{}{
			"duration": generateRandomDelay(),
		}

		err := client.HashSet(keyName, command)
		if err != nil {
			log.Printf("Error when running: %s", err.Error())
			continue
		}

		var expireInSeconds time.Duration
		expireInSeconds = 100
		duration := (time.Duration(expireInSeconds) * time.Second)
		err = client.Expire(keyName, duration)
		if err != nil {
			log.Printf("Error when running: %s", err.Error())
			continue
		}

		err = client.RightPush(activeQueue, keyName)
		if err != nil {
			log.Printf("Error when running: %s", err.Error())
			continue
		}
	}

	length, err := client.ListLength(activeQueue)
	if err != nil {
		log.Printf("Error when running: %s", err.Error())
	} else {
		log.Printf("Populated %d hashes out of %d\n", length, iterations)
	}
}

// Create the amount of workers specified
func populateWorkers(ctx context.Context, client *client.Client, workers int) {
	var wg sync.WaitGroup

	log.Printf("Creating %d workers\n", workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go doWork(ctx, client, i, &wg)
	}

	wg.Add(1)
	go pollWorkers(ctx, client, workers, &wg)

	wg.Wait()
	log.Printf("All workers completed\n")
}

func doWork(ctx context.Context, client *client.Client, workerID int, wg *sync.WaitGroup) {
	defer wg.Done()

	var sleptFor time.Duration
	sleptFor = 0
	for {
		keyID, err := client.LeftPop(activeQueue)
		if err != nil {
			if err != redis.Nil {
				log.Printf("Error when running: %s", err.Error())
			}
			break
		}

		keyName, err := client.HashGetAll(keyID)
		if err != nil {
			if err != redis.Nil {
				log.Printf("Error when running: %s", err.Error())
			}
		}

		duration, err := sleep(client.Context, keyID, keyName, workerID)
		if err != nil {
			log.Printf("Error when running: %s", err.Error())
		}

		err = client.Publish(completedQueue, keyID, workerID)
		if err != nil {
			log.Printf("Error when running: %s", err.Error())
		}

		sleptFor = sleptFor + duration
	}

	err := client.Publish(completedWorkersQueue, "", workerID)
	if err != nil {
		log.Printf("Error when running: %s", err.Error())
	}

	log.Printf("Worker %d slept for a total of %v seconds", workerID, sleptFor.Seconds())
}

// Sleeps for a specific time as set in Redis
func sleep(ctx context.Context, keyID string, keyName map[string]string, workerID int) (time.Duration, error) {
	duration, err := strconv.Atoi(keyName["duration"])
	if err != nil {
		return 0, errors.Wrapf(err, "Error while converting %v to an integer", keyName["duration"])
	}
	durationMS := (time.Duration(duration) * time.Millisecond)
	log.Printf("Worker %d performing sleep for %dms\n", workerID, durationMS.Milliseconds())
	time.Sleep(durationMS)

	return durationMS, nil
}

// Workers will publish their status back to the completedWorkersChannel
func pollWorkers(ctx context.Context, client *client.Client, workers int, wg *sync.WaitGroup) {
	defer wg.Done()
	finishedWorkers := 0
	completedWorkersChannel := client.Subscribe(completedWorkersQueue)

	for finishedWorkers < workers {
		select {
		case _ = <-completedWorkersChannel.Channel():
			finishedWorkers++
			log.Printf("Worker has finished. %d/%d are completed\n", finishedWorkers, workers)
		}
	}

	err := completedWorkersChannel.Close()
	if err != nil {
		log.Printf("Error when running: %s", err.Error())
	}
}
