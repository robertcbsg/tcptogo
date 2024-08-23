package engine

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex

type Server struct {
	Address  string
	Capacity int

	listener          net.Listener
	queue             chan net.Conn
	activeConnections int
	inactive          bool
}

func New(address string, capacity int) (server *Server) {
	return &Server{
		Address:           address,
		Capacity:          capacity,
		queue:             make(chan net.Conn, capacity),
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
	fmt.Printf("TCP server listening on port %s...\n\n", server.Address)

	defer server.listener.Close()

	server.inactive = false
	server.Open()
}

func (server *Server) Open() {
	go server.Consume()
	go server.AcceptLoop()

	server.handleTimeout(10)
}

func (server *Server) handleTimeout(timeout int) {
	var inactiveTime int = 0
	for {
		time.Sleep(1 * time.Second)
		if server.readActiveConnections() == 0 && len(server.queue) == 0 {
			inactiveTime++
			// if inactiveTime == 1 {
			// 	fmt.Printf("[Server] Inactive for %d second...\n", inactiveTime)
			// } else {
			// 	fmt.Printf("[Server] Inactive for %d seconds...\n", inactiveTime)
			// }
		} else {
			inactiveTime = 0
		}

		if inactiveTime == timeout {
			fmt.Printf("[Server] Server inactive for %ss.", server.Tim)
			fmt.Println("[Server] Shutting down...")
			server.inactive = true
			server.listener.Close()
			time.Sleep(1 * time.Second) // let goroutines terminate first
			fmt.Println("Goodbye.")
			return
		}
	}
}

// Accept incoming TCP connections, queue connections when capacity has been reached.
func (server *Server) AcceptLoop() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				fmt.Println("[Server] Listener closed. Terminating accept loop...")
				return
			}
			fmt.Println("Error accepting TCP connection: ", err.Error())
			continue
		}

		fmt.Printf("[%s] New connection received.\n", conn.RemoteAddr())

		if server.readActiveConnections() >= server.Capacity {
			server.queue <- conn
			fmt.Printf("[%s] No capacity, queueing connection...\n", conn.RemoteAddr())
		} else {
			server.updateActiveConnections(1)
			go server.handleConnection(conn)
		}
	}
}

// Handle queue connections once server has freed up.
func (server *Server) Consume() {
	for {
		if server.inactive {
			fmt.Println("[Server] Terminating queue consumer...")
			return
		}

		if server.readActiveConnections() < server.Capacity && len(server.queue) > 0 {
			conn := <-server.queue
			fmt.Printf("[%s] Available capacity, processing connection from queue.\n", conn.RemoteAddr())
			server.updateActiveConnections(1)
			go server.handleConnection(conn)
		}

	}
}

func (server *Server) handleConnection(conn net.Conn) {
	fmt.Printf("[%s] Processing connection...\n", conn.RemoteAddr())

	// server.activeConnections++
	// server.updateActiveConnections(1)
	buffer := make([]byte, 1024)

	// Read buffer
	if _, err := conn.Read(buffer); err != nil {
		log := fmt.Sprintf("Error reading buffer: %v", err)
		panic(log)
	}

	slowQuery()

	fmt.Printf(
		"\n*******Start buffer*******\n%s********End buffer********\n\n",
		string(buffer),
	)

	// Write buffer
	if _, err := conn.Write([]byte("Message Received.")); err != nil {
		log := fmt.Sprintf("Error writing buffer: %v", err)
		panic(log)
	}

	defer func() {
		fmt.Printf("[%s] Connection closed.\n", conn.RemoteAddr())
		time.Sleep(500 * time.Millisecond)
		server.updateActiveConnections(-1)
		conn.Close()
	}()
}

func (server *Server) readActiveConnections() int {
	mutex.Lock()
	defer mutex.Unlock()

	return server.activeConnections
}

func (server *Server) updateActiveConnections(step int) {
	mutex.Lock()
	defer mutex.Unlock()

	server.activeConnections += step
}

func slowQuery() {
	time.Sleep(3 * time.Second)
}
