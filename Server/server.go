package Server

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type questionAnswer struct {
	question string
	answer   string
	option1  string
	option2  string
	option3  string
}

var (
	clients         = make(map[net.Conn]bool)
	currentAnswer   string
	questionDict    = make(map[string][]questionAnswer)
	currentQuestion string
	keys            = []string{"Ciencia", "Deportes", "Entretenimiento", "Historia"}
	indiceKey       int
)

func Authenticate(username, password string) bool {
	file, err := os.Open("users.csv")
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return false
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error al leer el archivo CSV:", err)
		return false
	}

	for _, record := range records {
		if record[0] == username && record[1] == password {
			return true
		}
	}

	return false
}

func Register(username, password string) error {
	file, err := os.OpenFile("users.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.Write([]string{username, password})
	if err != nil {
		return err
	}

	writer.Flush()

	return nil
}

func addPointsToUser(username string) error {
	file, err := os.OpenFile("Points/puntos.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("Error al abrir el archivo CSV:", err)
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	userFound := false
	for i, record := range records {
		if record[0] == username {
			points, err := strconv.Atoi(record[1])
			if err != nil {
				fmt.Println("Error to convert")
				return err
			}
			points++
			records[i][1] = strconv.Itoa(points)
			userFound = true
			break
		}
	}

	if !userFound {
		newRecord := []string{username, "1"}
		records = append(records, newRecord)
	}

	file, err = os.Create("Points/puntos.csv")
	if err != nil {
		return err
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.WriteAll(records)
	if err != nil {
		return err
	}

	return nil

}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Nueva conexion:", conn.RemoteAddr())

	conn.Write([]byte("Bienvenido al servidor.\n"))
	conn.Write([]byte("1. Iniciar sesion\n"))
	conn.Write([]byte("2. Registrarse\n"))
	conn.Write([]byte("Ingrese opcion:"))

	reader := bufio.NewReader(conn)
	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	switch option {
	case "1": // Iniciar sesión
		conn.Write([]byte("Ingrese su nombre de usuario:"))
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		conn.Write([]byte("Ingrese su contraseña:"))
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		if Authenticate(username, password) {
			fmt.Println("Usuario autenticado:", username)
			conn.Write([]byte("Bienvenido, " + username + "!\n"))
		} else {
			fmt.Println("Autenticacion fallida para el usuario:", username)
			conn.Write([]byte("Autenticacion fallida. Cierre de la conexion.\n"))
			return
		}
	case "2": // Registrarse
		conn.Write([]byte("Ingrese un nombre de usuario para registrarse:"))
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		conn.Write([]byte("Ingrese una contraseña para registrarse:"))
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		if err := Register(username, password); err != nil {
			fmt.Println("Error al registrar el usuario:", err)
			conn.Write([]byte("Error al registrar el usuario.\n"))
			return
		}

		fmt.Println("Usuario registrado:", username)
		conn.Write([]byte("Usuario registrado con exito.\n"))
	default:
		conn.Write([]byte("Opción invalida. Cierre de la conexion.\n"))
		return
	}

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer mensaje del cliente:", err)
			delete(clients, conn)
			return
		}
		message = strings.TrimSpace(message)
		indice := strings.Index(message, ":")
		messageNameSave := message
		messageName := message
		if indice != -1 {
			fmt.Println("Message antes de trimspace: " + message)
			message = strings.TrimSpace(message[indice+2:])
			messageName = strings.TrimSpace(messageName[:indice-1])
		}

		// Dividir el mensaje en palabras
		parts := strings.Fields(message)
		if len(parts) > 0 {
			switch parts[0] {
			case "GET_QUESTION":
				SendQuestionToClient(conn)
			case "ANSWER":
				// Verificar si hay al menos dos partes (ANSWER y el número de respuesta)
				if len(parts) >= 2 {
					answer := strings.Join(parts[1:], " ")
					//answer := parts[1] // Obtener el número de respuesta
					fmt.Println("Respuesta recibida:", answer)
					if CheckAnswer(answer) {
						fmt.Println("Respuesta Correcta")
						addPointsToUser(messageName)
						conn.Write([]byte("CORRECT\n"))
					} else {
						fmt.Println("Respuesta Incorrecta")
						conn.Write([]byte("INCORRECT\n"))
					}
					SendQuestionToClient(conn)
				} else {
					fmt.Println("Mensaje de respuesta incorrecto:", message)
				}
			default:
				for client := range clients {
					_, err := client.Write([]byte(messageNameSave + "\n"))
					if err != nil {
						client.Close()
						delete(clients, client)
					}
				}
			}
		}
	}
}

func SendQuestionToClient(conn net.Conn) {
	q := RandomQuestion()
	currentQuestion = q.question
	currentAnswer = q.answer
	currentOption1 := q.option1
	currentOption2 := q.option2
	currentOption3 := q.option3
	options := []string{currentAnswer, currentOption1, currentOption2, currentOption3}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	conn.Write([]byte("CATEGORY:" + keys[indiceKey] + "\n"))

	conn.Write([]byte("QUESTION:" + q.question + "\n"))
	for _, value := range options {
		conn.Write([]byte("OPTION:" + value + "\n"))
	}
	conn.Write([]byte("END_OPTION\n"))
}

func LoadQuestions() {

	files_text := []string{"Questions/ciencia.csv", "Questions/deportes.csv", "Questions/entretenimiento.csv", "Questions/historia.csv"}
	for i, fi := range files_text {
		fmt.Println("Leyendo archivo: " + fi)
		questionList := make([]questionAnswer, 0)
		file, err := os.Open(fi)
		if err != nil {
			fmt.Println("Error al abrir el archivo:", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)

		records, err := reader.ReadAll()
		if err != nil {
			fmt.Println("Error al leer el archivo CSV:", err)
			return
		}

		for _, record := range records {
			q := questionAnswer{
				question: record[0],
				answer:   record[4],
				option1:  record[2],
				option2:  record[3],
				option3:  record[1],
			}
			questionList = append(questionList, q)
		}
		questionDict[keys[i]] = questionList
	}
}

func RandomQuestion() questionAnswer {
	rand.Seed(time.Now().UnixNano())
	indiceKey = rand.Intn(len(keys))
	indice_ques := rand.Intn(len(questionDict[keys[indiceKey]]))

	return questionDict[keys[indiceKey]][indice_ques]
}

func CheckAnswer(answer string) bool {
	fmt.Println("Answer: ", answer)
	fmt.Println("Current answer: ", currentAnswer)
	return answer == currentAnswer
}

func InitServer() {
	LoadQuestions()
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Servidor escuchando en el puerto 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error al aceptar la conexion:", err)
			continue
		}
		clients[conn] = true
		go HandleConnection(conn)
	}
}
