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
	ID          string `json:"id"`
	Temperature int    `json:"temperature"`
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
	if value < -10 {
		value = -10
	}
	return value
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nDigite o ID do sensor de temperatura: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	temp := rand.Intn(61) - 10

	fmt.Print("\033[H\033[2J")
	fmt.Printf("\nSensor de temperatura %s inicializado.\n", id)

	for {
		conn, err := net.Dial("udp", "127.0.0.1:7000")
		if err != nil {
			fmt.Println("Erro ao conectar o sensor de temperatura: ", id, err)
			continue
		}

		for {
			temp = step(temp)

			data := Sensor{
				ID:          id,
				Temperature: temp,
			}

			values, _ := json.Marshal(data)

			_, err := conn.Write(values)
			if err != nil {
				fmt.Println("Erro no envio do sensor de temperatura: ", id, err)
				conn.Close()
				break
			}

			time.Sleep(1 * time.Millisecond)
		}
	}
}
