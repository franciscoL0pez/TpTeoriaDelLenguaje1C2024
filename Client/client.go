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
	conn           net.Conn
	myApp          fyne.App
	loginWindow    fyne.Window
	users          fyne.Window
	messageDisplay *widget.Label
	questionLabel  *widget.Label
	answerEntry    *widget.Entry
	optionsButtons []*widget.Button
}

func (c *Client) sendCredentials(option, username, password string) {
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
		go c.receiveMessages()
		c.openChooseWindow()
		c.loginWindow.Hide()
	}
}

func (c *Client) receiveMessages() {
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

func (c *Client) updateChatMessage(message string) {
	c.messageDisplay.SetText(c.messageDisplay.Text + message)
}

func (c *Client) updateQuestion(question string) {
	c.questionLabel.SetText(strings.TrimPrefix(question, "QUESTION:"))
}

func (c *Client) sendAnswerToServer(answer string) {
	c.conn.Write([]byte("ANSWER " + answer + "\n"))
}

func (c *Client) updateAnswerMessage(message string) {
	if c.messageDisplay != nil {
		c.messageDisplay.SetText(message)
	}
}

func (c *Client) handleServerMessage(message string) {
	fmt.Println("Mensaje leído para el Handle: ", message)
	if strings.HasPrefix(message, "QUESTION:") {
		options := strings.Split(strings.TrimPrefix(message, "QUESTION:"), "\n")
		c.updateQuestion(options[0])
		c.updateOptions(options[1:])
	} else if strings.TrimSpace(message) == "CORRECT" {
		c.updateAnswerMessage("Respuesta Correcta")
	} else if strings.TrimSpace(message) == "INCORRECT" {
		c.updateAnswerMessage("Respuesta Incorrecta")
	} else {
		c.updateChatMessage(message)
	}
}

func (c *Client) updateOptions(options []string) {
	for _, button := range c.optionsButtons {
		button.Hide()
	}

	c.optionsButtons = make([]*widget.Button, len(options))
	for i, option := range options {
		c.optionsButtons[i] = widget.NewButton(option, func(option string) func() {
			return func() {
				c.sendAnswerToServer(option)
			}
		}(option))
		c.optionsButtons[i].Show()
	}
}

func (c *Client) openGameWindow() {
	gameWindow := c.myApp.NewWindow("Game")
	gameWindow.Resize(fyne.NewSize(400, 400))

	c.questionLabel = widget.NewLabel("")
	c.messageDisplay = widget.NewLabel("")

	c.answerEntry = widget.NewEntry()
	checkButton := widget.NewButton("Check", func() {
		c.sendAnswerToServer(c.answerEntry.Text)
		c.answerEntry.SetText("")
	})

	gameWindow.SetContent(container.NewVBox(
		c.questionLabel,
		c.messageDisplay,
		container.NewGridWithColumns(2,
			c.answerEntry,
			checkButton,
		),
	))

	c.conn.Write([]byte("GET_QUESTION\n"))

	gameWindow.Show()
}

func (c *Client) showLoginWindow() {
	c.loginWindow = c.myApp.NewWindow("Login")

	optionLabel := widget.NewLabel("Elige una opcion:")
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginButton := widget.NewButton("Login", func() {
		c.sendCredentials("1", usernameEntry.Text, passwordEntry.Text)
	})

	registerButton := widget.NewButton("Register", func() {
		c.sendCredentials("2", usernameEntry.Text, passwordEntry.Text)
		usernameEntry.SetText("")
		passwordEntry.SetText("")
	})

	c.loginWindow.SetContent(container.NewVBox(
		optionLabel,
		usernameEntry,
		passwordEntry,
		container.NewHBox(
			loginButton,
			registerButton,
		),
	))
	c.loginWindow.Resize(fyne.NewSize(400, 400))
	c.loginWindow.Show()
}

func (c *Client) openChooseWindow() {
	chooseWindow := c.myApp.NewWindow("Elegir")
	chooseWindow.SetContent(container.NewVBox(
		widget.NewButton("Chat", func() {
			c.openChatWindow()
			chooseWindow.Hide()
		}),
		widget.NewButton("Jugar", func() {
			c.openGameWindow()
			chooseWindow.Hide()
		}),
	))
	chooseWindow.Resize(fyne.NewSize(400, 400))
	chooseWindow.Show()
}

func (c *Client) openChatWindow() {
	chatWindow := c.myApp.NewWindow("Chat")
	messageEntry := widget.NewEntry()
	sendButton := widget.NewButton("Send", func() {
		message := messageEntry.Text
		c.sendChatMessage(message)
		messageEntry.SetText("")
	})

	c.messageDisplay = widget.NewLabel("")

	chatWindow.SetContent(container.NewVBox(
		c.messageDisplay,
		container.NewBorder(nil, nil, nil, sendButton, messageEntry),
	))

	chatWindow.Resize(fyne.NewSize(400, 400))
	chatWindow.Show()
}

func (c *Client) sendChatMessage(message string) {
	c.conn.Write([]byte(message + "\n"))
}

func InitUser() {
	myApp := app.New()

	client := &Client{myApp: myApp}

	welcomeWindow := myApp.NewWindow("Preguntados")
	welcomeWindow.SetContent(container.NewVBox(
		widget.NewLabel("Bienvenido a Preguntados"),
		widget.NewButton("Iniciar sesion", func() {
			client.showLoginWindow()
			welcomeWindow.Hide()
		}),
	))
	welcomeWindow.Resize(fyne.NewSize(400, 400))
	welcomeWindow.Show()

	myApp.Run()

}
