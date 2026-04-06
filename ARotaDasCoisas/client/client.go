package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
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
		fmt.Println("\nDesconectado ao servidor")
		return err
	}
	return nil
}

func receiveResponse(decoder *json.Decoder, response *Response) error {
	if err := decoder.Decode(response); err != nil {
		fmt.Println("\nDesconectado ao sevidor")
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
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\nPressione ENTER para continuar")
	reader.ReadString('\n')
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	clearTerminal()
	serverIP := os.Args[1]
	var conn net.Conn
	var err error

	for {
		conn, err = net.Dial("tcp", serverIP+":8000")
		if err != nil {
			fmt.Println("\nServidor não inicializado")
			time.Sleep(1 * time.Second)
			continue
		}

		break
	}

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
				conn.Close()
				pressEnter()
				return
			}

			for {
				if receiveResponse(decoder, &response) != nil {
					conn.Close()
					pressEnter()
					return
				}

				if response.Status == "end" {
					pressEnter()
					break
				}

				if response.Status == "error" {
					fmt.Printf("\n%s\n", response.Error)
					pressEnter()
					break
				}

				if response.Status == "success" {
					if option == "1" {
						sensor := response.DataSensor
						fmt.Printf("\n- %s (%s)\n", sensor.Type, sensor.ID)
					} else if option == "4" {
						actuator := response.DataActuator
						fmt.Printf("\n- %s (%s)\n", actuator.Type, actuator.ID)
					}
				}
			}

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
				conn.Close()
				pressEnter()
				return
			}

			latestSensors := make(map[string]Sensor)
			latestActuators := make(map[string]Actuator)

			for {
				if receiveResponse(decoder, &response) != nil {
					conn.Close()
					pressEnter()
					return
				}

				if response.Status == "end" {
					break
				}

				if response.Status == "error" {
					fmt.Printf("\n%s\n", response.Error)
					pressEnter()
					break
				}

				if response.Status == "endOfRound" {
					clearTerminal()

					if option == "2" {
						fmt.Println("\nSensores:")

						orderedSensors := make([]Sensor, 0, len(latestSensors))
						for _, s := range latestSensors {
							orderedSensors = append(orderedSensors, s)
						}

						sort.Slice(orderedSensors, func(i, j int) bool {
							if orderedSensors[i].Type == orderedSensors[j].Type {
								return orderedSensors[i].ID < orderedSensors[j].ID
							}
							return orderedSensors[i].Type < orderedSensors[j].Type
						})

						for _, sensor := range orderedSensors {
							unit := ""

							if sensor.Type == "Luminosidade" {
								unit = "lux"
							}
							if sensor.Type == "Umidade" {
								unit = "%"
							}
							if sensor.Type == "Temperatura" {
								unit = "°C"
							}
							if sensor.Type == "Fumaça" {
								unit = "ppm"
							}
							if sensor.Type == "Gás" {
								unit = "ppm"
							}

							fmt.Printf("\n- %s (%s) = %d %s", sensor.Type, sensor.ID, sensor.Value, unit)
						}
					} else if option == "5" {
						fmt.Println("\nAtuadores: ")

						orderedActuators := make([]Actuator, 0, len(latestActuators))
						for _, a := range latestActuators {
							orderedActuators = append(orderedActuators, a)
						}

						sort.Slice(orderedActuators, func(i, j int) bool {
							if orderedActuators[i].Type == orderedActuators[j].Type {
								return orderedActuators[i].ID < orderedActuators[j].ID
							}
							return orderedActuators[i].Type < orderedActuators[j].Type
						})

						for _, actuator := range orderedActuators {
							on := "Desligado"
							if actuator.On {
								on = "Ligado"
							}
							fmt.Printf("\n- %s (%s) = %s", actuator.Type, actuator.ID, on)
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
				conn.Close()
				pressEnter()
				return
			}

			for {
				if receiveResponse(decoder, &response) != nil {
					conn.Close()
					pressEnter()
					return
				}

				if response.Status == "end" {
					break
				}

				if response.Status == "error" {
					fmt.Printf("\n%s\n", response.Error)
					pressEnter()
					break
				}

				if response.Status == "success" {
					clearTerminal()

					sensor := response.DataSensor

					fmt.Println("\nSensor: ")
					unit := ""

					if sensor.Type == "Luminosidade" {
						unit = "lux"
					}
					if sensor.Type == "Umidade" {
						unit = "%"
					}
					if sensor.Type == "Temperatura" {
						unit = "°C"
					}
					if sensor.Type == "Fumaça" {
						unit = "ppm"
					}
					if sensor.Type == "Gás" {
						unit = "ppm"
					}

					fmt.Printf("\n- %s (%s) = %d %s", sensor.Type, sensor.ID, sensor.Value, unit)
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

			switch optionAc {
			case "1":
				fmt.Print("\nDigite o ID do atuador: ")
				id, _ := input.ReadString('\n')
				id = strings.TrimSpace(id)

				request = Request{
					ID:     id,
					Action: "selectActuator",
				}

				if sendRequest(encoder, request) != nil {
					conn.Close()
					pressEnter()
					return
				}

				for {
					if receiveResponse(decoder, &response) != nil {
						conn.Close()
						pressEnter()
						return
					}

					if response.Status == "end" {
						break
					}

					if response.Status == "error" {
						fmt.Printf("\n%s\n", response.Error)
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
						fmt.Printf("\n- %s (%s) = %s", actuator.Type, actuator.ID, on)
					}
				}
			case "2":
				fmt.Print("\nDigite o ID do atuador: ")
				id, _ := input.ReadString('\n')
				id = strings.TrimSpace(id)

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

				switch optionPower {
				case "1":
					request = Request{
						ID:     id,
						Action: "onActuator",
					}
				case "2":
					request = Request{
						ID:     id,
						Action: "offActuator",
					}
				case "3":
					continue
				default:
					fmt.Println("\nOpção inválida.")
					pressEnter()
					continue
				}

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
					conn.Close()
					pressEnter()
					return
				}

				if receiveResponse(decoder, &response) != nil {
					conn.Close()
					pressEnter()
					return
				}

				if response.Status == "error" {
					fmt.Printf("\n%s\n", response.Error)
					pressEnter()
					break
				}

				if response.Status == "success" {
					actuator := response.DataActuator

					on := "Desligado"
					if actuator.On {
						on = "Ligado"
					}
					fmt.Printf("\n- %s (%s) = %s\n", actuator.Type, actuator.ID, on)
				}

				pressEnter()
			case "3":
				continue
			default:
				fmt.Println("\nOpção inválida.")
				pressEnter()
				continue
			}
		case "7":
			fmt.Println("\nSessão finalizada")
			conn.Close()
			pressEnter()
			return

		default:
			fmt.Println("\nOpção inválida.")
			pressEnter()
			continue
		}
	}
}
