package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Sensor struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func step(value int) int {
	r := rand.Float64()
	if r > 0.5 {
		value += 1
	} else {
		value -= 1
	}

	if value > 50 {
		value = 50
	}
	if value < -10 {
		value = -10
	}
	return value
}

func readId(reader *bufio.Reader) string {
	for {
		clearTerminal()
		fmt.Print("\nDigite o ID do sensor de temperatura: ")
		idStr, _ := reader.ReadString('\n')
		idStr = strings.TrimSpace(idStr)

		_, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Println("\nDigite apenas números")
			reader = bufio.NewReader(os.Stdin)
			fmt.Println("\nPressione ENTER para tentar novamente")
			reader.ReadString('\n')
			continue
		}

		return idStr
	}
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	serverIP := os.Args[1]

	clearTerminal()

	reader := bufio.NewReader(os.Stdin)
	id := readId(reader)
	temp := rand.Intn(61) - 10

	clearTerminal()
	fmt.Printf("\nSensor de temperatura %s inicializado.\n", id)
	fmt.Println("\nValor: ", temp)

	for {
		conn, err := net.Dial("udp", serverIP+":7000")
		if err != nil {
			continue
		}

		counter := 0

		for {
			temp = step(temp)

			if counter >= 1000 {
				data := Sensor{
					ID:    id,
					Type:  "Temperatura",
					Value: temp,
				}

				values, _ := json.Marshal(data)

				_, err := conn.Write(values)
				if err != nil {
					fmt.Println("\nErro no envio do sensor de temperatura: ", id, err)
					conn.Close()
					break
				}

				buffer := make([]byte, 1024)
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))

				n, err := conn.Read(buffer)
				if err != nil {
					fmt.Println("\nServidor não respondeu:", err)
					conn.Close()
					break
				}

				var response Response
				if err := json.Unmarshal(buffer[:n], &response); err != nil {
					fmt.Println("\nErro ao decodificar resposta:", err)
					break
				}

				if response.Status == "error" {
					fmt.Println("\n", response.Error)
					fmt.Println("\nPressione ENTER para informar outro ID")
					reader.ReadString('\n')
					id = readId(reader)
					clearTerminal()
					fmt.Printf("\nSensor de temperatura %s inicializado.\n", id)
					fmt.Println("\nValor: ", temp)
					counter = 0
					continue
				}

				clearTerminal()
				fmt.Printf("\nSensor de temperatura %s inicializado.\n", id)
				fmt.Println("\nValor: ", temp)
				counter = 0
			}

			counter++
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func clearTerminal() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
