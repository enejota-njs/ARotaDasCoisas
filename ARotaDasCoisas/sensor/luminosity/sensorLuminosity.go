package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type Sensor struct {
	ID         string `json:"id"`
	Luminosity int    `json:"luminosity"`
}

func step(value int) int {
	r := rand.Float64()
	if r > 0.5 {
		value += 1
	} else {
		value -= 1
	}

	if value > 400 {
		value = 400
	}
	if value < 300 {
		value = 300
	}
	return value
}

func main() {
	rand.Seed(time.Now().UnixNano())

	id := fmt.Sprintf("Luminosity (%d)", time.Now().Unix())
	lumi := rand.Intn(101) // 300–400

	fmt.Printf("\nSensor de luminosidade %s inicializado.\n", id)

	for {
		conn, err := net.Dial("udp", "127.0.0.1:7000")
		if err != nil {
			fmt.Println("Erro ao conectar o sensor de luminosidade: ", id, err)
			continue
		}

		for {
			lumi = step(lumi)

			data := Sensor{
				ID:         id,
				Luminosity: lumi,
			}

			values, _ := json.Marshal(data)

			_, err := conn.Write(values)
			if err != nil {
				fmt.Println("Erro no envio do sensor de luminosidade: ", id, err)
				conn.Close()
				break
			}

			time.Sleep(1 * time.Millisecond)
		}
	}
}
