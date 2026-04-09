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

// Estrutura do Sensor (Gás)
type Sensor struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value int    `json:"value"`
}

// Estrutura de Resposta do servidor
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// Função para simular a variação do valor lido pelo sensor ao longo do tempo
func step(value int) int {
	r := rand.Float64()
	if r > 0.5 {
		value += 2
	} else {
		value -= 2
	}

	if value > 500 {
		value = 500
	}
	if value < 0 {
		value = 0
	}

	return value
}

// Função para ler e validar o ID do sensor inserido pelo usuário
func readId(reader *bufio.Reader) string {
	for {
		clearTerminal()
		fmt.Print("\nDigite o ID do sensor de gás: ")
		idStr, _ := reader.ReadString('\n')
		idStr = strings.TrimSpace(idStr)

		// Verifica se o ID digitado contém apenas números
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

/*
	Função principal:

- Lê o ID do sensor
- Recebe o IP do servidor via argumento e conecta via UDP
- Fica gerando e enviando os valores lidos para o servidor periodicamente
*/
func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	serverIP := os.Args[1] // Recebe IP do servidor

	clearTerminal()

	reader := bufio.NewReader(os.Stdin)
	id := readId(reader)
	gas := rand.Intn(101)

	clearTerminal()
	fmt.Printf("\nSensor de gás %s inicializado.\n", id)
	fmt.Println("\nValor: ", gas)

	// Loop infinito para manter a conexão e o envio de dados
	for {
		conn, err := net.Dial("udp", serverIP+":7000") // Cria conexão UDP com servidor
		if err != nil {
			continue
		}

		counter := 0

		// Loop de atualização e envio do sensor
		for {
			gas = step(gas) // Atualiza o valor lido pelo sensor

			if counter >= 1000 {
				// Dados do sensor a serem enviados
				data := Sensor{
					ID:    id,
					Type:  "Gás",
					Value: gas,
				}

				values, _ := json.Marshal(data) // Prepara os dados (JSON) para envio

				// Envia os dados do sensor para o servidor
				_, err := conn.Write(values)
				if err != nil {
					fmt.Println("\nErro no envio do sensor de gás: ", id, err)
					conn.Close()
					break
				}

				buffer := make([]byte, 1024)
				conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // Define um tempo limite para a resposta

				// Lê a resposta do servidor
				n, err := conn.Read(buffer)
				if err != nil {
					fmt.Println("\nServidor não respondeu:", err)
					conn.Close()
					break
				}

				// Aguarda a resposta do servidor sobre o envio/cadastro
				var response Response
				if err := json.Unmarshal(buffer[:n], &response); err != nil {
					fmt.Println("\nErro ao decodificar resposta:", err)
					break
				}

				// Trata erro retornado pelo servidor
				if response.Status == "error" {
					fmt.Println("\n", response.Error)
					fmt.Println("\nPressione ENTER para informar outro ID")
					reader.ReadString('\n')
					id = readId(reader)
					clearTerminal()
					fmt.Printf("\nSensor de gás %s inicializado.\n", id)
					fmt.Println("\nValor: ", gas)
					counter = 0
					continue
				}

				clearTerminal()
				fmt.Printf("\nSensor de gás %s inicializado.\n", id)
				fmt.Println("\nValor: ", gas)
				counter = 0
			}

			counter++
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// Função para limpar o terminal dependendo do sistema operacional
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
