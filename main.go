package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

func getHeadRequest(url string) (*http.Response, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getFileSize(url string) (int64, error) {
	resp, err := getHeadRequest(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("CONTENT LENGTH NOT FOUND")
	}
	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("ERROR CONVERTING TO INT64")
	}
	return size, nil

}

func getFileETag(url string) (string, error) {
	resp, err := getHeadRequest(url)
	if err != nil {
		return "failed", err
	}
	defer resp.Body.Close()
	etag := resp.Header.Get("ETag")
	if etag == "" {
		return "", fmt.Errorf("ETag header is missing")
	}
	etag = strings.Trim(etag, "\"")
	if strings.HasPrefix(etag, "W/") {
		etag = etag[2:]
	}
	return etag, nil
}

func createEmptyFile(path string, size int64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Seek(size-1, io.SeekStart)
	file.Write([]byte{0})
	return nil
}

func downloadFileInBit(url, destFilePath string, start, end int64, maxRetries int) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
		req.Header.Add("Range", rangeHeader)

		resp, err := client.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return err
			}
			fmt.Printf("Download attempt %d failed: %v\n", attempt+1, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusPartialContent {
			err = fmt.Errorf("server does not support range requests, got status code: %d", resp.StatusCode)
			if attempt == maxRetries {
				return err
			}
			fmt.Printf("Download attempt %d failed: %v\n", attempt+1, err)
			continue
		}

		out, err := os.OpenFile(destFilePath, os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = out.Seek(start, io.SeekStart)
		if err != nil {
			return err
		}

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			if attempt == maxRetries {
				return err
			}
			fmt.Printf("Download attempt %d failed: %v\n", attempt+1, err)
			continue
		}

		return nil
	}

	return fmt.Errorf("maximum number of retries (%d) exceeded", maxRetries)
}

func main() {
	var rawURL string
	var count int
	var chunkSize int64
	var maxRetries int

	flag.StringVar(&rawURL, "url", "", "URL of the file to download.")
	flag.IntVar(&count, "n", 4, "Number of goroutines to use for downloading.")
	flag.Int64Var(&chunkSize, "chunk-size", 0, "Chunk size for downloading (in bytes). If 0, it will be calculated automatically.")
	flag.IntVar(&maxRetries, "retries", 3, "Maximum number of retries for failed downloads")
	flag.Parse()

	if rawURL == "" {
		fmt.Println("URL is required")
		return
	}

	// Parse the raw URL to a URL object
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}
	fileName := path.Base(parsedURL.Path)
	size, err := getFileSize(rawURL)
	path := fileName
	if err != nil {
		fmt.Println("Error getting file size:", err)
		return
	}
	fmt.Printf("The size of this url - %s is %d\n", rawURL, size)
	etag, err := getFileETag(rawURL)
	if err != nil {
		fmt.Println("Error getting Etage:", err)
		return
	}
	fmt.Printf("The ETag of this url - %s is %s\n", rawURL, etag)
	effort := createEmptyFile(path, size)
	if effort != nil {
		return
	}
	start := time.Now()

	var wg sync.WaitGroup
	if chunkSize == 0 {
		chunkSize = size / int64(count)
	}
	for i := 0; i < count; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == count-1 {
			end = size
		}
		wg.Add(1)
		go func(start, end int64) {
			defer wg.Done()
			err := downloadFileInBit(rawURL, path, start, end, maxRetries)
			if err != nil {
				fmt.Printf("Error downloading chunk %d-%d: %v\n", start, end, err)
			}
		}(start, end)
	}
	wg.Wait()

	fmt.Println("Download completed successfully.")
	duration := time.Since(start)
	fmt.Printf("Downloaded file in %v\n", duration)
}
