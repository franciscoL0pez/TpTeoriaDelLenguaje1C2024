package Interfaces

import (
	"TpTeoriaDelLenguaje1C2024/Client"
	"fmt"
	"image/color"
	"math/rand"
	"strings"
	"time"
	"io/ioutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UI struct {
	client         *Client.Client
	myApp          fyne.App
	loginWindow    fyne.Window
	statsLabel     *widget.Label
	messageDisplay *widget.Label
	questionLabel  *widget.Label
	options        []string
	top            []string
	optionsLabel   *widget.Label
	categoryLabel  *widget.Label
	rivalLabel     *widget.Label
	gameWindow     fyne.Window
	waitWindow     fyne.Window
	incorrectShown bool
	category       string
	rival          string
	timerEnabled   bool
}


var (
	cienciaBytes, _         = ioutil.ReadFile("Pics/ciencia.jpeg")
	entretenimientoBytes, _ = ioutil.ReadFile("Pics/entretenimiento.jpeg")
	deporteBytes, _         = ioutil.ReadFile("Pics/deportes.jpeg")
	historiaBytes, _        = ioutil.ReadFile("Pics/historia.jpeg")

	resourceCienciaJPEG         = fyne.NewStaticResource("Pics/ciencia.jpeg", cienciaBytes)
	resourceEntretenimientoJPEG = fyne.NewStaticResource("Pics/entretenimiento.jpeg", entretenimientoBytes)
	resourceDeporteJPEG         = fyne.NewStaticResource("Pics/deportes.jpeg", deporteBytes)
	resourceHistoriaJPEG        = fyne.NewStaticResource("Pics/historia.jpeg", historiaBytes)
)

func NewUI(client *Client.Client, app fyne.App) *UI {
	return &UI{client: client, myApp: app}
}

func (ui *UI) ShowLoginWindow() {
    ui.loginWindow = ui.myApp.NewWindow("Login")

    optionLabel := widget.NewLabel("Elige una opción:")
    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Nombre de usuario")
    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Contraseña")

    loginButton := widget.NewButtonWithIcon("Iniciar sesión", theme.LoginIcon(), func() {
        err := ui.client.SendCredentials("1", usernameEntry.Text, passwordEntry.Text)
        if err == nil {
            go ui.client.ReceiveMessages(ui.handleServerMessage)
            ui.OpenChooseWindow()
            ui.loginWindow.Hide()
        } else {
            fmt.Println("Error al iniciar sesión:", err)
        }
    })

    registerButton := widget.NewButtonWithIcon("Registrarse", theme.AccountIcon(), func() {
        err := ui.client.SendCredentials("2", usernameEntry.Text, passwordEntry.Text)
        if err == nil {
            usernameEntry.SetText("")
            passwordEntry.SetText("")
        } else {
            fmt.Println("Error al registrarse:", err)
        }
    })

    loginButton.Importance = widget.HighImportance

    box := container.NewVBox(
        optionLabel,
        usernameEntry,
        passwordEntry,
        container.NewHBox(
            loginButton,
            registerButton,
        ),
    )

    ui.loginWindow.SetContent(box)
    ui.loginWindow.Resize(fyne.NewSize(400, 400))
    ui.loginWindow.CenterOnScreen()
    ui.loginWindow.Show()
}

