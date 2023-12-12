# P2P File Sharing

## Description

This is a simple P2P file sharing backend, that can store file sets, and return Merkle inclusion proofs upon file
retrieval. It is built using the [libp2p](https://libp2p.io/) library, and uses Gossip Sub for message propagation.

## Design

The system is designed to be run on multiple machines. Each node subscribes to a file topic, and will receive a
message each time a file is uploaded to the network. The node will store the file, and once the file set is 
completely uploaded, it can be retrieved from any node in the network.

The API is basic and allows for the following routes:

```shell
POST /api/sets/{set_id}/files/{index}

// BODY
{
  "content": "0x66696c6531...", // hex encoded file content
  "setCount": 13 // The total number of files in the set
}

// RESPONSE
{
  "success": true, // whether the file was successfully uploaded
  "hash": "0c2a4d2a..." // the file hash
}
```

```shell
GET /api/sets/{set_id}/files/{index}

// RESPONSE
{
  "file": "0x66696c6531...", // hex encoded file content
  "proof": {
    "proof": ["0x0c2a4d2a..."], // the merkle node hashes
    "index": 0, // the index of the file in the set
  }
}
```
Path parameters:
- `set_id`: The ID of the file set to upload to (if it doesn't exist, it will be created)
- `index`: The index of the file in the set (initial file order is set by the client)

The project also includes a small client library that can be used to upload and download files from the network. In 
order to prove that the files are being stored correctly, the client library includes a small persistence layer that 
saves:
- The file set id
- The file set size
- The merkle root of the file set

When a file is downloaded, the client library will verify that the merkle proof is valid (ie. the reconstructed 
Merkle root matches the one stored in the persistence layer), and will return an error if it is not.

## Usage

This project is set up using [Docker Compose](https://docs.docker.com/compose/). Each backend node is intended to
run on a different machine, so for demonstration purposes, the `docker-compose.yml` file is set up to run 3 nodes
on the same machine, each with a different port exposed.

The `docker-compose.yml` is also set up to run an example client script. In order to separate these two functions,
the backend / client containers are set up to be ran in different profiles.

To run the backend nodes, run the following command:

```shell
docker compose --profile backend up
```

To run the example client script, run the following command:

```shell
docker compose --profile client up
```

## Choices

### HTTP API

The API is currently a simple REST API that uses JSON to upload and download files. I chose to do this as the file
sizes are small, so including file content as a hex encoded string in the JSON payload is not too inefficient. In
addition, the current API design means that file metadata is not persisted, only the file content. For demonstration
purposes, I think this is sufficient to show how such a system could work, but depending on what the production use
case would actually be, we may design this differently. For instance, if the end product was intended to be a CLI
tool, then I would probably use `multipart/form-data` to upload files, instead of JSON.

### Gossip Sub
Gossip Sub is a pub-sub protocol that uses a mesh network topology. This means that each node will be connected to a 
limited set of full peers, and a larger set of metadata-only peers. Gossiping is done by randomly selecting a subset 
of the metadata only peers and reporting on which messages were seen. If by gossiping, a node finds out that it is
missing messages, it will request them, and broadcast them to full peers. It presents a good tradeoff between
scalability and efficiency, as it allows for a large number of nodes to be connected to the network, without
requiring each node to be connected to every other node.

It would have also been possible to use other forms of p2p communication between nodes, however I envisioned a scenario 
where nodes might not want to replicate the entire file set, so by using PubSub, we could eventually devise more complex 
protocols where nodes only subscribe to a limited set of topics, each representing a segregated fileset, and only 
replicate a subset of the network state. This is not how the current implementation works, but it could be a future
extension.

## Future Improvements

### Backsyncing

At the moment, when a node joins the network, it will subscribe to the `file-set` topic. This means that it will
receive a message each time there is a new file to store, and it should have a record of every file posted to
the network from that point on.

This is not ideal, as we currently don't have any way of syncing nodes to the current state of the network, they
will only replicate future state. It's possible for us to add a backsyncing mechanism, where nodes can request
missing file sets from other nodes, and then subscribe to the `file-set` topic. This does get a bit more complicated 
though, as we don't necessarily have trust between nodes, so we would have to build consensus about the current state.

### Filesharing

In the current design, files are shared between network nodes by publishing them to a fixed topic, subscribed to by
all nodes. This works fine as the challenge specified that the file sizes were small, however in a production
scenario, we would probably want the file topic to just announce file availability, and then to

### Node Discovery

Node discovery is currently done by mDNS, which works for demonstration purposes, as all nodes are running on the
same local network. In a production use case, this should be migrated to use a DHT or some other method, as we
wouldn't expect all nodes to be on the same network.

### Batch Uploading

Currently, the api is designed to upload files one at a time. This is fine for demonstration purposes, but in
production, if file sets start to get very long, the api should be updated to allow for batch uploading of files.

### Merkle Tree Optimizations
In the current implementation, Merkle proofs are generated when a file is requested from the network. This means 
that the backend will query the database for all files in the set, and then generate the Merkle proof. This is
not ideal, as for longer filesets, or filesets with large files, this could be a very expensive operation.

In order to improve this, we could store the Merkle tree nodes in the database, and upon generating a proof, select 
only the nodes that are relevant for the proof (this would be ~log(n) nodes, where n is the number of files in the 
set). This would allow us to generate proofs much more efficiently, and with lower memory costs, however it would 
come with the tradeoff of having to store more data in the database.

I would say the final decision on this would come down to the use-case. If we are trying to optimize for storage, 
and upload speeds, then the current implementation has it's benefits, however if we need to optimize for retrieval, 
then we should probably store the Merkle tree nodes in the database, or even precompute the proofs and store them 
with the hashes once a file set is complete.

There are further smaller optimizations we could make. For instance if we expect the file sizes to be large, we 
could already optimize the Merkle tree by storing the precomputed file hashes alongside the file content, which 
would mean we don't need to load the entire file into memory to compute the hash.

### Encryption

There is currently no encryption on files in the system, so when files are replicated across the network, they are
fully visible to all nodes. In a production use case, we would probably want to encrypt the files before they are
uploaded to the network, and then decrypt them when they are downloaded. In it's base form, this wouldn't require 
changes to the underlying protocol, but could be implemented as a wrapper around the file upload / download client.

## Dependencies

This project makes use of the following libraries:
- [libp2p](https://libp2p.io/): A modular networking stack for peer-to-peer applications
- [go-merkletree](https://github.com/wealdtech/go-merkletree): A simple merkle tree implementation in Golang
- [Gin Web Framework](https://github.com/gin-gonic/gin): A performant web framework for writing Golang APIs
- [Gorm](https://gorm.io/): A simple ORM for Golang
