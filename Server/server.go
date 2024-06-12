package Server

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn     net.Conn
	username string
	points   int
	indice   int
}

type questionAnswer struct {
	question string
	answer   string
	option1  string
	option2  string
	option3  string
}

var (
	clients            = make(map[net.Conn]*Client)
	clientsWaitingPlay = make(map[net.Conn]*Client)
	clientsPlaying     = make(map[int][]*Client) // Mapa de índice de partida a lista de jugadores
	currentAnswer      = make(map[net.Conn]string)
	currentQuestion    = make(map[net.Conn]string)
	questionDict       = make(map[string][]questionAnswer)
	keys               = []string{"Ciencia", "Deportes", "Entretenimiento", "Historia"}
	indiceKey          int
	indicePartida      = 0
	mutex              sync.Mutex
)

func assignRival() {
	for {
		mutex.Lock()
		for conn1, client1 := range clientsWaitingPlay {
			for conn2, client2 := range clientsWaitingPlay {
				if conn1 != conn2 {
					clientsPlaying[indicePartida] = []*Client{client1, client2}
					client1.indice = indicePartida
					client2.indice = indicePartida
					sendReadyMessage(clientsPlaying[indicePartida])
					indicePartida++
					delete(clientsWaitingPlay, conn1)
					delete(clientsWaitingPlay, conn2)
				}
			}
		}
		mutex.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

func sendReadyMessage(players []*Client) {
	for _, player := range players {
		opponent := players[0]
		if player == players[0] {
			opponent = players[1]
		}
		player.conn.Write([]byte("READY:" + opponent.username + "\n"))
	}
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
			clients[conn] = &Client{conn, username, 0, 0}
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
		clients[conn] = &Client{conn, username, 0, 0}
	default:
		conn.Write([]byte("Opción invalida. Cierre de la conexion.\n"))
		return
	}

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer mensaje del cliente:", err)
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			return
		}
		message = strings.TrimSpace(message)
		indice := strings.Index(message, ":")
		messageNameSave := message
		if indice != -1 {
			fmt.Println("Message antes de trimspace: " + message)
			message = strings.TrimSpace(message[indice+2:])
		}

		parts := strings.Fields(message)
		if len(parts) > 0 {
			switch parts[0] {
			case "GIVE_STATS":
				mutex.Lock()
				SendStatsToClient(conn)
				mutex.Unlock()
			case "WANT_PRACTISE":
				mutex.Lock()
				SendQuestionToClient(conn)
				conn.Write([]byte("READY_PRACTISE\n"))
				mutex.Unlock()
			case "NOT_WANT_PLAY":
				mutex.Lock()
				delete(clientsWaitingPlay, conn)
				mutex.Unlock()
			case "WANT_PLAY":
				mutex.Lock()
				clientsWaitingPlay[conn] = clients[conn]
				mutex.Unlock()
			case "GET_QUESTION":
				mutex.Lock()
				SendQuestionToClient(conn)
				mutex.Unlock()
			case "ANSWER_PRACTISE":
				answer := strings.Join(parts[1:], " ")
				fmt.Println("Respuesta recibida:", answer)
				mutex.Lock()
				SendQuestionToClient(conn)
				if CheckAnswer(answer, conn) {
					fmt.Println("Respuesta Correcta")
					conn.Write([]byte("CORRECT_PRACTISE\n"))
				} else {
					fmt.Println("Respuesta Incorrecta")
					conn.Write([]byte("INCORRECT_PRACTISE\n"))
				}
				mutex.Unlock()
			case "ANSWER":
				if len(parts) >= 2 {
					answer := strings.Join(parts[1:], " ")
					fmt.Println("Respuesta recibida:", answer)
					mutex.Lock()
					if CheckAnswer(answer, conn) {
						fmt.Println("Respuesta Correcta")
						client := clients[conn]
						client.points++
						addPointsToUser(client.username)
						if client.points >= 2 {
							endGame(client)
						} else {
							conn.Write([]byte("CORRECT\n"))
							SendQuestionToClient(conn)
						}
					} else {
						fmt.Println("Respuesta Incorrecta")
						conn.Write([]byte("INCORRECT\n"))
						SendQuestionToClient(conn)
					}
					mutex.Unlock()
				} else {
					fmt.Println("Mensaje de respuesta incorrecto:", message)
				}

			default:
				mutex.Lock()
				for _, client := range clients {
					_, err := client.conn.Write([]byte(messageNameSave + "\n"))
					if err != nil {
						client.conn.Close()
						delete(clients, client.conn)
					}
				}
				mutex.Unlock()
			}
		}
	}
}

func endGame(winner *Client) {
	players := clientsPlaying[winner.indice]
	for _, player := range players {
		player.points = 0
		if player == winner {
			player.conn.Write([]byte("WINNER\n"))
		} else {
			player.conn.Write([]byte("LOOSER\n"))
		}
	}
	delete(clientsPlaying, winner.indice)
}

func SendStatsToClient(conn net.Conn) {
	file, err := os.Open("Points/puntos.csv")
	if err != nil {
		fmt.Println("Error al abrir el archivo CSV:", err)
		conn.Write([]byte("Error al obtener estadísticas.\n"))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error al leer el archivo CSV:", err)
		conn.Write([]byte("Error al obtener estadísticas.\n"))
		return
	}

	type Player struct {
		Name   string
		Points int
	}

	var players []Player
	for _, record := range records {
		points, err := strconv.Atoi(record[1])
		if err != nil {
			fmt.Println("Error al convertir puntos:", err)
			continue
		}
		players = append(players, Player{Name: record[0], Points: points})
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Points > players[j].Points
	})

	for i := 0; i < 5 && i < len(players); i++ {
		conn.Write([]byte(fmt.Sprintf("TOP_PLAYER: Top %d: %s - %d\n", i+1, players[i].Name, players[i].Points)))
	}
	conn.Write([]byte("END_STATS\n"))
}

func SendQuestionToClient(conn net.Conn) {
	q := RandomQuestion()
	currentQuestion[conn] = q.question
	currentAnswer[conn] = q.answer
	currentOption1 := q.option1
	currentOption2 := q.option2
	currentOption3 := q.option3
	options := []string{currentAnswer[conn], currentOption1, currentOption2, currentOption3}
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

func CheckAnswer(answer string, conn net.Conn) bool {
	fmt.Println("Answer: ", answer)
	fmt.Println("Current answer: ", currentAnswer[conn])
	return answer == currentAnswer[conn]
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
	go assignRival()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error al aceptar la conexion:", err)
			continue
		}
		go HandleConnection(conn)
	}
}
