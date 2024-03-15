# Go Basic Concurrent Downloader

## Overview

This Go Concurrent Downloader is a command-line tool that allows you to download large files from the internet in chunks, leveraging Go's concurrency features to potentially speed up the download process. The tool splits the file into several parts and downloads each part concurrently, making efficient use of network bandwidth and reducing overall download time.

## Features

- Download files in chunks concurrently.
- Customize the number of concurrent downloads.
- Automatically determine file name and size from the URL.

## Installation

### Prerequisites

- Go 1.19 or later.

### Steps

1. Clone the repository:

```bash
git clone https://github.com/l-tech/go-concurrent-downloader.git
```

2. Navigate to the project directory:

```bash
cd go-concurrent-downloader
```

3. Build the project:

```bash
go build -o downloader
```

## Usage

After building the project, you can start downloading files using the following command format:

```bash
./downloader -url "<file_url>" -n <number_of_chunks> -chunk-size <bytes_per_batch > -retires <number_of_retires>
```

- `<file_url>`: The URL of the file you wish to download.
- `<number_of_chunks>`: The number of parts to split the file into for concurrent downloading. 
- `<chunk-size>`: Chunk size for downloading (in bytes). If 0, it will be calculated automatically.
- `<retries>`: Maximum number of retries for failed downloads.


### Example

To download a file using 10 concurrent connections with a chunk size of 8723352 and 5 retries if any error occurs:

```bash
./downloader -url "https://example.com/path/to/file.parquet" -n 10 -chunk-size 8723352 -retires 5
```
