package tcptogo

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

// var mutex sync.Mutex

type Server struct {
	Address  string
	Capacity int

	listener          net.Listener
	queue             chan net.Conn
	wg                sync.WaitGroup
	maxTimeInactive   int
	activeConnections int
	inactive          bool
}

func New(address string) (server *Server) {
	// for simulating loads
	capacity := 1
	maxQueued := 1

	return &Server{
		Address:           address,
		Capacity:          capacity,
		queue:             make(chan net.Conn, maxQueued),
		maxTimeInactive:   10, // in seconds
		activeConnections: 0,
		inactive:          true,
	}
}

func (server *Server) Start() {
	var err error
	server.listener, err = net.Listen("tcp", server.Address)
	if err != nil {
		fmt.Println("Error setting up tcp listener:", err)
		return
	}
	fmt.Printf("[Server] TCP server listening on port %s\n", server.Address)

	defer func() {
		server.listener.Close()
		defer fmt.Println("\nGoodbye.")
	}()

	server.inactive = false
	server.open()
}

func (server *Server) open() {
	defer fmt.Println("[Server] Processes terminated.")

	server.wg.Add(3)
	go server.consume()
	go server.accept()
	go server.timeout()

	server.wg.Wait()
}

// Handle queued connections if the server has the capacity to.
func (server *Server) consume() {
	defer server.wg.Done()
	// For simulating server capacity, i.e. maximum concurrent processes
	for {
		if server.inactive {
			fmt.Println("[Server] Terminating consumer...")
			return
		}

		if server.activeConnections < server.Capacity && len(server.queue) > 0 {
			conn := <-server.queue
			fmt.Printf("[Server] Available capacity, processing connection from queue: %s\n", conn.RemoteAddr())
			server.activeConnections++
			server.wg.Add(1)
			go server.handleConnection(conn)
		}
	}

	// Actual consumer should look like this
	// for conn := range server.queue {
	// 	fmt.Printf("[Server] Available capacity, processing connection from queue: %s\n", conn.RemoteAddr())
	// 	server.activeConnections++
	// server.wg.Add(1)
	// 	go server.handleConnection(conn)
	// }
}

// Accept incoming TCP connections, queue connections when capacity has been reached.
func (server *Server) accept() {
	defer server.wg.Done()

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				fmt.Println("[Server] Listener closed, terminating accept loop...")
				return
			}
			fmt.Println("Error accepting TCP connection: ", err.Error())
			continue
		}

		fmt.Printf("[Server] New connection received: %s\n", conn.RemoteAddr())

		if server.inactive {
			fmt.Println("[Server] Server inactive, terminating accept loop...")
			return
		}

		if len(server.queue) == cap(server.queue) {
			fmt.Printf("[Server] Queue is currently full, cannot process %s\n", conn.RemoteAddr())
			conn.Close()
			continue
		}

		server.queue <- conn
		fmt.Printf("[Server] Queued %s\n", conn.RemoteAddr())
	}
}

func (server *Server) timeout() {
	defer server.wg.Done()

	var timeInactive int = 0
	var dots string = "."

	for {
		time.Sleep(1 * time.Second)

		if server.activeConnections == 0 {
			timeInactive++
		} else {
			timeInactive = 0
			dots = "."
		}

		if timeInactive%2 != 0 {
			fmt.Printf("[Server] Waiting for connections%s\n", dots)
			dots += "."
		}

		if timeInactive == server.maxTimeInactive {
			fmt.Printf("[Server] Server inactive for %ds, shutting down.\n", server.maxTimeInactive)
			server.flush()
			return
		}
	}
}

func (server *Server) flush() {
	defer close(server.queue)
	server.inactive = true
	server.listener.Close()

	fmt.Println("[Server] Flushing connections...")
	for range len(server.queue) {
		conn := <-server.queue

		fmt.Printf("[Server] Force closing %s\n", conn.RemoteAddr())
		conn.Close()
	}
	fmt.Println("[Server] Connections closed.")
}

func (server *Server) handleConnection(conn net.Conn) {
	fmt.Printf("[%s] Processing connection...\n", conn.RemoteAddr())

	buffer := make([]byte, 1024)

	// Read buffer
	if _, err := conn.Read(buffer); err != nil {
		log := fmt.Sprintf("Error reading buffer: %v", err)
		panic(log)
	}

	slowQuery()

	// fmt.Printf(
	// 	"\n*******Start buffer*******\n%s********End buffer********\n\n",
	// 	string(buffer),
	// )

	// Write buffer
	if _, err := conn.Write([]byte("Message Received.")); err != nil {
		log := fmt.Sprintf("Error writing buffer: %v", err)
		panic(log)
	}

	defer func() {
		fmt.Printf("[%s] Connection closed.\n", conn.RemoteAddr())
		// time.Sleep(500 * time.Millisecond)
		server.activeConnections--
		conn.Close()
		server.wg.Done()
	}()
}

// func (server *Server) updateActiveConnections(step int) {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	server.activeConnections += step
// }

func slowQuery() {
	executionTime := rand.Intn(4) + 3
	time.Sleep(time.Duration(executionTime) * time.Second)
}
