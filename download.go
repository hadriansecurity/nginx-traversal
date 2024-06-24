package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

var maxConcurrency = 10

func downloadWorker(urlCh <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urlCh {
		outputDir, err := extractOutputDir(url)
		if err != nil {
			fmt.Printf("Error extracting output directory for URL %s: %v\n", url, err)
			continue
		}
		downloadFile(url, outputDir)
	}
}

func downloadFile(url, outputDir string) {
	// Extract the filename from the URL
	filename := path.Base(url)
	outputPath := filepath.Join(outputDir, filename)

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("error creating output directory: %v\n", err)
		return
	}

	// Create the file to which the URL will be downloaded
	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("error creating file: %v\n", err)
		return
	}
	defer file.Close()

	// Perform the GET request to fetch the file
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("error performing GET request: %v\n", err)
		return
	}
	defer response.Body.Close()

	// Check if the GET request was successful
	if response.StatusCode != http.StatusOK {
		fmt.Printf("failed to download file: %s\n", response.Status)
		return
	}

	// Write the body of the response to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		fmt.Printf("error writing to file: %v\n", err)
		return
	}

	fmt.Printf("Downloaded file from %s to %s\n", url, outputPath)
}

func extractOutputDir(url string) (string, error) {
	// Remove "https://" prefix
	url = strings.TrimPrefix(url, "https://")
	// Split URL by "/"
	parts := strings.Split(url, "/")
	// Check if there are enough parts
	if len(parts) < 4 {
		return "", fmt.Errorf("URL %s does not have enough parts", url)
	}
	// Join parts excluding the domain and file name, and include an additional part
	return filepath.Join(parts[1 : len(parts)-2]...), nil
}

func main() {
	// Command-line flags
	filePath := flag.String("file", "", "Path to the file containing URLs")
	flag.Parse()

	// Check if file path is provided
	if *filePath == "" {
		fmt.Println("Please provide the path to the file containing URLs using -file flag")
		return
	}

	// Open the file containing URLs
	file, err := os.Open(*filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup

	// Create a channel to communicate URLs to download workers
	urlCh := make(chan string)

	// Start download workers
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go downloadWorker(urlCh, &wg)
	}

	// Read each line from the file and send URLs to download workers
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		urlCh <- url
	}

	// Close the URL channel to signal download workers that no more URLs will be sent
	close(urlCh)

	// Wait for all workers to finish
	wg.Wait()

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file: %v\n", err)
		return
	}
}
