package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// Estrutura de Resposta enviada pelo servidor aos clientes
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

// Estrutura de Requisição recebida ou enviada pelo servidor
type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

// Estrutura interna para armazenar os dados do Atuador juntamente com sua conexão de rede
type ActuatorConn struct {
	Conn net.Conn `json:"conn"`
	ID   string   `json:"id"`
	Type string   `json:"type"`
	On   bool     `json:"on"`
}

// Estrutura representando os dados públicos do Atuador
type Actuator struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	On   bool   `json:"on"`
}

// Variáveis globais para armazenamento de estado, histórico e controle de concorrência
var (
	sensorsHistory     = make(map[string][]int)
	sensors            = make(map[string]Sensor)
	actuators          = make(map[string]ActuatorConn)
	muSensor           sync.Mutex
	muActuator         sync.Mutex
	permissionActuator = make(map[string]bool)
)

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

// == SERVER

// Função para verificar a compatibilidade entre o tipo de sensor e o tipo de atuador
func isCompatible(sensorType, actuatorType string) bool {
	compat := map[string]string{
		"Luminosidade": "Lâmpada",
		"Umidade":      "Umidificador",
		"Temperatura":  "Ar Condicionado",
		"Fumaça":       "Sprinkler",
		"Gás":          "Exaustor",
	}

	expectedActuator, ok := compat[sensorType]
	if !ok {
		return false
	}
	return actuatorType == expectedActuator
}

// Função para receber e decodificar a requisição JSON do cliente
func receiveRequest(decoder *json.Decoder, request *Request) error {
	if err := decoder.Decode(request); err != nil {
		fmt.Println("\nCliente desconectado: ", err)
		return err
	}
	return nil
}

// Função para codificar e enviar uma resposta JSON ao cliente
func sendResponse(conn net.Conn, response Response) error {
	encoder := json.NewEncoder(conn)

	if err := encoder.Encode(response); err != nil {
		fmt.Println("\nErro ao enviar resposta: ", err)
		return err
	}
	return nil
}

// Função para codificar e enviar uma requisição JSON (comando) ao atuador
func sendRequest(conn net.Conn, request Request) error {
	encoder := json.NewEncoder(conn)

	if err := encoder.Encode(request); err != nil {
		fmt.Println("\nErro ao enviar commando: ", err)
		return err
	}
	return nil
}

// Função para verificar se existe algum sensor cadastrado no sistema
func checkListSensors() bool {
	muSensor.Lock()
	copySensors := maps.Clone(sensors)
	muSensor.Unlock()

	if len(copySensors) == 0 {
		return false
	}
	return true
}

// Função para verificar se existe algum atuador conectado ao sistema
func checkListActuators() bool {
	muActuator.Lock()
	copyActuators := maps.Clone(actuators)
	muActuator.Unlock()

	if len(copyActuators) == 0 {
		return false
	}
	return true
}

// Função para enviar um comando (ligar/desligar) a um atuador específico e atualizar seu estado interno
func sendActuatorCommand(id, command string) error {
	// Busca o atuador no mapa
	muActuator.Lock()
	actuator, ok := actuators[id]

	if !ok {
		muActuator.Unlock()
		fmt.Printf("\nAtuador (%s) não encontrado\n", id)
		return fmt.Errorf("\nAtuador (%s) não encontrado", id)
	}

	// Ignora o comando se o atuador já estiver no estado desejado
	if (command == "on" && actuator.On) || (command == "off" && !actuator.On) {
		muActuator.Unlock()
		return nil
	}

	request := Request{
		ID:     id,
		Action: command,
	}

	// Envia a requisição e, em caso de erro na conexão, remove o atuador do mapa
	if sendRequest(actuator.Conn, request) != nil {
		delete(actuators, id)
		fmt.Printf("\nAtuador %s (%s) não encontrado\n", actuator.Type, id)
		muActuator.Unlock()
		return fmt.Errorf("\nAtuador (%s) não encontrado\n", id)
	}

	// Atualiza o estado local do atuador
	switch command {
	case "on":
		actuator.On = true
	case "off":
		actuator.On = false
	}

	actuators[id] = actuator
	muActuator.Unlock()

	return nil
}

