package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/scottrmalley/p2p-file-sharing/api"
	"github.com/scottrmalley/p2p-file-sharing/client"
	"github.com/scottrmalley/p2p-file-sharing/config"
	"github.com/scottrmalley/p2p-file-sharing/proof"
)

func mustResolve[T any](in T, err error) T {
	if err != nil {
		panic(err)
	}
	return in
}

// randomNodeUpload will take a fileset and upload each file to a
// random node in the network. Since each node is connected
// to each other node, the fileset will eventually be available
// on all nodes.
func randomNodeUpload(nFiles int, clients []*client.Client) {
	if len(clients) == 0 {
		panic("no clients provided")
	}

	fmt.Printf("Generating %d files\n", nFiles)
	files := make([][]byte, nFiles)
	for i := 0; i < nFiles; i++ {
		files[i] = []byte(uuid.New().String())
	}

	root := mustResolve(proof.Root(files))

	// create a fileset
	setId := mustResolve(clients[0].CreateSet(root, nFiles))

	fmt.Printf("Uploading %d files to random nodes\n", nFiles)
	for i, file := range files {
		// select a random node to upload to
		node := clients[rand.Intn(len(clients))]
		if err := node.AddFile(setId, i, file); err != nil {
			panic(err)
		}
	}

	// give them some time to gossip
	fmt.Printf("Waiting for gossip to propagate\n")
	time.Sleep(1 * time.Second)

	// for each other host we can also try to download the file
	fileToDownload := rand.Intn(nFiles)
	fmt.Printf("Downloading random file #%d\n", fileToDownload)
	for i, node := range clients {
		file, err := node.GetFile(setId, fileToDownload)
		if err != nil {
			fmt.Printf("error downloading from node %d: %s\n", i, err)
			continue
		}
		fmt.Printf("Downloaded file from node %d: %s\n", i, file)
	}

}

// singleNodeUpload will take a fileset and upload each file to a
// single node in the network.
func singleNodeUpload(nFiles int, clients []*client.Client) {
	if len(clients) == 0 {
		panic("no clients provided")
	}

	// generate files
	fmt.Printf("Generating %d files\n", nFiles)
	files := make([][]byte, nFiles)
	for i := 0; i < nFiles; i++ {
		files[i] = []byte(uuid.New().String())
	}

	// upload files to the first node
	fmt.Printf("Uploading %d files to node 0\n", nFiles)
	setId := mustResolve(clients[0].PostFiles(files))

	// give them some time to gossip
	fmt.Printf("Waiting for gossip to propagate\n")
	time.Sleep(1 * time.Second)

	// for each other host we can also try to download the file
	fileToDownload := rand.Intn(nFiles)
	fmt.Printf("Downloading random file #%d\n", fileToDownload)
	for i, node := range clients {
		file, err := node.GetFile(setId, fileToDownload)
		if err != nil {
			fmt.Printf("error downloading from node %d: %s\n", i, err)
			continue
		}
		fmt.Printf("Downloaded file from node %d: %s\n", i, file)
	}
}

func main() {
	cfg := config.ParseClientEnv("SVC")
	nFiles := cfg.N
	hostUrls := cfg.Hosts

	// create persistence
	persistence := client.NewInMemoryPersistence()

	// create clients for each host
	clients := make([]*client.Client, len(hostUrls))
	for i, hostUrl := range hostUrls {
		clients[i] = client.NewClient(
			persistence,
			mustResolve(api.NewClient(fmt.Sprintf("%s/api", hostUrl))),
		)
	}

	// for now, we have two simple test cases, one where we upload the
	// entire fileset to a single node, and another where we upload
	// each file to a random node.
	fmt.Println("--- Single Node Upload ---")
	singleNodeUpload(nFiles, clients)

	fmt.Println("\n--- Random Node Upload ---")
	randomNodeUpload(nFiles, clients)

}
