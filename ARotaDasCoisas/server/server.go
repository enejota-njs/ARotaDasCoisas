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

type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error"`
}

var (
	sensors = make(map[string]SensorHistory)
	mu      sync.Mutex
)

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

		mu.Lock()
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
		mu.Unlock()
	}
}

func saveFile() {
	for {
		mu.Lock()
		copySensors := maps.Clone(sensors)
		mu.Unlock()

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

func listSensors(conn net.Conn) {
	mu.Lock()
	copySensors := maps.Clone(sensors)
	mu.Unlock()

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
		mu.Lock()
		copySensors := maps.Clone(sensors)
		mu.Unlock()

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
	mu.Lock()
	copySensors := maps.Clone(sensors)
	mu.Unlock()

	id := fmt.Sprintf("%s", request.ID)

	sensor, ok := copySensors[id]
	if !ok {
		json.NewEncoder(conn).Encode(Response{
			Status: "failed",
			Error:  "Sensor não encontrado",
		})
		return
	}

	var result Sensor
	result.ID = id

	if len(sensor.Humidities) > 0 {
		value := sensor.Humidities[len(sensor.Humidities)-1]
		result.Humidity = &value
	}

	if len(sensor.Temperatures) > 0 {
		value := sensor.Temperatures[len(sensor.Temperatures)-1]
		result.Temperature = &value
	}

	if len(sensor.Luminosities) > 0 {
		value := sensor.Luminosities[len(sensor.Luminosities)-1]
		result.Luminosity = &value
	}

	response := Response{
		Status: "success",
		Data:   result,
	}

	json.NewEncoder(conn).Encode(response)
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

func listenClient() {
	listenerClient, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	defer listenerClient.Close()

	for {
		connClient, err := listenerClient.Accept()
		if err != nil {
			fmt.Println("\nErro na conexão com o cliente:", err)
			continue
		}

		fmt.Println("\nCliente conectado.")
		go handleClient(connClient)
	}
}

func main() {
	fmt.Println("\nServidor inicializado.")

	go listenSensor()
	go saveFile()
	listenClient()
}
