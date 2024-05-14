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
	users          fyne.Window
	messageDisplay *widget.Label
	questionLabel  *widget.Label
	answerEntry    *widget.Entry
	optionsButtons []*widget.Button
)

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
		go receiveMessages()
		openChooseWindow()
		loginWindow.Hide()
	}
}

func receiveMessages() {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer mensaje del servidor:", err)
			return
		}
		handleServerMessage(message)
	}
}

func updateChatMessage(message string) {
	messageDisplay.SetText(messageDisplay.Text + message)
}

func updateQuestion(question string) {
	questionLabel.SetText(strings.TrimPrefix(question, "QUESTION:"))
}

func sendAnswerToServer(answer string) {
	conn.Write([]byte("ANSWER " + answer + "\n"))
}

func updateAnswerMessage(message string) {
	// Actualizar el mensaje de respuesta en la interfaz
	if messageDisplay != nil {
		messageDisplay.SetText(message)
	}
}

func handleServerMessage(message string) {
	fmt.Println("Mensaje leído para el Handle: ", message)
	if strings.HasPrefix(message, "QUESTION:") {
		options := strings.Split(strings.TrimPrefix(message, "QUESTION:"), "\n")
		updateQuestion(options[0])
		updateOptions(options[1:])
	} else if strings.TrimSpace(message) == "CORRECT" {
		updateAnswerMessage("Respuesta Correcta")
	} else if strings.TrimSpace(message) == "INCORRECT" {
		updateAnswerMessage("Respuesta Incorrecta")
	} else {
		updateChatMessage(message)
	}
}

func updateOptions(options []string) {
	for _, button := range optionsButtons {
		button.Hide()
	}

	optionsButtons = make([]*widget.Button, len(options))
	for i, option := range options {
		optionsButtons[i] = widget.NewButton(option, func(option string) func() {
			return func() {
				sendAnswerToServer(option)
			}
		}(option))
		optionsButtons[i].Show()
	}
}

func openGameWindow() {
	gameWindow := myApp.NewWindow("Game")
	gameWindow.Resize(fyne.NewSize(400, 400))

	questionLabel = widget.NewLabel("")
	messageDisplay = widget.NewLabel("") // Inicializar messageDisplay aquí

	answerEntry = widget.NewEntry()
	checkButton := widget.NewButton("Check", func() {
		sendAnswerToServer(answerEntry.Text)
		answerEntry.SetText("")
	})

	gameWindow.SetContent(container.NewVBox(
		questionLabel,
		messageDisplay, // Agregar messageDisplay al contenido
		container.NewGridWithColumns(2,
			answerEntry,
			checkButton,
		),
	))

	conn.Write([]byte("GET_QUESTION\n"))

	gameWindow.Show()
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

func openChooseWindow() {
	chooseWindow := myApp.NewWindow("Elegir")
	chooseWindow.SetContent(container.NewVBox(
		widget.NewButton("Chat", func() {
			openChatWindow()
			chooseWindow.Hide()
		}),
		widget.NewButton("Jugar", func() {
			openGameWindow()
			chooseWindow.Hide()
		}),
	))
	chooseWindow.Resize(fyne.NewSize(400, 400))
	chooseWindow.Show()
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

func sendChatMessage(message string) {
	conn.Write([]byte(message + "\n"))
}

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
