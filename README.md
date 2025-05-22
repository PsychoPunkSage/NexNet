# NexNet - Distributed Encrypted File Storage System

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 🚀 Overview

NexNet is a **peer-to-peer distributed file storage system** built in Go that demonstrates advanced systems programming concepts including cryptography, concurrent networking, and distributed consensus. The system automatically replicates encrypted files across multiple nodes with content-addressable storage.

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   NexNet P2P Network                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Node A    │◄──►│   Node B    │◄──►│   Node C    │  │
│  │  :3000      │    │  :4000      │    │  :5000      │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│         │                   │                   │       │
│         ▼                   ▼                   ▼       │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Storage   │    │   Storage   │    │   Storage   │  │
│  │    Layer    │    │    Layer    │    │    Layer    │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│                                                         │
└─────────────────────────────────────────────────────────┘

  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
  │ TCP Transport│    │ Cryptography │    │   Storage    │
  │   • Peers    │    │   • AES-CTR  │    │   • CAS      │
  │   • Handshake│    │   • SHA-1    │    │   • PathKey  │
  │   • Streaming│    │   • Random IV│    │   • Chunking │
  └──────────────┘    └──────────────┘    └──────────────┘
```

## 🔧 Core Architecture Components

### 1. **P2P Transport Layer** (`p2p/`)
- **TCP-based peer-to-peer communication**
- **Concurrent connection handling** with goroutines
- **Custom RPC protocol** with message/stream differentiation
- **Peer lifecycle management** with proper cleanup

### 2. **Cryptographic Security** (`cryptography/`)
- **AES-256-CTR encryption** for all file data
<details>
<summary>
Why?
</summary>

Key Benefits:

- ✅ Stream encryption (no buffering entire file)
- ✅ Random access (can decrypt any part without decrypting whole file)
- ✅ Parallel processing (multiple chunks simultaneously)
- ✅ No padding required (unlike CBC mode)

</details>

- **Random IV generation** for each encryption operation
- **32KB streaming chunks** for memory-efficient processing
- **XORKeyStream** for real-time encrypt/decrypt

### 3. **Content-Addressable Storage** (`storage/`)
- **SHA-1 hash-based file organization**
- **Hierarchical directory structure** (5-char blocks)
- **Collision-resistant addressing**
- **Efficient file deduplication**

### 4. **Distributed File Server** (`server/`)
- **Multi-node replication** with automatic discovery
- **Network-wide file retrieval** when not available locally
- **Concurrent file operations** using Go channels
- **Bootstrap node connectivity**

## 🎯 Advanced Go Concepts Demonstrated

### **Concurrency & Goroutines**
```go
// Concurrent peer handling
go t.handleConn(conn, false)

// Bootstrap network connections
go func(addr string) {
    if err := s.Transport.Dial(addr); err != nil {
        log.Println("Dial error: ", err)
    }
}(addr)
```

### **I/O Streaming Patterns**
```go
// TeeReader for simultaneous read/write
tee := io.TeeReader(r, fileBuffer)

