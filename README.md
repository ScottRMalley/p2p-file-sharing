# P2P File Challenge

## Description

This is a simple P2P file sharing network. It is built using the [libp2p](https://libp2p.io/) library, and uses
gossipsub for message propagation.

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
### API
The API is currently a simple REST API that uses JSON to upload and download files. I chose to do this as the file 
sizes are small, so including file content as a hex encoded string in the JSON payload is not too inefficient. In 
addition, the current API design means that file metadata is not persisted, only the file content. For demonstration 
purposes, I think this is sufficient to show how such a system could work, but depending on what the production use 
case would actually be, we may design this differently. For instance, if the end product was intended to be a CLI 
tool, then I would probably use `multipart/form-data` to upload files, instead of JSON.

## Future Improvements
### Backsyncing
At the moment, when a node joins the network, it will subscribe to the `topics` topic. This means that it will 
receive a message each time there is a new file list to store, and it should have a record of every file posted to 
the network from that point on.

This is not ideal, as we currently don't have any way of syncing nodes to the current state of the network, they 
will only replicate future state. 

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