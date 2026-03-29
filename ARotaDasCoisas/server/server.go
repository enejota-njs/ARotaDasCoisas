package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"os"
	"sync"
	"time"
)

type ResponseSensor struct {
	Status string `json:"status"`
	Data   Sensor `json:"data"`
	Error  string `json:"error"`
}

type Sensor struct {
	ID          string `json:"id"`
	Temperature *int   `json:"temperature,omitempty"`
	Luminosity  *int   `json:"luminosity,omitempty"`
	Humidity    *int   `json:"humidity,omitempty"`
}

type SensorHistory struct {
	ID           string `json:"id"`
	Temperatures []int  `json:"temperatures,omitempty"`
	Luminosities []int  `json:"luminosities,omitempty"`
	Humidities   []int  `json:"humidities,omitempty"`
}

type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type Actuator struct {
	ID   string `json:"id"`
	On   bool   `json:"on"`
	Type string `json:"type"`
}

type ActuatorConn struct {
	Actuator Actuator `json:"actuator"`
	Conn     net.Conn `json:"conn"`
}

var (
	sensors    = make(map[string]SensorHistory)
	actuators  = make(map[string]ActuatorConn)
	muSensor   sync.Mutex
	muActuator sync.Mutex
)

func listSensors(conn net.Conn) {
	muSensor.Lock()
	copySensors := maps.Clone(sensors)
	muSensor.Unlock()

	for id, s := range copySensors {
		if len(s.Temperatures) == 0 &&
			len(s.Luminosities) == 0 &&
			len(s.Humidities) == 0 {
			continue
		}

		sensor := Sensor{
			ID: id,
		}

		responseSensor := ResponseSensor{
			Status: "success",
			Data:   sensor,
		}

		if err := json.NewEncoder(conn).Encode(responseSensor); err != nil {
			fmt.Println("\nErro ao enviar resposta para o cliente: ", err)
			return
		}
	}

	responseSensor := ResponseSensor{
		Status: "end",
	}

	if err := json.NewEncoder(conn).Encode(responseSensor); err != nil {
		fmt.Println("\nErro ao finalizar resposta do cliente: ", err)
		return
	}
}

func verifySensors(conn net.Conn) {
	for {
		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		for id, s := range copySensors {
			sensor := Sensor{
				ID: id,
			}

			if len(s.Temperatures) > 0 {
				value := s.Temperatures[len(s.Temperatures)-1]
				sensor.Temperature = &value
			}
			if len(s.Luminosities) > 0 {
				value := s.Luminosities[len(s.Luminosities)-1]
				sensor.Luminosity = &value
			}
			if len(s.Humidities) > 0 {
				value := s.Humidities[len(s.Humidities)-1]
				sensor.Humidity = &value
			}

			responseSensor := ResponseSensor{
				Status: "success",
				Data:   sensor,
			}

			if err := json.NewEncoder(conn).Encode(responseSensor); err != nil {
				fmt.Println("\nErro ao enviar resposta para o cliente: ", err)
				return
			}
		}

		responseSensor := ResponseSensor{
			Status: "endOfRound",
		}

		if err := json.NewEncoder(conn).Encode(responseSensor); err != nil {
			fmt.Println("\nErro ao finalizar resposta do cliente: ", err)
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func selectSensor(conn net.Conn, request Request) {
	for {
		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		current, ok := copySensors[request.ID]
		if !ok {
			json.NewEncoder(conn).Encode(ResponseSensor{
				Status: "failed",
				Error:  "Sensor não encontrado",
			})
			return
		}

		sensor := Sensor{
			ID: request.ID,
		}

		if len(current.Humidities) > 0 {
			value := current.Humidities[len(current.Humidities)-1]
			sensor.Humidity = &value
		}

		if len(current.Temperatures) > 0 {
			value := current.Temperatures[len(current.Temperatures)-1]
			sensor.Temperature = &value
		}

		if len(current.Luminosities) > 0 {
			value := current.Luminosities[len(current.Luminosities)-1]
			sensor.Luminosity = &value
		}

		responseSensor := ResponseSensor{
			Status: "success",
			Data:   sensor,
		}

		if err := json.NewEncoder(conn).Encode(responseSensor); err != nil {
			fmt.Println("\nErro ao enviar resposta para o cliente: ", err)
			return
		}

		time.Sleep(1 * time.Second)
	}
}

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
		muSensor.Lock()
		copySensors := maps.Clone(sensors)
		muSensor.Unlock()

		for _, s := range copySensors {
			if len(s.Luminosities) > 0 {
				luminosity := s.Luminosities[len(s.Luminosities)-1]

				if luminosity >= 50 {
					sendActuatorCommand(s.ID, "off")
				} else {
					sendActuatorCommand(s.ID, "on")
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
	actuators[actuator.ID] = ActuatorConn{
		Actuator: actuator,
		Conn:     conn,
	}
	muActuator.Unlock()

	fmt.Printf("\nAtuador registrado: %s (%s)", actuator.Type, actuator.ID)
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	for {
		var request Request

		err := json.NewDecoder(conn).Decode(&request)
		if err != nil {
			fmt.Println("\nErro na requisição do cliente.")
			return
		}

		switch request.Action {
		case "list":
			listSensors(conn)
		case "verify":
			verifySensors(conn)
		case "select":
			selectSensor(conn, request)
		}
	}
}

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
			fmt.Println("\nErro ao se comunicar com sensor:", err)
			continue
		}

		var received Sensor
		err = json.Unmarshal(bufferSensors[:n], &received)
		if err != nil {
			fmt.Println("Erro ao descompactar sensor:", err)
			continue
		}

		muSensor.Lock()
		current := sensors[received.ID]

		if received.Temperature != nil {
			current.Temperatures = append(current.Temperatures, *received.Temperature)
		}
		if received.Luminosity != nil {
			current.Luminosities = append(current.Luminosities, *received.Luminosity)
		}
		if received.Humidity != nil {
			current.Humidities = append(current.Humidities, *received.Humidity)
		}

		current.ID = received.ID
		sensors[received.ID] = current
		muSensor.Unlock()
	}
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

		fmt.Println("\nAtuador conectado.")

		go handleActuator(connActuator)
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

		fmt.Println("\nCliente conectado.")
		go handleClient(connClient)
	}
}

func saveFile() {
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
}

func main() {
	fmt.Println("\nServidor inicializado.")

	go listenSensor()
	go saveFile()
	go listenActuator()
	go actuatorControl()
	listenClient()
}