// MultiWriter for broadcasting to multiple peers
peers := []io.Writer{}
for _, peer := range s.peers {
    peers = append(peers, peer)
}
mw := io.MultiWriter(peers...)
```

<details>
<summary>
Concept
</summary>

## **The Core Concept Behind Zero-Copy I/O**

**Traditional Systems**: Think of it like a **photocopier workflow**
- Read document → Make copy 1 → Put away original → Get original again → Make copy 2...
- Each operation requires getting the original document again

**NexNet's Approach**: Think of it like a **water pipe with multiple outlets**
- Water flows ONCE through the main pipe
- Multiple taps can draw from the same flow simultaneously  
- No need to "re-flow" the water for each tap

**Technical Magic**:
- `TeeReader` = The "pipe splitter" (one input → multiple outputs)
- `MultiWriter` = The "broadcast valve" (one write → multiple destinations)
- Result = **Single data read feeds ALL operations simultaneously**

This is why your approach is genuinely sophisticated - you're eliminating the fundamental inefficiency of traditional file distribution systems!

</details>

### **Memory-Efficient Encryption**
```go
// 32KB chunked processing prevents memory overflow
buf := make([]byte, 32*1024)
for {
    n, err := src.Read(buf)
    if n > 0 {
        stream.XORKeyStream(buf, buf[:n])  // In-place encryption
        dst.Write(buf[:n])
    }
}
```

### **Content-Addressable Storage**
```go
func CASPathTransformFunc(key string) PathKey {
    hash := sha1.Sum([]byte(key))
    hashStr := hex.EncodeToString(hash[:])
    
    // Create hierarchical path: ab/cd/ef/gh/ij/abcdefghij...
    blocksize := 5
    sliceLen := len(hashStr) / blocksize
    paths := make([]string, sliceLen)
    
    for i := 0; i < sliceLen; i++ {
        from, to := i*blocksize, i*blocksize+blocksize
        paths[i] = hashStr[from:to]
    }
    
    return PathKey{
        PathName: strings.Join(paths, "/"),
        Filename: hashStr,
    }
}
```

## 🔐 Cryptographic Implementation

### **AES-CTR Encryption**
- **Counter Mode**: Enables streaming encryption without padding
- **Random IV**: 16-byte initialization vector for each file
- **Key Derivation**: 32-byte random keys for AES-256
- **Stream Cipher**: XORKeyStream for real-time processing

### **Why 32KB Chunks?**
1. **Memory Efficiency**: Prevents loading entire files into RAM
2. **Network Optimization**: Optimal TCP packet sizing
3. **Concurrent Processing**: Enables pipeline processing
4. **Cache Friendly**: Fits in CPU L2/L3 cache

## 🌐 Distributed System Features

### **Network Topology**
- **Mesh Network**: Each node can connect to multiple peers
- **Bootstrap Discovery**: New nodes discover network through bootstrap nodes
- **Automatic Replication**: Files are automatically replicated across connected peers
- **Fault Tolerance**: System continues operating with node failures

### **File Operations**
- **Store**: Encrypt and replicate files across network
- **Retrieve**: Fetch files from any node in the network
- **Delete**: Coordinate file deletion across all nodes
- **Deduplication**: Same content stored only once per node

## 🚀 Quick Start

### **Build & Run**
```bash
# Build the binary
make build

# Run the demo
make run

# Run tests
make test
```

### **Demo Scenario**
The main.go demonstrates a 3-node network:
```go
s := makeServer(":3000", "")           // Bootstrap node
s1 := makeServer(":4000", ":3000")     // Connects to 3000
s2 := makeServer(":5000", ":4000", ":3000") // Connects to both
```

## 📁 Project Structure

```
.
├── bin/                    # Compiled binaries
├── cryptography/          # Encryption/decryption logic
│   ├── crypto.go          # AES-CTR implementation
│   └── crypto_test.go
├── p2p/                   # Peer-to-peer networking
│   ├── encoding.go        # Message encoding/decoding
│   ├── handshake.go       # Peer handshake logic
│   ├── message.go         # RPC message definitions
│   ├── tcp_transport.go   # TCP transport implementation
│   └── transport.go       # Transport interface
├── server/                # Distributed file server
│   ├── server.go          # Main server logic
│   └── server_test.go
├── storage/               # Content-addressable storage
│   ├── store.go           # Storage implementation
│   └── store_test.go
├── main.go               # Demo application
├── Makefile             # Build configuration
└── README.md
```

## 🔬 Technical Highlights

### **1. Advanced Concurrency**
- **Channel-based communication** for RPC handling
- **WaitGroup synchronization** for stream operations
- **Mutex protection** for shared peer state
- **Graceful shutdown** with context cancellation

### **2. Cryptographic Streaming**
- **CTR mode encryption** for parallel processing
- **IV prepending** for secure key reuse
- **In-place XOR operations** for memory efficiency
- **Binary encoding** for network transmission

### **3. Content-Addressable Storage**
- **Collision-resistant hashing** with SHA-1
- **Hierarchical file organization** for filesystem efficiency
- **Automatic deduplication** through hash-based naming
- **Scalable directory structure** preventing single-directory limits

### **4. Distributed Coordination**
- **Gossip-style replication** for file distribution
- **Network-wide search** for file retrieval
- **Coordinated deletion** across all nodes
- **Bootstrap-based discovery** for network joining

## 🛠️ Advanced Features

- **Real-time file streaming** with encryption
- **Automatic peer discovery** and connection management
- **Fault-tolerant file retrieval** from multiple sources
- **Memory-efficient processing** of large files
- **Hierarchical storage organization** for scalability

## 🎓 Learning Outcomes

This project demonstrates mastery of:
- **Advanced Go concurrency patterns**
- **Cryptographic protocol implementation**
- **Distributed systems design**
- **Network programming with TCP**
- **Content-addressable storage systems**
- **Streaming I/O operations**
- **P2P network architecture**

---

*NexNet showcases production-ready distributed systems engineering with Go, emphasizing security, scalability, and concurrent processing.*