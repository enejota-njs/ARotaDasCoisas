package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

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

type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type Response struct {
	Status string `json:"status"`
	Data   Sensor `json:"data"`
	Error  string `json:"error"`
}

type Sensor struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func sendRequest(encoder *json.Encoder, request Request) error {
	if err := encoder.Encode(request); err != nil {
		fmt.Println("\nErro ao enviar requisição: ", err)
		return err
	}
	return nil
}

func receiveResponse(decoder *json.Decoder, response *Response) error {
	if err := decoder.Decode(response); err != nil {
		fmt.Println("\nErro na resposta do servidor: ", err)
		return err
	}
	return nil
}

func pressEnter() {
	fmt.Println("\n\nPressione ENTER para continuar.")
	fmt.Scanln()
}

func main() {
	clearTerminal()
	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println("\nErro ao conectar no servidor: ", err)
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	input := bufio.NewReader(os.Stdin)

	for {
		clearTerminal()
		fmt.Println("\n|------------------------------|")
		fmt.Println("|             MENU             |")
		fmt.Println("|------------------------------|")
		fmt.Println("|           SENSORES           |")
		fmt.Println("|                              |")
		fmt.Println("| [ 1 ] - Listar sensores      |")
		fmt.Println("| [ 2 ] - Verificar sensores   |")
		fmt.Println("| [ 3 ] - Selecionar sensor    |")
		fmt.Println("|                              |")
		fmt.Println("|------------------------------|")
		fmt.Println("|           ATUADORES          |")
		fmt.Println("|                              |")
		fmt.Println("| [ 4 ] - Listar atuadores     |")
		fmt.Println("| [ 5 ] - Verificar atuadores  |")
		fmt.Println("| [ 6 ] - Selecionar atuador   |")
		fmt.Println("|                              |")
		fmt.Println("|------------------------------|")
		fmt.Println("| [ 7 ] - Fechar               |")
		fmt.Println("|------------------------------|")

		fmt.Print("\nSelecione uma opção: ")

		option, _ := input.ReadString('\n')
		option = strings.TrimSpace(option)

		var request Request
		var response Response

		switch option {
		case "1":
			request = Request{
				Action: "listSensors",
			}

			if sendRequest(encoder, request) != nil {
				pressEnter()
				continue
			}

			for {
				if receiveResponse(decoder, &response) != nil {
					pressEnter()
					break
				}

				if response.Status == "end" {
					pressEnter()
					break
				}

				if response.Status == "error" {
					fmt.Println("\nErro: ", response.Error)
					pressEnter()
					break
				}

				if response.Status == "success" {
					sensor := response.Data
					fmt.Printf("\n%s (%s)", sensor.Type, sensor.ID)
				}
			}

			fmt.Println("\n\n")

		case "2":
			request = Request{
				Action: "verifySensors",
			}

			if sendRequest(encoder, request) != nil {
				pressEnter()
				continue
			}

			latest := make(map[string]Sensor)

			for {
				if receiveResponse(decoder, &response) != nil {
					pressEnter()
					break
				}

				if response.Status == "end" {
					break
				}

				if response.Status == "error" {
					fmt.Println("\nErro: ", response.Error)
					pressEnter()
					break
				}

				if response.Status == "endOfRound" {
					clearTerminal()
					fmt.Println("\nSensores: ")

					for _, sensor := range latest {
						fmt.Printf("\n%s (%s) = %d", sensor.Type, sensor.ID, sensor.Value)
					}

					continue
				}

				if response.Status == "success" {
					sensor := response.Data
					latest[sensor.ID] = sensor
				}
			}

		case "3":
			fmt.Print("\nDigite o ID do sensor: ")
			id, _ := input.ReadString('\n')
			id = strings.TrimSpace(id)

			request = Request{
				ID:     id,
				Action: "selectSensor",
			}

			if sendRequest(encoder, request) != nil {
				pressEnter()
				continue
			}

			for {
				if receiveResponse(decoder, &response) != nil {
					pressEnter()
					break
				}

				if response.Status == "end" {
					break
				}

				if response.Status == "error" {
					fmt.Println("\nErro: ", response.Error)
					pressEnter()
					break
				}

				if response.Status == "success" {
					clearTerminal()

					sensor := response.Data

					fmt.Println("\nSensor: ")
					fmt.Printf("\n%s (%s) = %d", sensor.Type, sensor.ID, sensor.Value)
				}
			}
		case "4":
		case "5":
		case "6":
		case "7":
			conn.Close()
			return

		default:
			fmt.Println("\nOpção inválida.")
			pressEnter()
			continue
		}
	}
}
