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

// Estrutura de Requisição do cliente
type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

// Estrutura de Resposta do servidor
type Response struct {
	Status       string   `json:"status"`
	DataSensor   Sensor   `json:"dataSensor"`
	DataActuator Actuator `json:"dataActuator"`
	Error        string   `json:"error"`
}

// Estrutura representando os dados do Sensor
type Sensor struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value int    `json:"value"`
}

// Estrutura representando os dados do Atuador
type Actuator struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	On   bool   `json:"on"`
}

// Função para enviar uma requisição JSON ao servidor
func sendRequest(encoder *json.Encoder, request Request) error {
	if err := encoder.Encode(request); err != nil {
		fmt.Println("\nDesconectado ao servidor")
		return err
	}
	return nil
}

// Função para receber e decodificar a resposta JSON do servidor
func receiveResponse(decoder *json.Decoder, response *Response) error {
	if err := decoder.Decode(response); err != nil {
		fmt.Println("\nDesconectado ao sevidor")
		return err
	}
	return nil
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

// Função para pausar a execução da tela até que o usuário pressione ENTER
func pressEnter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\nPressione ENTER para continuar")
	reader.ReadString('\n')
}

/*
	Função principal do cliente:

- Recebe o IP do servidor e tenta se conectar
- Exibe um menu interativo para visualizar/gerenciar sensores e atuadores
- Envia requisições e recebe respostas do servidor de acordo com a opção escolhida
*/
func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	clearTerminal()
	serverIP := os.Args[1] // Recebe o IP do servidor via argumento
	var conn net.Conn
	var err error

	// Loop para tentar conectar ao servidor continuamente até ter sucesso
	for {
		conn, err = net.Dial("tcp", serverIP+":8000")
		if err != nil {
			fmt.Println("\nServidor não inicializado")
			time.Sleep(1 * time.Second)
			continue
		}

		break
	}

	// Inicializa os encoders/decoders para comunicação JSON e o leitor de input do usuário
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	input := bufio.NewReader(os.Stdin)

	// Loop principal do menu interativo
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

		// Lê a opção escolhida pelo usuário
		option, _ := input.ReadString('\n')
		option = strings.TrimSpace(option)

		var request Request
		var response Response

		switch option {
		case "1", "4":

			var action string

			// Define a ação baseada na escolha do usuário
			if option == "1" {
				action = "listSensors"
			} else if option == "4" {
				action = "listActuators"
			}

			// Prepara a requisição com a ação definida
			request = Request{
				Action: action,
			}

			// Envia a requisição para o servidor
			if sendRequest(encoder, request) != nil {
				conn.Close()
				pressEnter()
				return
			}

			// Loop para receber e processar a lista de itens enviada pelo servidor
			for {
				if receiveResponse(decoder, &response) != nil {
					conn.Close()
					pressEnter()
					return
				}

				// Verifica se o servidor finalizou o envio da lista
				if response.Status == "end" {
					pressEnter()
					break
				}

				// Trata caso de erro retornado pelo servidor
				if response.Status == "error" {
					fmt.Printf("\n%s\n", response.Error)
					pressEnter()
					break
				}

				// Exibe os dados de acordo com a opção escolhida se a resposta for sucesso
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

			// Define a ação baseada na escolha do usuário (monitoramento contínuo)
			if option == "2" {
				action = "verifySensors"
			} else if option == "5" {
				action = "verifyActuators"
			}

			// Prepara a requisição com a ação definida
			request = Request{
				Action: action,
			}

			// Envia a requisição para o servidor
			if sendRequest(encoder, request) != nil {
				conn.Close()
				pressEnter()
				return
			}

			// Mapas para armazenar os dados mais recentes recebidos de cada dispositivo
			latestSensors := make(map[string]Sensor)
			latestActuators := make(map[string]Actuator)

			// Loop para receber e processar os dados continuamente
			for {
				if receiveResponse(decoder, &response) != nil {
					conn.Close()
					pressEnter()
					return
				}

				// Verifica se a rotina de monitoramento foi encerrada pelo servidor
				if response.Status == "end" {
					break
				}

				// Trata caso de erro retornado pelo servidor
				if response.Status == "error" {
					fmt.Printf("\n%s\n", response.Error)
					pressEnter()
					break
				}

				// Verifica se um ciclo de envio de dados foi concluído para atualizar a tela
				if response.Status == "endOfRound" {
					clearTerminal()

					if option == "2" {
						fmt.Println("\nSensores:")

						// Converte o mapa para um slice para permitir a ordenação
						orderedSensors := make([]Sensor, 0, len(latestSensors))
						for _, s := range latestSensors {
							orderedSensors = append(orderedSensors, s)
						}

						// Ordena os sensores primeiro por Tipo e depois por ID
						sort.Slice(orderedSensors, func(i, j int) bool {
							if orderedSensors[i].Type == orderedSensors[j].Type {
								return orderedSensors[i].ID < orderedSensors[j].ID
							}
							return orderedSensors[i].Type < orderedSensors[j].Type
						})

						// Percorre os sensores ordenados e define as unidades de medida corretas
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

						// Converte o mapa para um slice para permitir a ordenação
						orderedActuators := make([]Actuator, 0, len(latestActuators))
						for _, a := range latestActuators {
							orderedActuators = append(orderedActuators, a)
						}

						// Ordena os atuadores primeiro por Tipo e depois por ID
						sort.Slice(orderedActuators, func(i, j int) bool {
							if orderedActuators[i].Type == orderedActuators[j].Type {
								return orderedActuators[i].ID < orderedActuators[j].ID
							}
							return orderedActuators[i].Type < orderedActuators[j].Type
						})

						// Percorre os atuadores ordenados e define a string de estado (Ligado/Desligado)
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

				// Atualiza os mapas com os dados mais recentes quando o recebimento é um sucesso
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
			// Opção para selecionar e monitorar um sensor específico
			fmt.Print("\nDigite o ID do sensor: ")
			id, _ := input.ReadString('\n')
			id = strings.TrimSpace(id)

			// Prepara a requisição com a ação e o ID fornecido
			request = Request{
				ID:     id,
				Action: "selectSensor",
			}

			// Envia a requisição para o servidor
			if sendRequest(encoder, request) != nil {
				conn.Close()
				pressEnter()
				return
			}

			// Loop contínuo para receber e atualizar os dados do sensor selecionado
			for {
				if receiveResponse(decoder, &response) != nil {
					conn.Close()
					pressEnter()
					return
				}

				// Verifica se a rotina de monitoramento do sensor foi encerrada
				if response.Status == "end" {
					break
				}

				// Trata caso de erro retornado pelo servidor
				if response.Status == "error" {
					fmt.Printf("\n%s\n", response.Error)
					pressEnter()
					break
				}

				// Quando recebe os dados com sucesso, atualiza a exibição na tela
				if response.Status == "success" {
					clearTerminal()

					sensor := response.DataSensor

					fmt.Println("\nSensor: ")
					unit := ""

					// Define a unidade de medida com base no tipo do sensor
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

			// Exibe o submenu para interagir com um atuador específico
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
				// Opção para verificar o estado de um atuador continuamente
				fmt.Print("\nDigite o ID do atuador: ")
				id, _ := input.ReadString('\n')
				id = strings.TrimSpace(id)

				// Prepara a requisição para selecionar o atuador
				request = Request{
					ID:     id,
					Action: "selectActuator",
				}

				// Envia a requisição para o servidor
				if sendRequest(encoder, request) != nil {
					conn.Close()
					pressEnter()
					return
				}

				// Loop contínuo para receber e atualizar o estado do atuador selecionado
				for {
					if receiveResponse(decoder, &response) != nil {
						conn.Close()
						pressEnter()
						return
					}

					// Verifica se a rotina de monitoramento foi encerrada
					if response.Status == "end" {
						break
					}

					// Trata caso de erro retornado pelo servidor
					if response.Status == "error" {
						fmt.Printf("\n%s\n", response.Error)
						pressEnter()
						break
					}

					// Quando recebe os dados com sucesso, atualiza a exibição na tela
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
				// Opção para alterar o estado (ligar/desligar) de um atuador
				fmt.Print("\nDigite o ID do atuador: ")
				id, _ := input.ReadString('\n')
				id = strings.TrimSpace(id)

				clearTerminal()

				// Exibe submenu para escolher a ação de ligar ou desligar
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

				// Valida a opção e prepara a ação correspondente
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

				// Redefine a requisição com base na escolha
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

				// Envia o comando para o servidor
				if sendRequest(encoder, request) != nil {
					conn.Close()
					pressEnter()
					return
				}

				// Aguarda a resposta única do servidor sobre a alteração de estado
				if receiveResponse(decoder, &response) != nil {
					conn.Close()
					pressEnter()
					return
				}

				// Trata caso de erro retornado pelo servidor
				if response.Status == "error" {
					fmt.Printf("\n%s\n", response.Error)
					pressEnter()
					break
				}

				// Exibe o novo estado do atuador caso a alteração seja bem-sucedida
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
				continue // Volta ao menu principal
			default:
				fmt.Println("\nOpção inválida.")
				pressEnter()
				continue
			}

		case "7":
			// Opção para finalizar a sessão e encerrar o programa
			fmt.Println("\nSessão finalizada")
			conn.Close()
			pressEnter()
			return

		default:
			// Tratamento para opção inválida no menu principal
			fmt.Println("\nOpção inválida.")
			pressEnter()
			continue
		}
	}
}
