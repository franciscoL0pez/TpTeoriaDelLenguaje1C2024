package Client

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Client struct {
	Conn net.Conn
	Name string
}

func (c *Client) SendCredentials(option, username, password string) error {
	var err error
	c.Conn, err = net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		return fmt.Errorf("error al conectar al servidor: %v", err)
	}

	fmt.Println("Conexion establecida con el servidor.")

	reader := bufio.NewReader(c.Conn)
	for i := 0; i < 3; i++ {
		response, _ := reader.ReadString('\n')
		fmt.Print(response)
	}
	response, _ := reader.ReadString(':')
	fmt.Print(response)

	c.Conn.Write([]byte(option + "\n"))

	response, _ = reader.ReadString(':')
	fmt.Print(response)
	c.Conn.Write([]byte(username + "\n"))

	response, _ = reader.ReadString(':')
	fmt.Print(response)
	c.Conn.Write([]byte(password + "\n"))

	response, _ = bufio.NewReader(c.Conn).ReadString('\n')
	fmt.Print(response)

	if strings.Contains(response, "Bienvenido") {
		c.Name = username
		return nil
	}
	return fmt.Errorf("invalid credentials")
}

func (c *Client) ReceiveMessages(handler func(string)) {
	reader := bufio.NewReader(c.Conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer mensaje del servidor:", err)
			return
		}
		handler(message)
	}
}

func (c *Client) SendMessage(message string) {

	fmt.Println("ENVIO: " + c.Name + ":" + message + "\n")
	c.Conn.Write([]byte(c.Name + " : " + message + "\n"))
}