func (ui *UI) OpenPractiseWindow() {
	if ui.gameWindow != nil {
		ui.gameWindow.Hide()
	}
	if ui.waitWindow != nil {
		ui.waitWindow.Hide()
	}

	ui.timerEnabled = true

	gameWindow := ui.myApp.NewWindow("Game")
	ui.gameWindow = gameWindow

	ui.gameWindow.Resize(fyne.NewSize(400, 400))

	ui.questionLabel = widget.NewLabel("")
	ui.messageDisplay = widget.NewLabel("")
	ui.optionsLabel = widget.NewLabel("")

	timerLabel := widget.NewLabel("20")
	ui.categoryLabel = widget.NewLabel("Categoría: " + ui.category)

	timerContainer := container.NewHBox(
		widget.NewLabel("Tiempo restante: "),
		timerLabel,
		ui.categoryLabel,
	)

	backButton := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		ui.timerEnabled = false
		ui.gameWindow.Close()
		ui.OpenChooseWindow()
	})

	topContainer := container.NewHBox(
		backButton,
		timerContainer,
	)

	done := make(chan bool)

	go func() {
		defer close(done)
		timer := 20
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !ui.timerEnabled {
					return
				}
				timer--
				if timer <= 0 {
					if ui.timerEnabled {
						ui.SendPractiseAnswer("INCORRECTO")
						ui.ShowMessageWindow("Respuesta Incorrecta")
					}
					return
				}
				timerLabel.SetText(fmt.Sprintf("%d", timer))
			case <-done:
				return
			}
		}
	}()

	buttonTexts := []string{"A", "B", "C", "D"}

	buttons := make([]*widget.Button, len(buttonTexts))
	for i, text := range buttonTexts {
		text := text
		buttons[i] = widget.NewButton(text, func() {
			ui.SendPractiseAnswer(text)
			done <- true
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

	newMainContainer := container.NewVBox(
		topContainer,
		mainContainer,
	)
	ui.client.SendMessage("GET_QUESTION\n")

	var backgroundImage fyne.Resource
    switch ui.category {
    case "Ciencia":
        backgroundImage = resourceCienciaJPEG
    case "Entretenimiento":
        backgroundImage = resourceEntretenimientoJPEG
    case "Deportes":
        backgroundImage = resourceDeporteJPEG
    case "Historia":
        backgroundImage = resourceHistoriaJPEG
    default:
    }


    if backgroundImage != nil {
        bg := canvas.NewImageFromResource(backgroundImage)
        bg.FillMode = canvas.ImageFillOriginal
        ui.gameWindow.SetContent(container.NewMax(bg, newMainContainer))
    } else {
        ui.gameWindow.SetContent(newMainContainer)
    }

	
	if ui.waitWindow != nil {
		ui.waitWindow.Hide()
	}

	ui.gameWindow.SetContent(newMainContainer)
	ui.gameWindow.Show()
}

func (ui *UI) OpenWaitingWindow() {
	ui.waitWindow = ui.myApp.NewWindow("Esperando")

	waitLabel := widget.NewLabel("Buscando rival...")

	circle := canvas.NewCircle(color.NRGBA{R: 0xff, G: 0, B: 0, A: 0xff})
	circle.Resize(fyne.NewSize(25, 25))

	move := canvas.NewPositionAnimation(fyne.NewPos(0, 50), fyne.NewPos(375, 50), time.Second, circle.Move)
	move.AutoReverse = true
	go func() {
		for {
			move.Start()
			time.Sleep(2 * time.Second)
		}
	}()

	randomColor := func() color.Color {
		r := uint8(rand.Intn(256))
		g := uint8(rand.Intn(256))
		b := uint8(rand.Intn(256))
		return color.NRGBA{R: r, G: g, B: b, A: 0xff}
	}

	go func() {
		for {
			newColor := randomColor()
			oldColor := circle.FillColor
			start := time.Now()
			duration := time.Second

			for time.Since(start) < duration {
				progress := float64(time.Since(start)) / float64(duration)
				r1, g1, b1, _ := oldColor.RGBA()
				r2, g2, b2, _ := newColor.RGBA()

				r := uint8(float64(r1) + progress*float64(r2-r1))
				g := uint8(float64(g1) + progress*float64(g2-g1))
				b := uint8(float64(b1) + progress*float64(b2-b1))

				circle.FillColor = color.NRGBA{R: r, G: g, B: b, A: 0xff}
				canvas.Refresh(circle)
				time.Sleep(time.Millisecond * 400)
			}
		}
	}()

	circleContainer := container.NewWithoutLayout(circle)

	backButton := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		ui.waitWindow.Close()
		ui.client.SendMessage("NOT_WANT_PLAY\n")
		ui.OpenChooseWindow()
	})

	ui.waitWindow.SetContent(container.NewVBox(
		waitLabel,
		circleContainer,
		backButton,
	))

	ui.waitWindow.Resize(fyne.NewSize(400, 400))
	ui.waitWindow.Show()
}

