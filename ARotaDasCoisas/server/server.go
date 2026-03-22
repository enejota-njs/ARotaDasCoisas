package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type Sensor struct {
	Temperature float64 `json:"temperature"`
	Luminosity  float64 `json:"luminosity"`
	Humidity    float64 `json:"humidity"`
}

var (
	sensors Sensor
	mu      sync.Mutex
)

func listenSensor() {
	bufferSensors := make([]byte, 1024)

	connSensor, err := net.ListenPacket("udp", "127.0.0.1:7070")
	if err != nil {
		fmt.Println("Erro ao iniciar servidor UDP:", err)
		return
	}
	defer connSensor.Close()

	for {
		n, _, err := connSensor.ReadFrom(bufferSensors)
		if err != nil {
			fmt.Println("Erro no ReadFrom:", err)
			continue
		}

		var received Sensor
		err = json.Unmarshal(bufferSensors[:n], &received)
		if err != nil {
			fmt.Println("Erro no Unmarshal:", err)
			continue
		}

		mu.Lock()
		sensors = received
		mu.Unlock()
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	for {
		mu.Lock()
		current := sensors
		mu.Unlock()

		values := fmt.Sprintf(
			"Temperatura: %.2f | Lumimosidade: %.2f | Umidade: %.2f\n",
			current.Temperature,
			current.Luminosity,
			current.Humidity,
		)

		_, err := conn.Write([]byte(values))
		if err != nil {
			fmt.Println("Erro no envio:", err)
		}

		time.Sleep(1 * time.Second)
	}

}

func listenClient() {
	listenerClient, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listenerClient.Close()

	for {
		connClient, err := listenerClient.Accept()
		if err != nil {
			fmt.Println("Erro no Accept:", err)
			continue
		}

		fmt.Println("Cliente conectado.")
		go handleClient(connClient)
	}
}

func main() {
	fmt.Println("Servidor inicializado.")

	go listenSensor()
	go listenClient()

	select {}
}
