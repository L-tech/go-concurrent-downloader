package main

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetHeadRequest(t *testing.T) {
	url := "https://d37ci6vzurychx.cloudfront.net/trip-data/yellow_tripdata_2018-05.parquet"
	resp, err := getHeadRequest(url)
	if err != nil {
		t.Errorf("getHeadRequest failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	require.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestGetFileSize(t *testing.T) {
	url := "https://d37ci6vzurychx.cloudfront.net/trip-data/yellow_tripdata_2018-05.parquet"
	fz, err := getFileSize(url)
	if err != nil {
		return
	}
	actualSize := int64(130850283)
	require.Equal(t, fz, actualSize)

}

func TestCreateEmptyFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Errorf("Error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	size := int64(1024)
	err = createEmptyFile(tempFile.Name(), size)
	if err != nil {
		t.Errorf("createEmptyFile failed: %v", err)
	}

	stat, err := tempFile.Stat()
	if err != nil {
		t.Errorf("Error getting file stat: %v", err)
	}
	if stat.Size() != size {
		t.Errorf("Expected file size %d, got %d", size, stat.Size())
	}
	require.Equal(t, stat.Size(), size)
}

func TestDownloadFileInBit(t *testing.T) {
	rawURL := "https://d37ci6vzurychx.cloudfront.net/trip-data/yellow_tripdata_2018-05.parquet"
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		t.Errorf("Error parsing URL: %v", err)
	}
	fileName := path.Base(parsedURL.Path)
	size, err := getFileSize(rawURL)
	if err != nil {
		t.Errorf("Error getting file size: %v", err)
	}

	effort := createEmptyFile(fileName, size/1000)
	if effort != nil {
		return
	}
	filerr := downloadFileInBit(rawURL, fileName, 0, size/1000, 3)
	if filerr != nil {
		t.Errorf("Error downloading chunk: %v", err)
	}
	out, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		t.Errorf("Error Opening file: %v", err)
	}
	defer os.Remove(out.Name())
	stat, err := out.Stat()
	if err != nil {
		t.Errorf("createEmptyFile failed: %v", err)
	}
	require.Equal(t, stat.Size(), (size/1000)+1)

}
