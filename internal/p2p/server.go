package p2p

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
)

func BuildFileIndex() {
	IndexMutex.Lock()
	defer IndexMutex.Unlock()
	LocalFileIndex = make(map[string]struct{})
	err := os.MkdirAll(SharedDir, os.ModePerm)
	if err != nil {
		log.Fatalf("couldn't create a shared directory: %v", err)
	}
	err = fs.WalkDir(os.DirFS("."), SharedDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			filename := d.Name()
			LocalFileIndex[filename] = struct{}{}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error scanning the shared directory: %v", err)
	}
	log.Printf("File index built. Found %d files.", len(LocalFileIndex))
}

func StartTCPServer() {
	listenerAddr := fmt.Sprintf(":%d", TCPPort)
	listener, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		log.Fatalf("not able to start tcp server: %v", err)
	}
	defer listener.Close()
	log.Printf("TCP server listening on %s...", listenerAddr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("accepted new connection from %s...", conn.RemoteAddr().String())

	decoder := json.NewDecoder(conn)
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		return
	}

	peerAddr := conn.RemoteAddr().String()
	switch msg.Command {
	case "PING":
		log.Printf("got a PING from connection %s...", peerAddr)
	case "LIST_FILES":
		handleListFiles(conn, peerAddr)
	case "DOWNLOAD_FILE":
		handleDownloadFile(conn, peerAddr, msg.Payload)
	default:
		log.Printf("got an unknown command from the connection: %s... ", msg.Command)
	}
}

func handleListFiles(conn net.Conn, peerAddr string) {
	log.Printf("got a files listing request from %s...", peerAddr)
	IndexMutex.RLock()
	var files []string
	for file := range LocalFileIndex {
		files = append(files, file)
	}
	IndexMutex.RUnlock()
	payloadBytes, _ := json.Marshal(files)
	respMsg := Message{
		Command: "LIST_FILES_RESP",
		Payload: string(payloadBytes),
	}
	encoder := json.NewEncoder(conn)
	encoder.Encode(&respMsg)
}

func handleDownloadFile(conn net.Conn, peerAddr, filename string) {
	log.Printf("got a file download request from %s for %s", peerAddr, filename)
	filePath := filepath.Join(SharedDir, filename)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("error trying to open file %s : %v", filePath, err)
		return
	}
	defer file.Close()

	n, err := io.Copy(conn, file)
	if err != nil {
		log.Printf("error trying to send the file contents: %v", err)
		return
	}
	log.Printf("successfully sent %d bytes of data to %s", n, peerAddr)
}

func RequestFilesList(peer string) {
	targetAddr := fmt.Sprintf("%s:%d", peer, TCPPort)
	conn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("Could not connect to peer %s: %v", peer, err)
		return
	}
	defer conn.Close()
	reqMsg := Message{Command: "LIST_FILES"}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&reqMsg); err != nil {
		log.Printf("Failed to send LIST_FILES request: %v", err)
		return
	}
	log.Printf("Sent 'LIST_FILES' request to %s", peer)

	var respMsg Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&respMsg); err != nil {
		log.Printf("Failed to read response from peer %s: %v", peer, err)
		return
	}

	if respMsg.Command == "LIST_FILES_RESP" {
		var files []string
		if err := json.Unmarshal([]byte(respMsg.Payload), &files); err != nil {
			log.Printf("error decoding files list payload: %v", err)
			return
		}
		log.Printf("--- here are the files from %s ---", peer)
		for _, file := range files {
			log.Printf("  - %s", file)
		}
		log.Printf("-----------------------")
	} else {
		log.Printf("received unexpected response: %s", respMsg.Command)
	}
}

func RequestFileDownload(peer, filename string) {
	target := fmt.Sprintf("%s:%d", peer, TCPPort)
	conn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("could not connect to peer with ip %s: %v", target, err)
		return
	}
	defer conn.Close()

	msg := Message{Command: "DOWNLOAD_FILE", Payload: filename}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&msg); err != nil {
		log.Printf("error trying to send DOWNLOAD_FILE request: %v", err)
		return
	}
	log.Printf("Sent 'DOWNLOAD_FILE' request for '%s' to %s", filename, peer)

	savePath := filepath.Join(SharedDir, filename)
	newFile, err := os.Create(savePath)
	if err != nil {
		log.Printf("error trying to create file for download: %v", err)
		return
	}
	defer newFile.Close()

	n, err := io.Copy(newFile, conn)
	if err != nil {
		log.Printf("error trying to copy the contents of the requested file: %v", err)
	}
	log.Printf("successfully received %d bytes and contents saved at %s", n, savePath)
	BuildFileIndex()
}

func SendPing(peer string) {
	target := fmt.Sprintf("%s:%d", peer, TCPPort)
	conn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("could not connect to peer with ip %s: %v", target, err)
		return
	}
	defer conn.Close()

	msg := Message{Command: "PING"}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&msg); err != nil {
		log.Printf("error trying to encode message: %v", err)
		return
	}
	log.Printf("sent command PING to %s", peer)
}
