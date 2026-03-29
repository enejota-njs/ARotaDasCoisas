package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}
type Actuator struct {
	ID   string `json:"id"`
	On   bool   `json:"on"`
	Type string `json:"type"`
}

var light Actuator

func listenServer(conn net.Conn) {
	decoder := json.NewDecoder(conn)
	request := Request{}

	for {
		if err := decoder.Decode(&request); err != nil {
			fmt.Println("\nErro na requisição do cliente: ", err)
			return
		}

		if request.Action == "on" {
			light.On = true
		}
		if request.Action == "off" {
			light.On = false
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Erro ao conectar no servidor: ", err)
		return
	}
	defer conn.Close()

	fmt.Println("\nConectado ao servidor.")

	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("Light (%d)", time.Now().Unix())

	actuator := Actuator{
		ID:   id,
		Type: "light",
	}

	if err := json.NewEncoder(conn).Encode(actuator); err != nil {
		fmt.Println("\nErro ao cadastrar atuador: ", err)
		return
	}

	light.ID = id
	light.Type = "light"

	go listenServer(conn)

	for {
		fmt.Print("\033[H\033[2J")

		on := "Desligado"
		if light.On {
			on = "Ligado"
		}

		fmt.Printf("%s = %s", id, on)
	}
}
