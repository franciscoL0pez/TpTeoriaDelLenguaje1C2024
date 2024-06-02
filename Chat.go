package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"fyne.io/fyne"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ChatApp struct {
	app         fyne.App
	window      fyne.Window
	chatHistory *widget.Box
	username    string
}

func NewChatApp() *ChatApp {
	app := app.New()
	window := app.NewWindow("Chat App")
	chatHistory := widget.NewVBox()

	window.SetContent(container.NewVBox(
		chatHistory,
		widget.NewEntry(),
	))

	return &ChatApp{
		app:         app,
		window:      window,
		chatHistory: chatHistory,
		username:    "",
	}
}

func (c *ChatApp) Run() {
	c.window.ShowAndRun()
}

func (c *ChatApp) handleNewMessage(msg string) {
	chatMessage := fmt.Sprintf("%s: %s", c.username, msg)
	c.chatHistory.Append(widget.NewLabel(chatMessage))
}

func (c *ChatApp) handleEntryTyped(entry *widget.Entry) {
	message := entry.Text

	if c.username == "" {
		if c.authenticateUser(message) {
			entry.SetText("")
			return
		}
	}

	c.handleNewMessage(message)
	entry.SetText("")
}

func (c *ChatApp) authenticateUser(username string) bool {
	file, err := os.Open("usuarios.csv")
	if err != nil {
		fmt.Println("Error abriendo el archivo:", err)
		return false
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error leyendo el archivo:", err)
			return false
		}

		if record[0] == username {
			c.username = username
			return true
		}
	}

	return false
}

func main() {
	chatApp := NewChatApp()

	chatApp.window.Canvas().SetOnTypedKey(func(keyEvent *fyne.KeyEvent) {
		if keyEvent.Name == fyne.KeyReturn {
			entry := chatApp.window.Content().Children[1].(*widget.Entry)
			chatApp.handleEntryTyped(entry)
		}
	})

	chatApp.Run()
}
