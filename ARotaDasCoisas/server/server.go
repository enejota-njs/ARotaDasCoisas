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

type Sensor struct {
	ID          string `json:"id"`
	Temperature *int   `json:"temperature"`
	Luminosity  *int   `json:"luminosity"`
	Humidity    *int   `json:"humidity"`
}

type SensorHistory struct {
	ID           string `json:"id"`
	Temperatures []int  `json:"temperatures"`
	Luminosities []int  `json:"luminosities"`
	Humidities   []int  `json:"humidities"`
}

type Request struct {
	ID     int    `json:"id"`
	Action string `json:"action"`
}

type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
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

		file, err := os.Create("readings.json")
		if err != nil {
			fmt.Println("\nErro ao criar arquivo JSON.")
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		encoder.Encode(copySensors)

		time.Sleep(5 * time.Second)
	}
}

func listSensors(conn net.Conn) {
	mu.Lock()
	copySensors := maps.Clone(sensors)
	mu.Unlock()

	var ids []string

	for id, s := range copySensors {
		if len(s.Temperatures) > 0 ||
			len(s.Luminosities) > 0 ||
			len(s.Humidities) > 0 {

			ids = append(ids, id)
		}
	}

	response := Response{
		Status: "sucess",
		Data:   ids,
	}

	json.NewEncoder(conn).Encode(response)
}

func verifySensors(conn net.Conn) {
	for {
		mu.Lock()
		copySensors := maps.Clone(sensors)
		mu.Unlock()

		var result []Sensor

		for id, s := range copySensors {
			var sensor Sensor
			sensor.ID = id

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

			result = append(result, sensor)
		}

		response := Response{
			Status: "success",
			Data:   result,
		}

		json.NewEncoder(conn).Encode(response)

		time.Sleep(1 * time.Second)
	}
}

func selectSensor(conn net.Conn, request Request) {
	mu.Lock()
	copySensors := maps.Clone(sensors)
	mu.Unlock()

	id := fmt.Sprintf("%d", request.ID)

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
