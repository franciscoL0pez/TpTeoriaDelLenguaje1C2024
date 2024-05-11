package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
)

func authenticate(username, password string) bool {
	file, err := os.Open("users.csv")
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return false
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error al leer el archivo CSV:", err)
		return false
	}

	for _, record := range records {
		if record[0] == username && record[1] == password {
			return true
		}
	}

	return false
}

func register(username, password string) error {
	file, err := os.OpenFile("users.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.Write([]string{username, password})
	if err != nil {
		return err
	}

	writer.Flush()

	return nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Nueva conexion establecida:", conn.RemoteAddr())

	conn.Write([]byte("Bienvenido al servidor.\n"))
	conn.Write([]byte("1. Iniciar sesión\n"))
	conn.Write([]byte("2. Registrarse\n"))
	conn.Write([]byte("Ingrese opcion:"))

	reader := bufio.NewReader(conn)
	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	switch option {
	case "1": // Iniciar sesión
		conn.Write([]byte("Ingrese su nombre de usuario:"))
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		conn.Write([]byte("Ingrese su contraseña:"))
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		if authenticate(username, password) {
			fmt.Println("Usuario autenticado:", username)
			conn.Write([]byte("Bienvenido, " + username + "!\n"))
		} else {
			fmt.Println("Autenticacion fallida para el usuario:", username)
			conn.Write([]byte("Autenticacion fallida. Cierre de la conexion.\n"))
			return
		}
	case "2": // Registrarse
		conn.Write([]byte("Ingrese un nombre de usuario para registrarse:"))
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		conn.Write([]byte("Ingrese una contraseña para registrarse:"))
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		if err := register(username, password); err != nil {
			fmt.Println("Error al registrar el usuario:", err)
			conn.Write([]byte("Error al registrar el usuario.\n"))
			return
		}

		fmt.Println("Usuario registrado:", username)
		conn.Write([]byte("Usuario registrado con exito.\n"))
	default:
		conn.Write([]byte("Opción no valida. Cierre de la conexion.\n"))
		return
	}
}

func main() {
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
		go handleConnection(conn)
	}
}
