package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"p2p-fs/internal/p2p"
	"strings"
)

func main() {
	myIP, err := p2p.GetLocalIP()
	if err != nil {
		log.Fatalf("Could not get local IP, check network connection. Error: %v", err)
	}
	log.Printf("Starting node for fs-sys... My IP is %s", myIP)

	p2p.BuildFileIndex()
	peerStore := p2p.NewPeerStore()

	go p2p.DiscoverPeers(peerStore, myIP)
	go p2p.BroadcastSelf()
	go p2p.StartTCPServer()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		parts := strings.Split(text, " ")
		cmd := parts[0]

		switch cmd {
		case "peers":
			peers := peerStore.GetPeers()
			log.Printf("--- Currently %d peers ---", len(peers))
			for _, peer := range peers {
				log.Println("->", peer)
			}
			log.Println("------------------------")
		case "ping":
			if len(parts) < 2 {
				log.Println("Usage: ping <peer_ip>")
				continue
			}
			go p2p.SendPing(parts[1])
		case "list":
			if len(parts) < 2 {
				log.Println("Usage: list <peer_ip>")
				continue
			}
			go p2p.RequestFilesList(parts[1])
		case "download":
			if len(parts) < 3 {
				log.Println("Usage: download <peer_ip> <filename>")
				continue
			}
			go p2p.RequestFileDownload(parts[1], parts[2])
		case "exit":
			return
		default:
			if cmd != "" {
				log.Println("Unknown command. Available: peers, ping, list, download, exit")
			}
		}
	}
}
