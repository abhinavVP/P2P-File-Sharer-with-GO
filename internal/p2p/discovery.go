package p2p

import (
	"fmt"
	"log"
	"net"
	"time"
)

func NewPeerStore() *PeerStore {
	return &PeerStore{Peers: make(map[string]Peer)}
}

func (ps *PeerStore) AddOrUpdate(addr string) {
	ps.Mu.Lock()
	defer ps.Mu.Unlock()
	if _, exists := ps.Peers[addr]; exists {
		ps.Peers[addr] = Peer{Address: addr, LastSeen: time.Now()}
		return
	}
	ps.Peers[addr] = Peer{Address: addr, LastSeen: time.Now()}
	log.Printf("!!! New peer discovered: %s !!!", addr)
}

func (ps *PeerStore) GetPeers() []string {
	ps.Mu.Lock()
	defer ps.Mu.Unlock()
	var activePeers []string
	for addr, peer := range ps.Peers {
		if time.Since(peer.LastSeen) < (PeerTimeout * time.Second) {
			activePeers = append(activePeers, addr)
		} else {
			delete(ps.Peers, addr)
			log.Printf("Removed peer due to timeout: %s", addr)
		}
	}
	return activePeers
}

func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func BroadcastSelf() {
	broadcastAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", DiscoveryPort))
	conn, _ := net.ListenPacket("udp4", ":0")
	defer conn.Close()
	myIP, _ := GetLocalIP()
	message := []byte(fmt.Sprintf("helloo:%s", myIP))
	ticker := time.NewTicker(BroadcastInterval * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		conn.(*net.UDPConn).WriteToUDP(message, broadcastAddr)
	}
}

func DiscoverPeers(ps *PeerStore, myIP string) {
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", DiscoveryPort))
	conn, _ := net.ListenUDP("udp4", addr)
	defer conn.Close()
	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}
		peerAddr := remoteAddr.IP.String()
		if peerAddr == myIP {
			continue
		}
		message := string(buffer[:n])
		if len(message) > 7 && message[:7] == "helloo:" {
			ps.AddOrUpdate(peerAddr)
		}
	}
}
