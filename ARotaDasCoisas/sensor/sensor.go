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
	fmt.Println("Sensores inicializados.")

	rand.Seed(time.Now().UnixNano())

	for {
		conn, err := net.Dial("udp", "127.0.0.1:7070")
		if err != nil {
			fmt.Println("Erro ao conectar: ", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for {
			data := Sensor{
				Temperature: rand.Float64(),
				Luminosity:  rand.Float64(),
				Humidity:    rand.Float64(),
			}

			values, _ := json.Marshal(data)

			_, err := conn.Write(values)
			if err != nil {
				fmt.Println("Erro no envio:", err)
				conn.Close()
				break
			}

			time.Sleep(1 * time.Second)
		}
	}
}
