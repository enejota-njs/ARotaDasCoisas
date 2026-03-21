package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type Sensor struct {
	Temperature float64 `json:"temperature"`
	Luminosity  float64 `json:"luminosity"`
	Humidity    float64 `json:"humidity"`
}

func main() {
	conn, err := net.Dial("udp", "localhost:5050")
	if err != nil {
		fmt.Println("Erro ao conectar.")
		return
	}
	defer conn.Close()

	fmt.Println("Sensores inicializados.")

	rand.Seed(time.Now().UnixNano())

	for {
		data := Sensor{
			Temperature: rand.Float64(),
			Luminosity:  rand.Float64(),
			Humidity:    rand.Float64(),
		}

		values, _ := json.Marshal(data)

		conn.Write(values)
		time.Sleep(1 * time.Second)
	}
}
