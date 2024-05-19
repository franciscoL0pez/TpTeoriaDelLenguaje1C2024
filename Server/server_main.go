// ProjectRoot/Server/server_main.go
package main

import (
	"fmt"
	"net"
)

func main() {
	loadQuestions()
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Servidor escuchando en el puerto 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error al aceptar la conexion:", err)
			continue
		}
		clients[conn] = true
		go handleConnection(conn)
	}
}
