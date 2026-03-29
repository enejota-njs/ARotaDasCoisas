package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
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

	if value > 100 {
		value = 100
	}
	if value < 0 {
		value = 0
	}
	return value
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nDigite o ID do sensor luminosidade: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	lumi := rand.Intn(101)

	fmt.Print("\033[H\033[2J")
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