func (ui *UI) OpenStatsWindow() {
	statsWindow := ui.myApp.NewWindow("Top 5 - Mejores jugadores")

	backButton := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		statsWindow.Close()
		ui.OpenChooseWindow()
	})

	if ui.statsLabel == nil {
		ui.statsLabel = widget.NewLabel("")
	}
	ui.statsLabel.SetText("")

	var statsText string
	for _, player := range ui.top {
		statsText += player + "\n\n"
	}
	ui.top = nil

	ui.statsLabel.SetText(statsText)

	statsWindow.SetContent(container.NewVBox(
		ui.statsLabel,
		backButton,
	))

	statsWindow.Resize(fyne.NewSize(400, 400))
	statsWindow.Show()
}

func (ui *UI) OpenChooseWindow() {
	chooseWindow := ui.myApp.NewWindow("Elegir")

	chooseWindow.SetContent(container.NewVBox(
		widget.NewButtonWithIcon("Jugar", theme.MediaPlayIcon(), func() {
			ui.client.SendMessage("WANT_PLAY\n")
			ui.OpenWaitingWindow()
			chooseWindow.Hide()
		}),
		widget.NewButtonWithIcon("Practicar", theme.InfoIcon(), func() {
			chooseWindow.Hide()
			ui.OpenPractiseWindow()
		}),
		widget.NewButtonWithIcon("Chat", theme.MailComposeIcon(), func() {
			ui.OpenChatWindow()
			chooseWindow.Hide()
		}),
		widget.NewButtonWithIcon("Reglas", theme.QuestionIcon(), func() {
			chooseWindow.Hide()
			ui.OpenRulesWindow(chooseWindow)
		}),
		widget.NewButtonWithIcon("Estadisticas", theme.StorageIcon(), func() {
			ui.client.SendMessage("GIVE_STATS\n")
			chooseWindow.Hide()
			ui.OpenStatsWindow()
		}),
	))
	chooseWindow.Resize(fyne.NewSize(400, 400))
	chooseWindow.Show()
}

