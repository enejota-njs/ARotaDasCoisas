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

type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type Response struct {
	Status       string   `json:"status"`
	DataSensor   Sensor   `json:"dataSensor"`
	DataActuator Actuator `json:"dataActuator"`
	Error        string   `json:"error"`
}

type Sensor struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type Actuator struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	On   bool   `json:"on"`
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
		fmt.Println("\n|--------------------------------|")
		fmt.Println("|              MENU              |")
		fmt.Println("|--------------------------------|")
		fmt.Println("|            SENSORES            |")
		fmt.Println("|                                |")
		fmt.Println("| [ 1 ] - Listar sensores        |")
		fmt.Println("| [ 2 ] - Verificar sensores     |")
		fmt.Println("| [ 3 ] - Selecionar sensor      |")
		fmt.Println("|                                |")
		fmt.Println("|--------------------------------|")
		fmt.Println("|            ATUADORES           |")
		fmt.Println("|                                |")
		fmt.Println("| [ 4 ] - Listar atuadores       |")
		fmt.Println("| [ 5 ] - Verificar atuadores    |")
		fmt.Println("| [ 6 ] - Selecionar atuador     |")
		fmt.Println("|                                |")
		fmt.Println("|--------------------------------|")
		fmt.Println("| [ 7 ] - Fechar                 |")
		fmt.Println("|--------------------------------|")

		fmt.Print("\nSelecione uma opção: ")

		option, _ := input.ReadString('\n')
		option = strings.TrimSpace(option)

		var request Request
		var response Response

		switch option {
		case "1", "4":

			var action string

			if option == "1" {
				action = "listSensors"
			} else if option == "4" {
				action = "listActuators"
			}

			request = Request{
				Action: action,
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
					if option == "1" {
						sensor := response.DataSensor
						fmt.Printf("\n%s (%s)", sensor.Type, sensor.ID)
					} else if option == "4" {
						actuator := response.DataActuator
						fmt.Printf("\n%s (%s)", actuator.Type, actuator.ID)
					}
				}
			}

			fmt.Println("\n\n")

		case "2", "5":

			var action string

			if option == "2" {
				action = "verifySensors"
			} else if option == "5" {
				action = "verifyActuators"
			}
			request = Request{
				Action: action,
			}

			if sendRequest(encoder, request) != nil {
				pressEnter()
				continue
			}

			latestSensors := make(map[string]Sensor)
			latestActuators := make(map[string]Actuator)

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

					if option == "2" {
						fmt.Println("\nSensores: ")
						for _, sensor := range latestSensors {
							fmt.Printf("\n%s (%s) = %d", sensor.Type, sensor.ID, sensor.Value)
						}
					} else if option == "5" {
						fmt.Println("\nAtuadores: ")
						for _, actuator := range latestActuators {
							on := "Desligado"
							if actuator.On {
								on = "Ligado"
							}
							fmt.Printf("\n%s (%s) = %s", actuator.Type, actuator.ID, on)
						}
					}

					continue
				}

				if response.Status == "success" {
					if option == "2" {
						sensor := response.DataSensor
						latestSensors[sensor.ID] = sensor
					} else if option == "5" {
						actuator := response.DataActuator
						latestActuators[actuator.ID] = actuator
					}
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

					sensor := response.DataSensor

					fmt.Println("\nSensor: ")
					fmt.Printf("\n%s (%s) = %d", sensor.Type, sensor.ID, sensor.Value)
				}
			}
		case "6":
			clearTerminal()

			fmt.Println("\n|--------------------------------|")
			fmt.Println("|       SELECIONAR ATUADOR       |")
			fmt.Println("|--------------------------------|")
			fmt.Println("|                                |")
			fmt.Println("| [ 1 ] - Verificar atuador      |")
			fmt.Println("| [ 2 ] - Ligar/Desligar atuador |")
			fmt.Println("|                                |")
			fmt.Println("|--------------------------------|")
			fmt.Println("| [ 3 ] - Voltar                 |")
			fmt.Println("|--------------------------------|")

			fmt.Print("\nSelecione uma opção: ")
			optionAc, _ := input.ReadString('\n')
			optionAc = strings.TrimSpace(optionAc)

			fmt.Print("\nDigite o ID do atuador: ")
			id, _ := input.ReadString('\n')
			id = strings.TrimSpace(id)

			switch optionAc {
			case "1":
				request = Request{
					ID:     id,
					Action: "selectActuator",
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

						actuator := response.DataActuator

						fmt.Println("\nAtuador: ")
						on := "Desligado"
						if actuator.On {
							on = "Ligado"
						}
						fmt.Printf("\n%s (%s) = %s", actuator.Type, actuator.ID, on)
					}
				}
			case "2":
				clearTerminal()

				fmt.Println("\n|--------------------------------|")
				fmt.Println("|     LIGAR/DESLIGAR ATUADOR     |")
				fmt.Println("|--------------------------------|")
				fmt.Println("|                                |")
				fmt.Println("| [ 1 ] - Ligar atuador          |")
				fmt.Println("| [ 2 ] - Desligar atuador       |")
				fmt.Println("|                                |")
				fmt.Println("|--------------------------------|")
				fmt.Println("| [ 3 ] - Voltar                 |")
				fmt.Println("|--------------------------------|")

				fmt.Print("\nSelecione uma opção: ")
				optionPower, _ := input.ReadString('\n')
				optionPower = strings.TrimSpace(optionPower)

				if optionPower == "1" {
					request = Request{
						ID:     id,
						Action: "onActuator",
					}
				} else if optionPower == "2" {
					request = Request{
						ID:     id,
						Action: "offActuator",
					}
				}

				if sendRequest(encoder, request) != nil {
					pressEnter()
					continue
				}

				if receiveResponse(decoder, &response) != nil {
					pressEnter()
					break
				}

				if response.Status == "error" {
					fmt.Println("\nErro: ", response.Error)
					pressEnter()
					break
				}

				if response.Status == "success" {
					actuator := response.DataActuator

					on := "Desligado"
					if actuator.On {
						on = "Ligado"
					}
					fmt.Printf("\n%s (%s) = %s", actuator.Type, actuator.ID, on)
				}

				pressEnter()
			}
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
