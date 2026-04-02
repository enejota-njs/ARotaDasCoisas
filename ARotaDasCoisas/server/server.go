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

func receiveRequest(decoder *json.Decoder, request *Request) error {
	if err := decoder.Decode(request); err != nil {
		fmt.Println("\nErro na requisição: ", err)
		return err
	}
	return nil
} // Finalizada

func sendResponse(conn net.Conn, response Response) error {
	encoder := json.NewEncoder(conn)

	if err := encoder.Encode(response); err != nil {
		fmt.Println("\nErro ao enviar resposta: ", err)
		return err
	}
	return nil
} // Finalizada

func sendRequest(conn net.Conn, request Request) error {
	encoder := json.NewEncoder(conn)

	if err := encoder.Encode(request); err != nil {
		fmt.Println("\nErro ao enviar commando: ", err)
		return err
	}
	return nil
} // Finalizada

func checkListSensors() bool {
	muSensor.Lock()
	copySensors := maps.Clone(sensors)
	muSensor.Unlock()

	if len(copySensors) == 0 {
		return false
	}
	return true
} // Finalizada

func checkListActuators() bool {
	muActuator.Lock()
	copyActuators := maps.Clone(actuators)
	muActuator.Unlock()

	if len(copyActuators) == 0 {
		return false
	}
	return true
} // Finalizada

func sendActuatorCommand(id, command string) error {
	muActuator.Lock()
	actuator, ok := actuators[id]
	if !ok {
		muActuator.Unlock()
		return fmt.Errorf("\nAtuador (%s) não encontrado", id)
	}

	request := Request{
		ID:     id,
		Action: command,
	}

	if sendRequest(actuator.Conn, request) != nil {
		muActuator.Unlock()
		return fmt.Errorf("\nErro encontrado", id)
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
} //Finalizada

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
				if sensor.Value >= 50 {
					_ = sendActuatorCommand(sensor.ID, "off")
				} else {
					_ = sendActuatorCommand(sensor.ID, "on")
				}
			case "Umidade":
				if sensor.Value >= 70 {
					_ = sendActuatorCommand(id, "off")
				} else {
					_ = sendActuatorCommand(id, "on")
				}
			case "Temperatura":
				if sensor.Value >= 20 {
					_ = sendActuatorCommand(id, "on")
				} else {
					_ = sendActuatorCommand(id, "off")
				}
			}

		}
		time.Sleep(1 * time.Second)
	}
} // Finalizada

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

	muActuator.Lock()
	actuators[actuator.ID] = ActuatorConn{
		Conn: conn,
		ID:   actuator.ID,
		Type: actuator.Type,
		On:   actuator.On,
	}
	muActuator.Unlock()

	fmt.Printf("\nAtuador registrado: %s (%s)\n", actuator.Type, actuator.ID)
} // Finalizada

func listenActuator() {
	listenerActuator, err := net.Listen("tcp", "127.0.0.1:9000")
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

		fmt.Println("\nAtuador conectado.")

		go handleActuator(connActuator)
	}
} // Finalizada

// == SENSOR

func listenSensor() {
	bufferSensors := make([]byte, 1024)

	connSensor, err := net.ListenPacket("udp", "127.0.0.1:7000")
	if err != nil {
		fmt.Println("\nErro ao iniciar servidor UDP:", err)
		return
	}
	defer connSensor.Close()

	for {
		n, _, err := connSensor.ReadFrom(bufferSensors)
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
		sensors[received.ID] = received
		muSensor.Unlock()
	}
} // Finalizada

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
	listenerClient, err := net.Listen("tcp", "127.0.0.1:8000")
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

		fmt.Println("\nCliente conectado.")
		go handleClient(connClient)
	}
} // Finalizada

/*func saveFile() {
	for {
		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		file, err := os.Create("../dataBase.json")
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
	fmt.Println("\nServidor inicializado.")

	go listenSensor()
	go listenActuator()
	go listenClient()
	go actuatorControl()
	//go saveFile()

	select {}
}
