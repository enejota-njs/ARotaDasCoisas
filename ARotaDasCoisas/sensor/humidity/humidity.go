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

	if value > 100 {
		value = 100
	}
	if value < 0 {
		value = 0
	}

	return value
}

func readId(reader *bufio.Reader) string {
	for {
		clearTerminal()
		fmt.Print("\nDigite o ID do sensor de umidade: ")
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
	clearTerminal()

	reader := bufio.NewReader(os.Stdin)
	id := readId(reader)
	humi := rand.Intn(101)

	clearTerminal()
	fmt.Printf("\nSensor de umidade %s inicializado.\n", id)
	fmt.Println("\nValor: ", humi)

	for {
		conn, err := net.Dial("udp", "127.0.0.1:7000")
		if err != nil {
			continue
		}

		counter := 0

		for {
			humi = step(humi)

			if counter >= 1000 {
				data := Sensor{
					ID:    id,
					Type:  "Umidade",
					Value: humi,
				}

				values, _ := json.Marshal(data)

				_, err := conn.Write(values)
				if err != nil {
					fmt.Println("\nErro no envio do sensor de umidade: ", id, err)
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
					fmt.Printf("\nSensor de umidade %s inicializado.\n", id)
					fmt.Println("\nValor: ", humi)
					counter = 0
					continue
				}

				clearTerminal()
				fmt.Printf("\nSensor de umidade %s inicializado.\n", id)
				fmt.Println("\nValor: ", humi)
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