func (ui *UI) OpenRulesWindow(parentWindow fyne.Window) {
	rulesWindow := ui.myApp.NewWindow("Reglas del Juego")

	backButton := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		rulesWindow.Close()
		parentWindow.Show()
	})

	rulesText := widget.NewRichText(
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "1. Objetivo del Juego:\n",
		},
		&widget.TextSegment{
			Text: "El objetivo del juego es responder correctamente al mayor número de preguntas\nen el menor tiempo posible, compitiendo contra otro jugador en tiempo real.\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "2. Inicio del Juego:\n",
		},
		&widget.TextSegment{
			Text: "Cada jugador debe registrarse para poder acceder al juego. El juego comienza\ncuando se encuentra un rival dispuesto a jugar.\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "3. Desarrollo del Juego:\n",
		},
		&widget.TextSegment{
			Text: "Cada ronda consta de una pregunta con cuatro opciones de respuesta: A, B, C y D.\nEl jugador tiene 20 segundos para seleccionar una respuesta.\nSi no se selecciona ninguna respuesta en el tiempo estipulado, se considera como incorrecta.\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "4. Sistema de Puntuación:\n",
		},
		&widget.TextSegment{
			Text: "Cada respuesta correcta otorga un punto. No se restan puntos por respuestas incorrectas.\nEl primer jugador en llegar a los 10 puntos es declarado ganador.\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "5. Comunicación:\n",
		},
		&widget.TextSegment{
			Text: "Los jugadores pueden comunicarse entre sí a través de un sistema de chat integrado en eljuego.\nSe espera que los jugadores mantengan un comportamiento\nrespetuoso y adecuado en el chat.\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "6. Conducta y Fair Play:\n",
		},
		&widget.TextSegment{
			Text: "Está prohibido el uso de cualquier tipo de trampas o ayudas externas.\nLos jugadores deben respetar las decisiones del sistema de juego y de los moderadores.\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "7. Penalizaciones:\n",
		},
		&widget.TextSegment{
			Text: "El incumplimiento de las normas de conducta puede llevar a sanciones como la suspensión\nde la cuenta o la expulsión definitiva del juego.\nEl uso de lenguaje inapropiado o comportamiento tóxico en el chat también será penalizado.\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "8. Ayuda y Soporte:\n",
		},
		&widget.TextSegment{
			Text: "Para cualquier problema técnico o dudas sobre el juego, los jugadores pueden contactar\ncon el soporte técnico a través del correo \"preguntados_support@fi.uba.ar\".\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "9. Actualizaciones y Mantenimiento:\n",
		},
		&widget.TextSegment{
			Text: "El juego puede estar sujeto a actualizaciones periódicas para mejorar la experiencia\ndel usuario. Durante los periodos de mantenimiento,\nalgunas funcionalidades del juego pueden no estar disponibles temporalmente.\n\n",
		},
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "10. Privacidad y Seguridad:\n",
		},
		&widget.TextSegment{
			Text: "La información personal de los jugadores se maneja de acuerdo con las políticas de\nprivacidad establecidas en la plataforma.\nSe recomienda no compartir información personal sensible en el chat del juego.\n",
		},
	)

	backButtonContainer := container.NewHBox(backButton)

	rulesContainer := container.NewVBox(rulesText)
	rulesScroll := container.NewVScroll(rulesContainer)
	rulesScroll.SetMinSize(fyne.NewSize(400, 300))

	mainContainer := container.NewBorder(
		backButtonContainer,
		nil,
		nil,
		nil,
		rulesScroll,
	)

	rulesWindow.SetContent(mainContainer)
	rulesWindow.Resize(fyne.NewSize(400, 400))
	rulesWindow.Show()
}

func (ui *UI) OpenChatWindow() {
	chatWindow := ui.myApp.NewWindow("Chat")

	backButton := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		ui.OpenChooseWindow()
		chatWindow.Close()
	})

	messageEntry := widget.NewEntry()
	messageEntry.SetPlaceHolder("Type your message...")

	sendButton := widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		message := messageEntry.Text
		if message != "" {
			ui.client.SendMessage(message)
			messageEntry.SetText("")
		}
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

func (ui *UI) updateRivalLabel() {
	if ui.rivalLabel != nil {
		ui.rivalLabel.SetText("Rival: " + ui.rival)
	} else {
		fmt.Println("Error: rivalLabel is nil")
	}
}

