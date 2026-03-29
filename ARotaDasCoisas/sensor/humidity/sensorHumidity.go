package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type Sensor struct {
	ID       string `json:"id"`
	Humidity int    `json:"humidity"`
}

func step(value int) int {
	r := rand.Float64()
	if r > 0.5 {
		value += 1
	} else {
		value -= 1
	}

	if value > 50 {
		value = 50
	}
	if value < 40 {
		value = 40
	}
	return value
}

func main() {
	rand.Seed(time.Now().UnixNano())

	id := fmt.Sprintf("Humidity (%d)", time.Now().Unix())
	humi := 40 + rand.Intn(11) // 40–50

	fmt.Printf("\nSensor de umidade %s inicializado.\n", id)

	for {
		conn, err := net.Dial("udp", "127.0.0.1:7000")
		if err != nil {
			fmt.Println("Erro ao conectar o sensor de humidade: ", id, err)
			continue
		}

		for {
			humi = step(humi)

			data := Sensor{
				ID:       id,
				Humidity: humi,
			}

			values, _ := json.Marshal(data)

			_, err := conn.Write(values)
			if err != nil {
				fmt.Println("Erro no envio do sensor de umidade: ", id, err)
				conn.Close()
				break
			}

			time.Sleep(1 * time.Millisecond)
		}
	}
}
