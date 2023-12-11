package main

import (
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/scottrmalley/p2p-file-challenge/api"
	"github.com/scottrmalley/p2p-file-challenge/client"
	"math/rand"
	"strings"
	"time"
)

func mustResolve[T any](in T, err error) T {
	if err != nil {
		panic(err)
	}
	return in
}

func main() {
	nFiles := flag.Int("n", 500, "number of files to upload")
	hosts := flag.String("hosts", "http://127.0.0.1:8080", "comma separated list of hosts")
	flag.Parse()

	fmt.Printf("Generating %d files\n", *nFiles)
	files := make([][]byte, *nFiles)
	for i := 0; i < *nFiles; i++ {
		files[i] = []byte(uuid.New().String())
	}

	hostUrls := strings.Split(*hosts, ",")
	if len(hostUrls) == 0 {
		panic("no hosts specified")
	}

	fmt.Printf("Uploading %d files\n", *nFiles)
	persistence := client.NewInMemoryPersistence()
	client0 := client.NewClient(
		persistence,
		mustResolve(api.NewClient(fmt.Sprintf("%s/api", hostUrls[0]))),
	)

	setId, err := client0.PostFiles(files)
	if err != nil {
		panic(err)
	}

	fileToDownload := rand.Intn(*nFiles)
	fmt.Printf("Downloading random file #%d\n", fileToDownload)
	file, err := client0.GetFile(setId, fileToDownload)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Downloaded file: %s\n", file)

	if len(hostUrls) < 2 {
		fmt.Println("No other hosts specified, exiting")
		return
	}

	// give them some time to gossip
	fmt.Printf("Waiting for gossip to propagate\n")
	time.Sleep(5 * time.Second)

	fmt.Printf("attempting to download from other hosts\n")
	for _, hostUrl := range hostUrls[1:] {
		nodeClient := client.NewClient(
			persistence,
			mustResolve(api.NewClient(fmt.Sprintf("%s/api", hostUrl))),
		)
		file, err := nodeClient.GetFile(setId, fileToDownload)
		if err != nil {
			fmt.Printf("error downloading from %s: %s\n", hostUrl, err)
			continue
		}
		fmt.Printf("Downloaded file from %s: %s\n", hostUrl, file)
	}
}
