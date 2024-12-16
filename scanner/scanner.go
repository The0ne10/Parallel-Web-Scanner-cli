package scanner

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

type ResponseData struct {
	Url     string
	Headers map[string]string
	Status  string
}

var (
	input   string // для указания пути к файлу с URL.
	workers int    // для задания числа горутин.
	timeout int    // для настройки таймаута HTTP-запросов.
)

func Run() {
	readArg()
	validateInputParameters()
	readInput()
}

func readArg() {
	flag.StringVar(&input, "input", "", "input file path")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "number of workers")
	flag.IntVar(&timeout, "timeout", 5, "timeout in seconds")
	flag.Parse()
}

func validateInputParameters() {
	if input == "" {
		log.Fatal("input file path is required")
	}

	if workers <= 0 {
		log.Fatal("number of workers must be greater than zero")
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		log.Fatal("input file path does not exist")
	}
}

func readInput() {
	wg := sync.WaitGroup{}
	scanUrls := make(chan string) // Канал с адрессами
	results := make(chan ResponseData)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	file, err := os.Open(input)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go Worker(ctx, scanUrls, results, &wg)
	}

	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := scanner.Text()
			select {
			case scanUrls <- url:
			case <-ctx.Done():
				break
			}
		}

		close(scanUrls)
	}()

	go func() {
		for result := range results {
			jsonData, err := json.MarshalIndent(result, "", "")
			if err != nil {
				fmt.Println("Error marshalling json data:", err)
				return
			}
			fmt.Println(string(jsonData))
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Timeout reached, terminating...")
	case <-time.After(time.Duration(timeout) * time.Second): // Дополнительный safeguard
		fmt.Println("All tasks processed.")
	}
}
