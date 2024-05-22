package Client

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Client struct {
	conn net.Conn
}

func (c *Client) SendCredentials(option, username, password string) {
	var err error
	c.conn, err = net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error al conectar al servidor:", err)
		return
	}

	fmt.Println("Conexion establecida con el servidor.")

	reader := bufio.NewReader(c.conn)
	for i := 0; i < 3; i++ {
		response, _ := reader.ReadString('\n')
		fmt.Print(response)
	}
	response, _ := reader.ReadString(':')
	fmt.Print(response)

	c.conn.Write([]byte(option + "\n"))

	response, _ = reader.ReadString(':')
	fmt.Print(response)
	c.conn.Write([]byte(username + "\n"))

	response, _ = reader.ReadString(':')
	fmt.Print(response)
	c.conn.Write([]byte(password + "\n"))

	response, _ = bufio.NewReader(c.conn).ReadString('\n')
	fmt.Print(response)

	if strings.Contains(response, "Bienvenido") {
		go c.ReceiveMessages()
		c.openChooseWindow()
		c.loginWindow.Hide()
	}
}

func (c *Client) ReceiveMessages() {
	reader := bufio.NewReader(c.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer mensaje del servidor:", err)
			return
		}
		c.handleServerMessage(message)
	}
}

func (c *Client) UpdateChatMessage(message string) {
	c.messageDisplay.SetText(c.messageDisplay.Text + message)
}

func (c *Client) sendAnswerToServer(answer string) {
	c.conn.Write([]byte("ANSWER " + answer + "\n"))
}

func (c *Client) UpdateAnswerMessage(message string) {
	if c.messageDisplay != nil {
		c.messageDisplay.SetText(message)
	}
}

func (c *Client) HandleServerMessage(message string) {
	fmt.Println("Mensaje leído para el Handle: ", message)
	if strings.HasPrefix(message, "QUESTION:") {
		options := strings.Split(strings.TrimPrefix(message, "QUESTION:"), "\n")
		c.UpdateQuestion(options[0])
		c.UpdateQuestion(options[1:])
	} else if strings.TrimSpace(message) == "CORRECT" {
		c.updateAnswerMessage("Respuesta Correcta")
	} else if strings.TrimSpace(message) == "INCORRECT" {
		c.updateAnswerMessage("Respuesta Incorrecta")
	} else {
		c.UpdateChatMessage(message)
	}
}

func (c *Client) SendChatMessage(message string) {
	c.conn.Write([]byte(message + "\n"))
}

func InitUser() {
	myApp := app.New()

	client := &Client{myApp: myApp}

	welcomeWindow := myApp.NewWindow("Preguntados")
	welcomeWindow.SetContent(container.NewVBox(
		widget.NewLabel("Bienvenido a Preguntados"),
		widget.NewButton("Iniciar sesion", func() {
			client.ShowLoginWindow()
			welcomeWindow.Hide()
		}),
	))
	welcomeWindow.Resize(fyne.NewSize(400, 400))
	welcomeWindow.Show()

	myApp.Run()

}
