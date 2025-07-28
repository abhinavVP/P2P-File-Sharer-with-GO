package p2p

import (
	"sync"
	"time"
)

const DiscoveryPort = 8829
const TCPPort = 8830
const BroadcastInterval = 5
const PeerTimeout = 15
const SharedDir = "shared_files"

type Peer struct {
	Address  string
	LastSeen time.Time
}

type PeerStore struct {
	Mu    sync.Mutex
	Peers map[string]Peer
}

type Message struct {
	Command string
	Payload string
}

var LocalFileIndex map[string]struct{}
var IndexMutex = &sync.RWMutex{}