// Função executada em loop para monitorar os valores dos sensores e acionar automaticamente os atuadores correspondentes
func actuatorControl() {
	for {
		if !checkListSensors() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Cria uma cópia dos sensores para iteração
		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		for id, sensor := range copySensors {
			muActuator.Lock()
			locked := permissionActuator[id]
			muActuator.Unlock()

			// Ignora o controle automático se o atuador estiver bloqueado (sob controle manual)
			if locked {
				continue
			}

			// Aplica as regras para ligar/desligar os atuadores com base nos valores lidos
			switch sensor.Type {

			case "Luminosidade":
				if sensor.Value < 200 {
					_ = sendActuatorCommand(sensor.ID, "on")
				} else if sensor.Value > 300 {
					_ = sendActuatorCommand(sensor.ID, "off")
				}
			case "Umidade":
				if sensor.Value < 45 {
					_ = sendActuatorCommand(sensor.ID, "on")
				} else if sensor.Value > 55 {
					_ = sendActuatorCommand(sensor.ID, "off")
				}
			case "Temperatura":
				if sensor.Value > 25 {
					_ = sendActuatorCommand(sensor.ID, "on")
				} else if sensor.Value < 20 {
					_ = sendActuatorCommand(sensor.ID, "off")
				}
			case "Fumaça":
				if sensor.Value > 150 {
					_ = sendActuatorCommand(sensor.ID, "on")
				} else if sensor.Value < 80 {
					_ = sendActuatorCommand(sensor.ID, "off")
				}
			case "Gás":
				if sensor.Value > 300 {
					_ = sendActuatorCommand(sensor.ID, "on")
				} else if sensor.Value < 150 {
					_ = sendActuatorCommand(sensor.ID, "off")
				}
			}

		}
		time.Sleep(1 * time.Second)
	}
}

// Função que processa as requisições dos clientes relacionadas aos atuadores
func actuatorClientRequest(conn net.Conn, request Request) {
	// Verifica se existem atuadores disponíveis antes de processar a requisição
	if !checkListActuators() {
		response := Response{
			Status: "error",
			Error:  "Lista de atuadores vazia",
		}

		_ = sendResponse(conn, response)
		return
	}

	switch request.Action {
	case "listActuators":
		muActuator.Lock()
		copyActuators := maps.Clone(actuators)
		muActuator.Unlock()

		// Envia a lista com as informações básicas (ID e Tipo) de todos os atuadores
		for _, actuator := range copyActuators {
			response := Response{
				Status: "success",
				DataActuator: Actuator{
					ID:   actuator.ID,
					Type: actuator.Type,
				},
			}

			if sendResponse(conn, response) != nil {
				return
			}
		}

		response := Response{
			Status: "end",
		}
		_ = sendResponse(conn, response)

	case "verifyActuators":
		start := time.Now()

		// Loop de 10 segundos enviando continuamente o estado de todos os atuadores
		for {
			if time.Since(start) >= 10*time.Second {
				response := Response{
					Status: "end",
				}
				_ = sendResponse(conn, response)
				return
			}

			muActuator.Lock()
			copyActuators := maps.Clone(actuators)
			muActuator.Unlock()

			for _, actuator := range copyActuators {
				response := Response{
					Status: "success",
					DataActuator: Actuator{
						ID:   actuator.ID,
						Type: actuator.Type,
						On:   actuator.On,
					},
				}

				if sendResponse(conn, response) != nil {
					return
				}
			}

			// Marca o fim de uma rodada de envio para o cliente atualizar a tela
			response := Response{
				Status: "endOfRound",
			}

			if sendResponse(conn, response) != nil {
				return
			}

			time.Sleep(1 * time.Second)
		}
	case "selectActuator":
		start := time.Now()

		// Loop de 10 segundos enviando o estado de um atuador específico
		for {
			if time.Since(start) >= 10*time.Second {
				response := Response{
					Status: "end",
				}
				_ = sendResponse(conn, response)
				return
			}

			muActuator.Lock()
			copyActuators := maps.Clone(actuators)
			muActuator.Unlock()

			actuator, ok := copyActuators[request.ID]
			if !ok {
				response := Response{
					Status: "error",
					Error:  "Atuador não encontrado",
				}
				_ = sendResponse(conn, response)
				return
			}

			response := Response{
				Status: "success",
				DataActuator: Actuator{
					ID:   actuator.ID,
					Type: actuator.Type,
					On:   actuator.On,
				},
			}

			if sendResponse(conn, response) != nil {
				return
			}

			time.Sleep(1 * time.Second)
		}

	case "onActuator", "offActuator":
		// Converte a ação recebida para o formato interno de comando (on/off)
		var action string
		if request.Action == "onActuator" {
			action = "on"
		} else if request.Action == "offActuator" {
			action = "off"
		}

		// Envia o comando diretamente ao atuador
		if err := sendActuatorCommand(request.ID, action); err != nil {
			_ = sendResponse(conn, Response{
				Status: "error",
				Error:  err.Error(),
			})
			return
		}

		// Bloqueia o controle automático para este atuador (controle manual acionado)
		muActuator.Lock()
		permissionActuator[request.ID] = true
		actuator := actuators[request.ID]
		muActuator.Unlock()

		response := Response{
			Status: "success",
			DataActuator: Actuator{
				ID:   actuator.ID,
				Type: actuator.Type,
				On:   actuator.On,
			},
		}

		if sendResponse(conn, response) != nil {
			return
		}

		// Goroutine que libera o controle manual do atuador após 10 segundos
		go func(id string) {
			time.Sleep(10 * time.Second)
			muActuator.Lock()
			permissionActuator[id] = false
			muActuator.Unlock()
		}(request.ID)
	}
}

