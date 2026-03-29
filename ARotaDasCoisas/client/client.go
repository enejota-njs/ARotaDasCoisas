package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error"`
}

type ResponseSensor struct {
	Status string `json:"status"`
	Data   Sensor `json:"data"`
	Error  string `json:"error"`
}

type Sensor struct {
	ID          string `json:"id"`
	Temperature *int   `json:"temperature"`
	Luminosity  *int   `json:"luminosity"`
	Humidity    *int   `json:"humidity"`
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		fmt.Println("Erro ao conectar no servidor: ", err)
		return
	}
	defer conn.Close()

	fmt.Println("Conectado ao servidor.")

	input := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n|----------- MENU -----------|")
		fmt.Println("| [ 1 ] - Listar sensores    |")
		fmt.Println("| [ 2 ] - Verificar sensores |")
		fmt.Println("| [ 3 ] - Selecionar sensor  |")
		fmt.Println("|----------------------------|")

		fmt.Print("\nSelecione uma opção: ")

		option, _ := input.ReadString('\n')
		option = strings.TrimSpace(option)

		fmt.Print("\033[H\033[2J")
		fmt.Println("")

		var request Request
		var response Response

		switch option {
		case "1":
			request = Request{
				Action: "list",
			}

			if err := json.NewEncoder(conn).Encode(request); err != nil {
				fmt.Println("\nErro ao enviar requisição para o servidor: ", err)
				return
			}

			decoder := json.NewDecoder(conn)

			for {
				var responseSensor ResponseSensor

				if err := decoder.Decode(&responseSensor); err != nil {
					fmt.Println("\nErro ao receber resposta do servidor: ", err)
					return
				}

				if responseSensor.Status == "end" {
					break
				}

				if responseSensor.Status != "success" {
					fmt.Println("\nFalha: ", responseSensor.Error)
					return
				}

				sensorResult := responseSensor.Data

				fmt.Println(sensorResult.ID)
			}

		case "2":
			request = Request{
				Action: "verify",
			}

			if err := json.NewEncoder(conn).Encode(request); err != nil {
				fmt.Println("\nErro ao enviar requisição para o servidor: ", err)
				return
			}

			decoder := json.NewDecoder(conn)
			latest := make(map[string]Sensor)

			for {
				var responseSensor ResponseSensor

				if err := decoder.Decode(&responseSensor); err != nil {
					fmt.Println("\nErro ao receber resposta do servidor: ", err)
					return
				}

				if responseSensor.Status == "endOfRound" {
					fmt.Print("\033[H\033[2J")

					for _, sensor := range latest {
						if sensor.Temperature != nil {
							fmt.Printf("\n%s = %d ", sensor.ID, *sensor.Temperature)
						}
						if sensor.Humidity != nil {
							fmt.Printf("\n%s = %d ", sensor.ID, *sensor.Humidity)
						}
						if sensor.Luminosity != nil {
							fmt.Printf("\n%s = %d ", sensor.ID, *sensor.Luminosity)
						}
					}

					continue
				}

				if responseSensor.Status == "sucess" {
					fmt.Println("\nFalha: ", responseSensor.Error)
					return
				}

				sensorResult := responseSensor.Data
				latest[sensorResult.ID] = sensorResult
			}

		case "3":
			fmt.Print("\nDigite o ID do sensor: ")
			id, _ := input.ReadString('\n')
			id = strings.TrimSpace(id)

			request = Request{
				ID:     id,
				Action: "select",
			}

			if err := json.NewEncoder(conn).Encode(request); err != nil {
				fmt.Println("\nErro ao enviar requisição para o servidor: ", err)
				return
			}

			if err := json.NewDecoder(conn).Decode(&response); err != nil {
				fmt.Println("\nErro ao receber resposta do servidor: ", err)
				return
			}

			if response.Status != "success" {
				fmt.Println("\nFalha: ", response.Error)
				return
			}

			result, ok := response.Data.([]interface{})
			if !ok {
				fmt.Println("\nErro no formato da resposta do servidor.")
				return
			}

			for _, v := range result {
				id, _ := v.(string)
				fmt.Println("s\n", id)
			}
		default:
			fmt.Println("Opção inválida.")
			continue
		}

	}
}
