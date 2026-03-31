package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"sync"
	"time"
)

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

type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type Actuator struct {
	Conn net.Conn `json:"conn"`
	ID   string   `json:"id"`
	Type string   `json:"type"`
	On   bool     `json:"on"`
}

var (
	sensors    = make(map[string]Sensor)
	actuators  = make(map[string]Actuator)
	muSensor   sync.Mutex
	muActuator sync.Mutex
)

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
		fmt.Println("\nErro ao enviar command: ", err)
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

/*func listActuators(conn net.Conn) {

}

func verifyActuators(conn net.Conn) {

}

func selectActuator() {

}*/

// == ACTUATOR

func sendActuatorCommand(id, command string) {
	muActuator.Lock()
	actuator, ok := actuators[id]
	muActuator.Unlock()

	if !ok {
		fmt.Println("\nAtuador não encontrado")
		return
	}

	request := Request{
		ID:     id,
		Action: command,
	}

	if err := json.NewEncoder(actuator.Conn).Encode(request); err != nil {
		fmt.Println("\nErro ao enviar comando para atuador: ", err)
		return
	}
}

func actuatorControl() {
	for {
		if !checkListActuators() || !checkListSensors() {
			continue
		}

		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		for id, sensor := range copySensors {
			switch sensor.Type {

			case "Luminosidade":
				if sensor.Value >= 50 {
					sendActuatorCommand(id, "off")
				} else {
					sendActuatorCommand(id, "on")
				}
			case "Umidade":
				if sensor.Value >= 70 {
					sendActuatorCommand(id, "off")
				} else {
					sendActuatorCommand(id, "on")
				}
			case "Temperatura":
				if sensor.Value >= 20 {
					sendActuatorCommand(id, "on")
				} else {
					sendActuatorCommand(id, "off")
				}
			}

		}
		time.Sleep(1 * time.Second)
	}
}

func handleActuator(conn net.Conn) {
	var actuator Actuator

	if err := json.NewDecoder(conn).Decode(&actuator); err != nil {
		fmt.Println("\nErro ao registrar atuador no servidor: ", err)
		conn.Close()
		return
	}

	muActuator.Lock()
	actuators[actuator.ID] = Actuator{
		Conn: conn,
		ID:   actuator.ID,
		Type: actuator.Type,
		On:   actuator.On,
	}
	muActuator.Unlock()

	fmt.Printf("\nAtuador registrado: %s (%s)", actuator.Type, actuator.ID)
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
				Status: "success",
				Data:   sensor,
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
						Status: "success",
						Data:   sensor,
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
					Status: "success",
					Data:   sensor,
				}

				if sendResponse(conn, response) != nil {
					return
				}
			}

			time.Sleep(1 * time.Second)
		}
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
	fmt.Println("\nServidor inicializado.")

	go listenSensor()
	go listenActuator()
	go listenClient()
	go actuatorControl()
	//go saveFile()

	select {}
}

/* Terminando a função do atuador, parece ser bom mudar as funções de request e response pra criar o encoder e decoder dentro dele mesmo, e verificar porque que tá passando o decoder da função request por ponteiro. Parei na função sendActuatorCommand.
