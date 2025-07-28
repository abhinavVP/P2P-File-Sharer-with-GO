# GoP2P File Sharer
A command-line peer-to-peer (P2P) file sharing application built from the ground up in Go. This system allows users on the same local network to discover each other and transfer files directly without the need for a central server.

# About The Project
This project is a practical implementation of core networking and concurrency concepts in Go. It demonstrates a decentralized approach to file sharing, where each running instance of the application acts as both a client and a server.

The system uses a dual-protocol approach:

UDP Broadcasts for efficient, low-overhead peer discovery on the local network.

TCP Connections for reliable, error-checked communication for sending commands and streaming file data.

# Features
Automatic Peer Discovery: Nodes automatically find and maintain a list of other active peers on the LAN.

Direct Peer Communication: A custom JSON-based protocol runs over TCP for sending commands.

File Indexing: Each peer indexes files in a local shared_files directory upon startup.

Remote File Listing: Peers can request and receive a list of shareable files from any other peer.

Direct File Transfers: Securely download files directly from another peer by streaming the raw file data.

Interactive CLI: A simple command-line interface to interact with the network (peers, list, download).

# How to use
Pull the repo and build it using the 'go build' tool.
After building the executable, run it in two different machines (can use VMs on Bridged Mode for lack of another physical machine). Ensure that they are both running on the same local network so that the peers can discover each other.

# Note
This application relies on network broadcasts (UDP) and direct connections (TCP). Your operating system's firewall will likely block these connections by default.

For the application to work correctly, you must create a firewall exception for the compiled executable.
