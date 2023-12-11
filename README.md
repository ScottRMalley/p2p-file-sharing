# P2P File Challenge

## Description

## Improvements
### Features
#### Backsyncing
At the moment, when a node joins the network, it will subscribe to the `topics` topic. This means that it will 
receive a message each time there is a new file list to store, and it should have a record of every file posted to 
the network from that point on.

This is not ideal, as we currently don't have any way of syncing nodes to the current state of the network, they 
will only replicate future state. 