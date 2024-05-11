package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error al conectar al servidor:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Conexión establecida con el servidor.")

	reader := bufio.NewReader(conn)
	for i := 0; i < 3; i++ {
		response, _ := reader.ReadString('\n')
		fmt.Print(response)
	}
	response, _ := reader.ReadString(':')
	fmt.Print(response)

	reader_os := bufio.NewReader(os.Stdin)
	option, _ := reader_os.ReadString('\n')
	option = strings.TrimSpace(option)
	conn.Write([]byte(option + "\n"))

	switch option {
	case "1": // Iniciar sesión
		response, _ = reader.ReadString(':')
		fmt.Print(response)
		username, _ := reader_os.ReadString('\n')
		username = strings.TrimSpace(username)
		conn.Write([]byte(username + "\n"))

		response, _ = reader.ReadString(':')
		fmt.Print(response)
		password, _ := reader_os.ReadString('\n')
		password = strings.TrimSpace(password)
		conn.Write([]byte(password + "\n"))

		response, _ = bufio.NewReader(conn).ReadString('\n')
		fmt.Print(response)
	case "2": // Registrarse
		response, _ = reader.ReadString(':')
		fmt.Print(response)
		username, _ := reader_os.ReadString('\n')
		username = strings.TrimSpace(username)
		conn.Write([]byte(username + "\n"))

		response, _ = reader.ReadString(':')
		fmt.Print(response)
		password, _ := reader_os.ReadString('\n')
		password = strings.TrimSpace(password)
		conn.Write([]byte(password + "\n"))

		response, _ = bufio.NewReader(conn).ReadString('\n')
		fmt.Print(response)
	default:
		fmt.Println("Opcion no válida.")
		return
	}
}
