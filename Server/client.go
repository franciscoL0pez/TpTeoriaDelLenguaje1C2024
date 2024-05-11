package main

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

var (
	conn           net.Conn
	myApp          fyne.App
	loginWindow    fyne.Window
	messageDisplay *widget.Label
)

func main() {
	myApp = app.New()

	welcomeWindow := myApp.NewWindow("Preguntados")
	welcomeWindow.SetContent(container.NewVBox(
		widget.NewLabel("Bienvenido a Preguntados"),
		widget.NewButton("Iniciar sesion", func() {
			showLoginWindow(myApp)
			welcomeWindow.Hide()
		}),
	))
	welcomeWindow.Resize(fyne.NewSize(400, 400))
	welcomeWindow.Show()

	myApp.Run()
}

func showLoginWindow(app fyne.App) {
	loginWindow = app.NewWindow("Login")

	optionLabel := widget.NewLabel("Elige una opcion:")
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginButton := widget.NewButton("Login", func() {
		sendCredentials("1", usernameEntry.Text, passwordEntry.Text)
	})

	registerButton := widget.NewButton("Register", func() {
		sendCredentials("2", usernameEntry.Text, passwordEntry.Text)
		usernameEntry.SetText("")
		passwordEntry.SetText("")
	})

	loginWindow.SetContent(container.NewVBox(
		optionLabel,
		usernameEntry,
		passwordEntry,
		container.NewHBox(
			loginButton,
			registerButton,
		),
	))
	loginWindow.Resize(fyne.NewSize(400, 400))
	loginWindow.Show()
}

func sendCredentials(option, username, password string) {
	var err error
	conn, err = net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error al conectar al servidor:", err)
		return
	}

	fmt.Println("Conexion establecida con el servidor.")

	reader := bufio.NewReader(conn)
	for i := 0; i < 3; i++ {
		response, _ := reader.ReadString('\n')
		fmt.Print(response)
	}
	response, _ := reader.ReadString(':')
	fmt.Print(response)

	conn.Write([]byte(option + "\n"))

	response, _ = reader.ReadString(':')
	fmt.Print(response)
	conn.Write([]byte(username + "\n"))

	response, _ = reader.ReadString(':')
	fmt.Print(response)
	conn.Write([]byte(password + "\n"))

	response, _ = bufio.NewReader(conn).ReadString('\n')
	fmt.Print(response)

	if strings.Contains(response, "Bienvenido") {
		openChatWindow()
		go receiveChatMessage()
		loginWindow.Hide()
	}
}

func openChatWindow() {
	chatWindow := myApp.NewWindow("Chat")
	messageEntry := widget.NewEntry()
	sendButton := widget.NewButton("Send", func() {
		message := messageEntry.Text
		sendChatMessage(message)
		messageEntry.SetText("")
	})

	messageDisplay = widget.NewLabel("")

	chatWindow.SetContent(container.NewVBox(
		messageDisplay,
		container.NewBorder(nil, nil, nil, sendButton, messageEntry),
	))

	chatWindow.Resize(fyne.NewSize(400, 400))
	chatWindow.Show()
}

func receiveChatMessage() {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer mensaje:", err)
			return
		}
		updateChatMessage(message)
	}
}

func sendChatMessage(message string) {
	conn.Write([]byte(message + "\n"))
}

func updateChatMessage(message string) {
	messageDisplay.SetText(messageDisplay.Text + message)
}