// Função que processa as requisições dos clientes relacionadas aos sensores
func sensorClientRequest(conn net.Conn, request Request) {
	// Verifica se existem sensores disponíveis antes de processar a requisição
	if !checkListSensors() {
		response := Response{
			Status: "error",
			Error:  "Lista de sensores vazia",
		}

		_ = sendResponse(conn, response)
		return
	}

	switch request.Action {
	case "listSensors":
		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		// Envia a lista com as informações de todos os sensores cadastrados
		for _, sensor := range copySensors {
			response := Response{
				Status:     "success",
				DataSensor: sensor,
			}

			if sendResponse(conn, response) != nil {
				return
			}
		}

		response := Response{
			Status: "end",
		}
		_ = sendResponse(conn, response)

	case "verifySensors", "selectSensor":
		start := time.Now()

		// Inicia um loop de monitoramento contínuo por até 10 segundos
		for {
			if time.Since(start) >= 10*time.Second {
				response := Response{
					Status: "end",
				}
				_ = sendResponse(conn, response)
				return
			}

			muSensor.Lock()
			copySensors := maps.Clone(sensors)
			muSensor.Unlock()

			if request.Action == "verifySensors" {
				// Envia o estado atualizado de todos os sensores
				for _, sensor := range copySensors {
					response := Response{
						Status:     "success",
						DataSensor: sensor,
					}
					if sendResponse(conn, response) != nil {
						return
					}
				}

				// Marca o fim de uma rodada de envio para o cliente atualizar a tela
				response := Response{
					Status: "endOfRound",
				}

				if sendResponse(conn, response) != nil {
					return
				}
			} else if request.Action == "selectSensor" {
				// Busca e envia o estado atualizado de um sensor específico
				sensor, ok := copySensors[request.ID]
				if !ok {
					response := Response{
						Status: "error",
						Error:  "Sensor não encontrado",
					}
					_ = sendResponse(conn, response)
					return
				}

				response := Response{
					Status:     "success",
					DataSensor: sensor,
				}

				if sendResponse(conn, response) != nil {
					return
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
}

// == ACTUATOR

// Função que lida com o registro e conexão contínua de um novo atuador ao servidor
func handleActuator(conn net.Conn) {
	decoder := json.NewDecoder(conn)
	var actuator ActuatorConn

	// Recebe os dados de identificação e configuração inicial do atuador
	if err := decoder.Decode(&actuator); err != nil {
		fmt.Println("\nErro ao registrar atuador no servidor: ", err)
		conn.Close()
		return
	}

	muSensor.Lock()
	sensor, sensorExists := sensors[actuator.ID]
	muSensor.Unlock()

	muActuator.Lock()
	_, actuatorExists := actuators[actuator.ID]

	// Verifica se o ID do atuador já está em uso no sistema
	if actuatorExists {
		muActuator.Unlock()
		fmt.Println("\nAtuador já existe")

		response := Response{
			Status: "error",
			Error:  "Atuador já existe",
		}

		if err := json.NewEncoder(conn).Encode(response); err != nil {
			fmt.Println("\nErro ao enviar resposta ao atuador: ", err)
		}

		conn.Close()
		return
	}

	// Verifica compatibilidade caso já exista um sensor com o mesmo ID
	if sensorExists && !isCompatible(sensor.Type, actuator.Type) {
		muActuator.Unlock()
		fmt.Printf("\nErro: atuador %s (%s) incompatível com sensor (%s)\n",
			actuator.ID, actuator.Type, sensor.Type)

		response := Response{
			Status: "error",
			Error:  "Atuador incompatível com o sensor",
		}

		if err := json.NewEncoder(conn).Encode(response); err != nil {
			fmt.Println("\nErro ao enviar resposta ao atuador: ", err)

		}

		conn.Close()
		return
	}

	// Registra oficialmente o atuador, salvando sua conexão de rede
	actuators[actuator.ID] = ActuatorConn{
		Conn: conn,
		ID:   actuator.ID,
		Type: actuator.Type,
		On:   actuator.On,
	}
	muActuator.Unlock()

	fmt.Printf("\nAtuador registrado: %s (%s)\n", actuator.Type, actuator.ID)

	response := Response{
		Status: "success",
	}

	if err := json.NewEncoder(conn).Encode(response); err != nil {
		fmt.Println("\nErro ao enviar resposta ao atuador: ", err)
		conn.Close()
	}

	// Goroutine que mantém a conexão aberta e detecta quando o atuador se desconecta
	go func(id string, c net.Conn) {
		defer c.Close()

		dec := json.NewDecoder(c)
		var message map[string]any

		for {
			if err := dec.Decode(&message); err != nil {
				muActuator.Lock()
				a, ok := actuators[id]
				if ok && a.Conn == c {
					delete(actuators, id)
					fmt.Printf("\nAtuador desconectado: %s\n", id)
				}
				muActuator.Unlock()
				return
			}
		}
	}(actuator.ID, conn)
}

// Função responsável por escutar e aceitar novas conexões TCP de atuadores na porta 9000
func listenActuator() {
	listenerActuator, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}
	defer listenerActuator.Close()

	for {
		connActuator, err := listenerActuator.Accept()
		if err != nil {
			fmt.Println("\nErro na conexão com atuador: ", err)
			continue
		}

		go handleActuator(connActuator)
	}
}

// == SENSOR

// Função responsável por escutar e receber dados contínuos dos sensores via UDP na porta 7000
func listenSensor() {
	bufferSensors := make([]byte, 1024)

	conn, err := net.ListenPacket("udp", ":7000")
	if err != nil {
		fmt.Println("\nErro ao iniciar servidor UDP:", err)
		return
	}
	defer conn.Close()

	for {
		// Lê os dados brutos enviados pelo sensor
		n, addr, err := conn.ReadFrom(bufferSensors)
		if err != nil {
			fmt.Println("\nErro ao se comunicar com sensor: ", err)
			continue
		}

		var received Sensor
		err = json.Unmarshal(bufferSensors[:n], &received)
		if err != nil {
			fmt.Println("\nErro ao descompactar sensor: ", err)
			continue
		}

		muSensor.Lock()
		oldSensor, sensorExists := sensors[received.ID]

		muActuator.Lock()
		actuator, actuatorExists := actuators[received.ID]

		// Impede que um novo sensor utilize um ID de um sensor já existente de outro tipo
		if sensorExists && oldSensor.Type != received.Type {
			muActuator.Unlock()
			muSensor.Unlock()

			fmt.Printf("\nSensor (%s) já existe e é de outro modelo (%s)\n", received.ID, oldSensor.Type)

			response := Response{
				Status: "error",
				Error:  "Sensor já existe e é de outro modelo",
			}
			if b, err := json.Marshal(response); err == nil {
				_, _ = conn.WriteTo(b, addr)
			}
			continue
		}

		// Impede o registro caso exista um atuador com o mesmo ID, mas que seja incompatível
		if actuatorExists && !isCompatible(received.Type, actuator.Type) {
			muActuator.Unlock()
			muSensor.Unlock()

			fmt.Printf("\nSensor (%s) incompatível com atuador (%s)\n",
				received.ID, actuator.Type)

			response := Response{
				Status: "error",
				Error:  "Sensor incompatível com atuador",
			}
			if b, err := json.Marshal(response); err == nil {
				_, _ = conn.WriteTo(b, addr)
			}
			continue
		}

		// Envia a confirmação de recebimento para o sensor
		response := Response{
			Status: "success",
		}
		if b, err := json.Marshal(response); err == nil {
			_, _ = conn.WriteTo(b, addr)
		}

		// Atualiza o histórico e o estado atual do sensor no servidor
		sensorsHistory[received.ID] = append(sensorsHistory[received.ID], received.Value)
		sensors[received.ID] = received

		muActuator.Unlock()
		muSensor.Unlock()

	}
}

// == CLIENT

// Função que mantém a conexão com um cliente e roteia suas requisições
func handleClient(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	// Loop para continuar recebendo requisições enquanto o cliente estiver conectado
	for {
		var request Request

		if receiveRequest(decoder, &request) != nil {
			return
		}

		// Direciona a requisição com base na ação solicitada
		switch request.Action {
		case "listSensors", "verifySensors", "selectSensor":
			sensorClientRequest(conn, request)
		case "listActuators", "verifyActuators", "selectActuator", "onActuator", "offActuator":
			actuatorClientRequest(conn, request)
		}
	}
}

// Função responsável por escutar e aceitar novas conexões TCP de clientes na porta 8000
func listenClient() {
	listenerClient, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	defer listenerClient.Close()

	for {
		connClient, err := listenerClient.Accept()
		if err != nil {
			fmt.Println("\nErro na conexão com o cliente: ", err)
			continue
		}

		fmt.Println("\nCliente conectado")
		// Inicia uma nova goroutine para processar as requisições deste cliente
		go handleClient(connClient)
	}
}

// Função executada em loop para salvar periodicamente o estado atual dos sensores em um arquivo
func saveFile() {
	for {
		// Cria uma cópia segura dos dados dos sensores
		muSensor.Lock()
		copySensors := maps.Clone(sensorsHistory)
		muSensor.Unlock()

		// Cria ou sobrescreve o arquivo JSON de destino
		os.MkdirAll("/data", os.ModePerm)
		file, err := os.Create("/data/data.json")
		if err != nil {
			fmt.Println("\nErro ao criar arquivo JSON.")
			return
		}

		// Formata e grava os dados dos sensores no arquivo
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		encoder.Encode(copySensors)

		file.Close()

		// Aguarda 5 segundos antes de salvar novamente
		time.Sleep(5 * time.Second)
	}
}

// Função principal que inicializa todos os serviços concorrentes do servidor
func main() {
	clearTerminal()
	fmt.Println("\nServidor inicializado")

	// Inicia as rotinas principais de comunicação, controle automático e backup de dados
	go listenSensor()
	go listenActuator()
	go listenClient()
	go actuatorControl()
	go saveFile()

	// Bloqueia a execução principal para manter o servidor rodando indefinidamente
	select {}
}
