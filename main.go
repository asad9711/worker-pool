package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	webPages = []string{
		"yahoo.com",
		"google.com",
		"bing.com",
		"amazon.com",
		"github.com",
		"gitlab.com",
	}

	results struct {
		// put here content length of each page
		ContentLength map[string]int

		// accumulate here the content length of all pages
		TotalContentLength int

		sync.Mutex
	}
)

// utility to print result
func printResult(resultMap map[string]int) {
	fmt.Println("\n ******** RESULT **********")
	for k, v := range resultMap {
		fmt.Println(k, " - ", v)
	}

}

// utility to get html content length of webpage
func fetchHTML(url string) int {
	fullURL := "https://" + url
	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("error in retrieving url ", url, err.Error())
		return -1
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return len(html)

}

// worker to get content length of received webaddress and update the result map
func worker(workerID int, wg *sync.WaitGroup, ctx context.Context, buffChan chan string) {

	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker:", workerID, "TERMINATION SIGNAL RECEIVED. TERMINATING ....")
			return

		case webAddress, areWebPagesPending := <-buffChan:
			fmt.Println("worker:", workerID, "received address :", webAddress, "to process ")
			if !areWebPagesPending {
				fmt.Println("no more web pages to process")
				return
			}

			contentLength := fetchHTML(webAddress)
			// fmt.Println("acquiring mutex lock")
			results.Mutex.Lock()
			results.ContentLength[webAddress] = contentLength
			// fmt.Println("releasing mutex lock")
			results.Mutex.Unlock()
			if contentLength != -1 {
				results.TotalContentLength += contentLength
			}
			fmt.Println("worker:", workerID, ": done with processing the web addresss: ", webAddress)
		}
	}

}

func main() {

	/*
		HOW TO RUN ?

		go run main.go <timeout_value(integer)>

	*/

	// validate that timeout parameter is passed
	if len(os.Args) < 2 {
		fmt.Println("Please provide timeout value(in seconds). e.g go run main.go 10")
		return
	}

	// validate that provided timeout is a +ve number
	timeoutValue, _ := strconv.Atoi(os.Args[1])
	if timeoutValue < 1 {
		fmt.Println("please pass +ve value for timeout")
		return
	}
	timeoutDuration := time.Second * time.Duration(timeoutValue)

	// allocate memory to map inside results
	results.ContentLength = make(map[string]int)

	// create context with provided timeout
	ctx, cancelFunction := context.WithTimeout(context.Background(), timeoutDuration)

	// buffered channel to communicate tasks to workers
	buffChan := make(chan string, 2)

	var wg sync.WaitGroup

	wg.Add(1)
	go worker(1, &wg, ctx, buffChan)

	wg.Add(1)
	go worker(2, &wg, ctx, buffChan)

	go func() {
		for _, webPage := range webPages {
			buffChan <- webPage
		}
		close(buffChan)
	}()

	// listen for timeout event, and accordingly cancel the context to terminate the workers
	go func() {
		<-time.After(timeoutDuration)
		fmt.Println("\n ^^^^^^^^  EXECUTION TIME OVER. TERMINATING ALL OPEN WORKERS GRACEFULLY ^^^^^^^^")
		cancelFunction()

	}()

	// wait for the workers to complete
	wg.Wait()

	printResult(results.ContentLength)
}
