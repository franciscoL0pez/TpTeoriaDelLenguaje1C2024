package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Botones y Consola")

	file, err := os.OpenFile("puntos.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("Error al abrir el archivo CSV:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	record, err := reader.Read()
	var puntos int
	if err == nil && len(record) > 0 {
		puntos, _ = strconv.Atoi(record[0])
	}

	buttonTexts := []string{"Respuesta1", "Respuesta2", "Respuesta3", "Respuesta4"}

	updatePoints := func() {
		file.Seek(0, 0)
		writer := csv.NewWriter(file)
		defer writer.Flush()

		puntosStr := strconv.Itoa(puntos)
		err := writer.Write([]string{puntosStr})
		if err != nil {
			fmt.Println("Error al escribir en el archivo CSV:", err)
		}
	}

	printText := func(text string) {

		if text == "Respuesta1" {
			puntos++
			updatePoints()
			print("Repuesta correcta!")
		} else {
			print("Respuesta incorrecta!")
		}
	}

	buttons := make([]*widget.Button, len(buttonTexts))
	for i, text := range buttonTexts {
		text := text
		buttons[i] = widget.NewButton(text, func() {
			printText(text)
		})
	}

	buttonGrid := container.NewGridWithColumns(2,
		buttons[0], buttons[1],
		buttons[2], buttons[3],
	)

	myWindow.SetContent(buttonGrid)

	myWindow.ShowAndRun()
}