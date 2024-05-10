package main

import (
	"fmt"
	"net"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

func main() {
	conexion, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)

	if err != nil {
		panic(err)
	}

	_, err = conexion.Write([]byte("Buenas! bienvenido."))
	buffer := make([]byte, 1024)
	mLen, err := conexion.Read(buffer)

	if err != nil {
		fmt.Println("Error de lectura:", err.Error())
	}
	fmt.Println("Recibido:", string(buffer[:mLen]))
	defer conexion.Close()
}
