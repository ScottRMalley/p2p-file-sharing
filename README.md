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

## Improvements
### Features
#### Backsyncing
At the moment, when a node joins the network, it will subscribe to the `topics` topic. This means that it will 
receive a message each time there is a new file list to store, and it should have a record of every file posted to 
the network from that point on.

This is not ideal, as we currently don't have any way of syncing nodes to the current state of the network, they 
will only replicate future state. 