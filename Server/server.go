package Server

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var (
	clients         = make(map[net.Conn]bool)
	questionList    []questionAnswer
	currentQuestion string
	currentAnswer   string
)

type questionAnswer struct {
	question string
	answer   string
	option1  string
	option2  string
	option3  string
}

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
		if indice != -1 {
			message = strings.TrimSpace(message[indice+2:])
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
					answer := parts[1] // Obtener el número de respuesta
					fmt.Println("Respuesta recibida:", answer)
					if CheckAnswer(answer) {
						fmt.Println("Respuesta Correcta")
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
					_, err := client.Write([]byte(message + "\n"))
					if err != nil {
						fmt.Println("Error al enviar mensaje al cliente:", err)
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

	conn.Write([]byte("QUESTION:" + q.question + "\n"))
	for _, value := range options {
		conn.Write([]byte("OPTION:" + value + "\n"))
	}
	conn.Write([]byte("END_OPTION\n"))
}

func LoadQuestions() {
	file, err := os.Open("questions.csv")
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
			option1:  record[1],
			option2:  record[2],
			option3:  record[3],
		}
		questionList = append(questionList, q)
	}
}

func RandomQuestion() questionAnswer {
	rand.Seed(time.Now().UnixNano())
	indice := rand.Intn(len(questionList))
	return questionList[indice]
}

func CheckAnswer(answer string) bool {
	fmt.Println("Answer: ", answer)
	fmt.Println("Current answer: ", currentAnswer)
	return answer == currentAnswer // Simplificado para este ejemplo, deberías verificar con la respuesta correcta
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
