package Backend

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type questionAnswer struct {
	question string
	answer   string
}

func WriteCSVQuestionsAndAnswer(question, answer string) error {
	// Abre el archivo CSV en modo append
	file, err := os.OpenFile("questions.csv", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{question, answer})
	if err != nil {
		return err

	}

	return nil
}

func NewQuestionAnswer() (string, string) {
	var question, answer string

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Ingrese la pregunta : ")
	scanner.Scan()
	question = scanner.Text()
	question = strings.TrimSpace(string(question))

	fmt.Print("Ingrese la respuesta: ")
	scanner.Scan()
	answer = scanner.Text()
	answer = strings.TrimSpace(string(answer))

	return question, answer

}

func SelecRandomCategory() string {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(4)

	categoryList := []string{"deportes.csv", "ciencia.csv", "entretenimiento.csv", "historia.csv"}

	category := categoryList[randomNumber]

	return category
}

func CreateQuestionList(nameArchvie string) ([]questionAnswer, error) {
	file, err := os.Open(nameArchvie)

	if err != nil {
		return nil, err
	}

	defer file.Close()
	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var questionList []questionAnswer

	// Iterar sobre las filas del archivo CSV y agregar preguntas y respuestas a la lista
	for _, row := range rows {
		questionAnswer := questionAnswer{
			question: row[0],
			answer:   row[1],
		}
		questionList = append(questionList, questionAnswer)
	}

	return questionList, nil
}

func RandomQuestion(questionList []questionAnswer) questionAnswer {
	rand.Seed(time.Now().UnixNano())

	indice := rand.Intn(len(questionList))

	return questionList[indice]
}

func GiveRandomQuestionToPlayer(questionList []questionAnswer) bool {
	q := RandomQuestion(questionList)
	fmt.Print(q.question)

	var answer string
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Ingrese la respuesta: ")

	scanner.Scan()

	answer = scanner.Text()
	answer = strings.TrimSpace(string(answer))

	if answer == q.answer {

		fmt.Print("Respuesta correcta!")
		return true

	} else {

		fmt.Print("Respuesta incorrecta!")
		return false
	}
}
