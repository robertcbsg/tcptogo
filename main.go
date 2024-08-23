package main

import (
	tcp "tcptogo/engine"
)

func main() {
	/* Initiate TCP connection */
	// fmt.Println("TCP client connecting to localhost:8080")
	// listener, err := net.Dial("tcp", ":8080")

	// if err != nil {
	// 	fmt.Println("Error initiating TCP connection:", err)
	// 	return
	// }
	// defer listener.Close()

	// listener.Read()

	server := tcp.New(":8080")

	server.Start()
}
