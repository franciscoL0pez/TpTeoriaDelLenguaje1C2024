package Interfaces

import (
	"TpTeoriaDelLenguaje1C2024/Client"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type UI struct {
	client         *Client.Client
	myApp          fyne.App
	loginWindow    fyne.Window
	messageDisplay *widget.Label
	questionLabel  *widget.Label
	options        []string
	optionsLabel   *widget.Label
}

func NewUI(client *Client.Client, app fyne.App) *UI {
	return &UI{client: client, myApp: app}
}

func (ui *UI) ShowLoginWindow() {
	ui.loginWindow = ui.myApp.NewWindow("Login")

	optionLabel := widget.NewLabel("Elige una opción:")
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginButton := widget.NewButton("Login", func() {
		err := ui.client.SendCredentials("1", usernameEntry.Text, passwordEntry.Text)
		if err == nil {
			go ui.client.ReceiveMessages(ui.handleServerMessage)
			ui.OpenChooseWindow()
			ui.loginWindow.Hide()
		} else {
			fmt.Println("Login failed:", err)
		}
	})

	registerButton := widget.NewButton("Register", func() {
		err := ui.client.SendCredentials("2", usernameEntry.Text, passwordEntry.Text)
		if err == nil {
			usernameEntry.SetText("")
			passwordEntry.SetText("")
		} else {
			fmt.Println("Registration failed:", err)
		}
	})

	ui.loginWindow.SetContent(container.NewVBox(
		optionLabel,
		usernameEntry,
		passwordEntry,
		container.NewHBox(
			loginButton,
			registerButton,
		),
	))
	ui.loginWindow.Resize(fyne.NewSize(400, 400))
	ui.loginWindow.Show()
}

func (ui *UI) OpenChooseWindow() {
	chooseWindow := ui.myApp.NewWindow("Elegir")
	chooseWindow.SetContent(container.NewVBox(
		widget.NewButton("Chat", func() {
			ui.OpenChatWindow()
			chooseWindow.Hide()
		}),
		widget.NewButton("Jugar", func() {
			ui.OpenGameWindow()
			chooseWindow.Hide()
		}),
	))
	chooseWindow.Resize(fyne.NewSize(400, 400))
	chooseWindow.Show()
}

func (ui *UI) OpenChatWindow() {
	chatWindow := ui.myApp.NewWindow("Chat")

	backButton := widget.NewButton("Back", func() {
		ui.OpenChooseWindow()
		chatWindow.Close()
	})

	messageEntry := widget.NewEntry()
	messageEntry.SetPlaceHolder("Type your message...")

	sendButton := widget.NewButton("Send", func() {
		message := messageEntry.Text
		ui.client.SendMessage(message)
		messageEntry.SetText("")
	})

	ui.messageDisplay = widget.NewLabel("")

	chatContent := container.NewVBox(ui.messageDisplay)
	chatScroll := container.NewVScroll(chatContent)
	chatScroll.SetMinSize(fyne.NewSize(400, 300))
	inputContainer := container.NewBorder(nil, nil, nil, sendButton, messageEntry)

	mainContainer := container.NewBorder(
		container.NewBorder(nil, nil, backButton, nil, nil),
		inputContainer,
		nil,
		nil,
		chatScroll,
	)

	chatWindow.SetContent(mainContainer)
	chatWindow.Resize(fyne.NewSize(400, 400))
	chatWindow.Show()
}

func (ui *UI) OpenGameWindow() {
	gameWindow := ui.myApp.NewWindow("Game")
	gameWindow.Resize(fyne.NewSize(400, 400))

	ui.questionLabel = widget.NewLabel("")
	ui.messageDisplay = widget.NewLabel("")
	ui.optionsLabel = widget.NewLabel("")

	SendAnser := func(text string) {
		if text == "A" {
			ui.client.SendMessage("ANSWER " + ui.options[0])
		} else if text == "B" {
			ui.client.SendMessage("ANSWER " + ui.options[1])
		} else if text == "C" {
			ui.client.SendMessage("ANSWER " + ui.options[2])
		} else {
			ui.client.SendMessage("ANSWER " + ui.options[3])
		}
	}

	buttonTexts := []string{"A", "B", "C", "D"}

	buttons := make([]*widget.Button, len(buttonTexts))
	for i, text := range buttonTexts {
		text := text
		buttons[i] = widget.NewButton(text, func() {
			SendAnser(text)
		})
	}
	buttonGrid := container.NewGridWithColumns(2,
		buttons[0], buttons[1],
		buttons[2], buttons[3],
	)

	mainContainer := container.NewVBox(
		ui.questionLabel,
		ui.optionsLabel,
		buttonGrid,
	)

	gameWindow.SetContent(mainContainer)
	ui.client.SendMessage("GET_QUESTION\n")

	gameWindow.Show()
}

func (ui *UI) handleServerMessage(message string) {
	fmt.Println("Mensaje leído para el Handle: ", message)
	if strings.HasPrefix(message, "QUESTION:") {
		ui.options = ui.options[:0]
		options := strings.Split(strings.TrimPrefix(message, "QUESTION:"), "\n")
		ui.updateQuestion(options[0])
	} else if strings.HasPrefix(message, "OPTION:") {
		opt := strings.Split(strings.TrimPrefix(message, "OPTION:"), "\n")
		fmt.Println("Opcion guardada: " + opt[0])
		ui.options = append(ui.options, opt[0])
	} else if strings.TrimSpace(message) == "END_OPTION" {
		ui.optionsLabel.SetText("")
		ui.updateOptionLabel()
	} else if strings.TrimSpace(message) == "CORRECT" {
		ui.updateAnswerMessage("Respuesta Correcta")
	} else if strings.TrimSpace(message) == "INCORRECT" {
		ui.updateAnswerMessage("Respuesta Incorrecta")
	} else {
		ui.updateChatMessage(message)
	}
}

func (ui *UI) updateChatMessage(message string) {
	ui.messageDisplay.SetText(ui.messageDisplay.Text + message)
}

func (ui *UI) updateQuestion(question string) {
	ui.questionLabel.SetText(strings.TrimPrefix(question, "QUESTION:"))
}

func (ui *UI) updateAnswerMessage(message string) {
	if ui.messageDisplay != nil {
		ui.messageDisplay.SetText(message)
	}
}

func (ui *UI) updateOptionLabel() {
	ui.optionsLabel.SetText("A: " + ui.options[0] + "\n" + "B: " + ui.options[1] + "\n" + "C: " + ui.options[2] + "\n" + "D: " + ui.options[3])

}

func InitUser() {
	myApp := app.New()
	client := &Client.Client{}
	ui := NewUI(client, myApp)
	welcomeWindow := myApp.NewWindow("Preguntados")
	welcomeWindow.SetContent(container.NewVBox(
		widget.NewLabel("Bienvenido a Preguntados"),
		widget.NewButton("Iniciar sesion", func() {
			ui.ShowLoginWindow()
			welcomeWindow.Hide()
		}),
	))
	welcomeWindow.Resize(fyne.NewSize(400, 400))
	welcomeWindow.Show()

	myApp.Run()

}