func (ui *UI) OpenGameWindow() {
	if ui.gameWindow != nil {
		ui.gameWindow.Hide()
	}
	ui.timerEnabled = true

	gameWindow := ui.myApp.NewWindow("Game")
	ui.gameWindow = gameWindow

	ui.gameWindow.Resize(fyne.NewSize(400, 400))

	ui.questionLabel = widget.NewLabel("")
	ui.messageDisplay = widget.NewLabel("")
	ui.optionsLabel = widget.NewLabel("")

	timerLabel := widget.NewLabel("20")
	ui.categoryLabel = widget.NewLabel("Categoría: " + ui.category)
	ui.rivalLabel = widget.NewLabel("Rival: " + ui.rival)

	timerContainer := container.NewHBox(
		widget.NewLabel("Tiempo restante: "),
		timerLabel,
		ui.categoryLabel,
	)

	if ui.waitWindow != nil {
		ui.waitWindow.Close()
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		timer := 20
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !ui.timerEnabled {
					return
				}
				timer--
				if timer <= 0 {
					if ui.timerEnabled {
						ui.SendAnswer("INCORRECTO")
						ui.ShowMessageWindow("Respuesta Incorrecta")
					}
					return
				}
				timerLabel.SetText(fmt.Sprintf("%d", timer))
			case <-done:
				return
			}
		}
	}()

	buttonTexts := []string{"A", "B", "C", "D"}

	buttons := make([]*widget.Button, len(buttonTexts))
	for i, text := range buttonTexts {
		text := text
		buttons[i] = widget.NewButton(text, func() {
			ui.SendAnswer(text)
			done <- true
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
		ui.rivalLabel,
	)

	newMainContainer := container.NewVBox(
		timerContainer,
		mainContainer,
	)

	var backgroundImage fyne.Resource
	switch ui.category {
	case "ciencia":
		backgroundImage = resourceCienciaJPEG
	case "entretenimiento":
		backgroundImage = resourceEntretenimientoJPEG
	case "deportes":
		backgroundImage = resourceDeporteJPEG
	case "historia":
		backgroundImage = resourceHistoriaJPEG
	default:
	}

	if backgroundImage != nil {
		bg := canvas.NewImageFromResource(backgroundImage)
		bg.FillMode = canvas.ImageFillOriginal
		ui.gameWindow.SetContent(container.NewMax(bg, newMainContainer))
	} else {
		ui.gameWindow.SetContent(newMainContainer)
	}

	if ui.waitWindow != nil {
		ui.waitWindow.Hide()
	}

	ui.gameWindow.Show()
}

func (ui *UI) SendAnswer(text string) {
	if text == "A" {
		ui.client.SendMessage("ANSWER " + ui.options[0])
	} else if text == "B" {
		ui.client.SendMessage("ANSWER " + ui.options[1])
	} else if text == "C" {
		ui.client.SendMessage("ANSWER " + ui.options[2])
	} else if text == "D" {
		ui.client.SendMessage("ANSWER " + ui.options[3])
	} else {
		ui.client.SendMessage("ANSWER " + "TIME OUT")
	}
}

func (ui *UI) SendPractiseAnswer(text string) {
	if text == "A" {
		ui.client.SendMessage("ANSWER_PRACTISE " + ui.options[0])
	} else if text == "B" {
		ui.client.SendMessage("ANSWER_PRACTISE " + ui.options[1])
	} else if text == "C" {
		ui.client.SendMessage("ANSWER_PRACTISE " + ui.options[2])
	} else if text == "D" {
		ui.client.SendMessage("ANSWER_PRACTISE " + ui.options[3])
	} else {
		ui.client.SendMessage("ANSWER_PRACTISE " + "TIME OUT")
	}
}

func (ui *UI) ShowMessageWindow(message string) {
	if ui.incorrectShown {
		return
	}

	ui.incorrectShown = true

	messageWindow := ui.myApp.NewWindow("Mensaje")
	messageLabel := widget.NewLabel(message)

	messageWindow.SetContent(container.NewVBox(
		messageLabel,
	))

	if ui.gameWindow != nil {
		ui.gameWindow.Hide()
	}

	messageWindow.Resize(fyne.NewSize(400, 400))
	messageWindow.Show()

	time.AfterFunc(3*time.Second, func() {
		messageWindow.Close()
		ui.incorrectShown = false
		ui.OpenGameWindow()
	})
}

func (ui *UI) ShowPractiseMessageWindow(message string) {
	if ui.incorrectShown {
		return
	}
	ui.incorrectShown = true

	messageWindow := ui.myApp.NewWindow("Mensaje")
	messageLabel := widget.NewLabel(message)

	messageWindow.SetContent(container.NewVBox(
		messageLabel,
	))

	if ui.gameWindow != nil {
		ui.gameWindow.Hide()
	}

	messageWindow.Resize(fyne.NewSize(400, 400))
	messageWindow.Show()

	time.AfterFunc(3*time.Second, func() {
		messageWindow.Close()
		ui.incorrectShown = false
		ui.OpenPractiseWindow()
	})
}

func (ui *UI) handleServerMessage(message string) {

	fmt.Println("Mensaje leído para el Handle: ", message)
	if strings.HasPrefix(message, "READY:") {
		time.Sleep(2 * time.Second)
		res := strings.Split(strings.TrimPrefix(message, "READY:"), "\n")
		fmt.Println("Partida VS: " + res[0])
		ui.rival = res[0]
		ui.updateRivalLabel()
		if ui.waitWindow != nil {
			ui.waitWindow.Close()
			ui.OpenGameWindow()
		}
	} else if strings.HasPrefix(message, "CATEGORY:") {
		res := strings.Split(strings.TrimPrefix(message, "CATEGORY:"), "\n")
		fmt.Println("Categoria leida: " + res[0])
		ui.category = res[0]
		ui.categoryLabel.SetText("Categoría: " + ui.category)
	} else if strings.HasPrefix(message, "QUESTION:") {
		ui.options = ui.options[:0]
		options := strings.Split(strings.TrimPrefix(message, "QUESTION:"), "\n")
		ui.questionLabel.SetText(strings.TrimPrefix(options[0], "QUESTION:"))
	} else if strings.HasPrefix(message, "OPTION:") {
		opt := strings.Split(strings.TrimPrefix(message, "OPTION:"), "\n")
		fmt.Println("Opcion guardada: " + opt[0])
		ui.options = append(ui.options, opt[0])
	} else if strings.TrimSpace(message) == "END_OPTION" {
		ui.optionsLabel.SetText("")
		ui.updateOptionLabel()
	} else if strings.HasPrefix(message, "TOP_PLAYER:") {
		opt := strings.Split(strings.TrimPrefix(message, "TOP_PLAYER:"), "\n")
		fmt.Println("Top Player guardado: " + opt[0])
		ui.top = append(ui.top, opt[0])
	} else if strings.TrimSpace(message) == "END_STATS" {

	} else if strings.TrimSpace(message) == "CORRECT" {
		ui.ShowMessageWindow("Respuesta Correcta")
	} else if strings.TrimSpace(message) == "INCORRECT" {
		ui.ShowMessageWindow("Respuesta Incorrecta")
	} else if strings.TrimSpace(message) == "CORRECT_PRACTISE" {
		ui.ShowPractiseMessageWindow("Respuesta Correcta")
	} else if strings.TrimSpace(message) == "INCORRECT_PRACTISE" {
		ui.ShowPractiseMessageWindow("Respuesta Incorrecta")
	} else if strings.TrimSpace(message) == "WINNER" {
		ui.timerEnabled = false
		ui.closeWindows()
		ui.ShowEndMessageWindow("¡Has ganado!")
	} else if strings.TrimSpace(message) == "LOOSER" {
		ui.timerEnabled = false
		ui.closeWindows()
		ui.ShowEndMessageWindow("¡Has perdido!")
	} else {
		ui.updateChatMessage(message)
	}
}

func (ui *UI) ShowEndMessageWindow(message string) {
	if ui.incorrectShown {
		return
	}
	ui.incorrectShown = true

	messageWindow := ui.myApp.NewWindow("Fin de partida")
	messageLabel := widget.NewLabel(message)

	messageWindow.SetContent(container.NewVBox(
		messageLabel,
	))

	messageWindow.Resize(fyne.NewSize(400, 400))
	messageWindow.Show()

	time.AfterFunc(3*time.Second, func() {
		messageWindow.Close()
		ui.incorrectShown = false
		ui.timerEnabled = false
		ui.OpenChooseWindow()
	})
}

func (ui *UI) closeWindows() {
	if ui.waitWindow != nil {
		ui.waitWindow.Close()
		ui.waitWindow = nil
	}
	if ui.gameWindow != nil {
		ui.gameWindow.Close()
		ui.gameWindow = nil
	}
}

func (ui *UI) updateChatMessage(message string) {
	ui.messageDisplay.SetText(ui.messageDisplay.Text + message)
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
		widget.NewButtonWithIcon("Login", theme.LoginIcon(), func() {
			ui.ShowLoginWindow()
			welcomeWindow.Hide()
		}),
	))
	welcomeWindow.Resize(fyne.NewSize(400, 400))
	welcomeWindow.Show()

	myApp.Run()
}
