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

type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type ActuatorConn struct {
	Conn net.Conn `json:"conn"`
	ID   string   `json:"id"`
	Type string   `json:"type"`
	On   bool     `json:"on"`
}

type Actuator struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	On   bool   `json:"on"`
}

var (
	sensors            = make(map[string]Sensor)
	actuators          = make(map[string]ActuatorConn)
	muSensor           sync.Mutex
	muActuator         sync.Mutex
	permissionActuator = make(map[string]bool)
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

// == SERVER

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

func receiveRequest(decoder *json.Decoder, request *Request) error {
	if err := decoder.Decode(request); err != nil {
		fmt.Println("\nCliente desconectado: ", err)
		return err
	}
	return nil
}

func sendResponse(conn net.Conn, response Response) error {
	encoder := json.NewEncoder(conn)

	if err := encoder.Encode(response); err != nil {
		fmt.Println("\nErro ao enviar resposta: ", err)
		return err
	}
	return nil
}

func sendRequest(conn net.Conn, request Request) error {
	encoder := json.NewEncoder(conn)

	if err := encoder.Encode(request); err != nil {
		fmt.Println("\nErro ao enviar commando: ", err)
		return err
	}
	return nil
}

func checkListSensors() bool {
	muSensor.Lock()
	copySensors := maps.Clone(sensors)
	muSensor.Unlock()

	if len(copySensors) == 0 {
		return false
	}
	return true
}

func checkListActuators() bool {
	muActuator.Lock()
	copyActuators := maps.Clone(actuators)
	muActuator.Unlock()

	if len(copyActuators) == 0 {
		return false
	}
	return true
}

func sendActuatorCommand(id, command string) error {
	muActuator.Lock()
	actuator, ok := actuators[id]

	if !ok {
		muActuator.Unlock()
		fmt.Printf("\nAtuador (%s) não encontrado\n", id)
		return fmt.Errorf("\nAtuador (%s) não encontrado", id)
	}

	if (command == "on" && actuator.On) || (command == "off" && !actuator.On) {
		muActuator.Unlock()
		return nil
	}

	request := Request{
		ID:     id,
		Action: command,
	}

	if sendRequest(actuator.Conn, request) != nil {
		delete(actuators, id)
		fmt.Printf("\nAtuador %s (%s) não encontrado\n", actuator.Type, id)
		muActuator.Unlock()
		return fmt.Errorf("\nAtuador (%s) não encontrado\n", id)
	}

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

func actuatorControl() {
	for {
		if !checkListSensors() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		for id, sensor := range copySensors {
			muActuator.Lock()
			locked := permissionActuator[id]
			muActuator.Unlock()

			if locked {
				continue
			}

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

func actuatorClientRequest(conn net.Conn, request Request) {
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
		var action string
		if request.Action == "onActuator" {
			action = "on"
		} else if request.Action == "offActuator" {
			action = "off"
		}

		if err := sendActuatorCommand(request.ID, action); err != nil {
			_ = sendResponse(conn, Response{
				Status: "error",
				Error:  err.Error(),
			})
			return
		}

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

		go func(id string) {
			time.Sleep(10 * time.Second)
			muActuator.Lock()
			permissionActuator[id] = false
			muActuator.Unlock()
		}(request.ID)
	}
}

func sensorClientRequest(conn net.Conn, request Request) {
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
					Status: "endOfRound",
				}

				if sendResponse(conn, response) != nil {
					return
				}
			} else if request.Action == "selectSensor" {
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

func handleActuator(conn net.Conn) {
	decoder := json.NewDecoder(conn)
	var actuator ActuatorConn

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

func listenSensor() {
	bufferSensors := make([]byte, 1024)

	conn, err := net.ListenPacket("udp", ":7000")
	if err != nil {
		fmt.Println("\nErro ao iniciar servidor UDP:", err)
		return
	}
	defer conn.Close()

	for {
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

		response := Response{
			Status: "success",
		}
		if b, err := json.Marshal(response); err == nil {
			_, _ = conn.WriteTo(b, addr)
		}

		sensors[received.ID] = received

		muActuator.Unlock()
		muSensor.Unlock()

	}
}

// == CLIENT

func handleClient(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	for {
		var request Request

		if receiveRequest(decoder, &request) != nil {
			return
		}

		switch request.Action {
		case "listSensors", "verifySensors", "selectSensor":
			sensorClientRequest(conn, request)
		case "listActuators", "verifyActuators", "selectActuator", "onActuator", "offActuator":
			actuatorClientRequest(conn, request)
		}
	}
}

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
		go handleClient(connClient)
	}
}

/*func saveFile() {
	for {
		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		file, err := os.Create("../data.json")
		if err != nil {
			fmt.Println("\nErro ao criar arquivo JSON.")
			return
		}

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		encoder.Encode(copySensors)

		file.Close()

		time.Sleep(5 * time.Second)
	}
}*/

func main() {
	clearTerminal()
	fmt.Println("\nServidor inicializado")

	go listenSensor()
	go listenActuator()
	go listenClient()
	go actuatorControl()
	//go saveFile()

	select {}
}
