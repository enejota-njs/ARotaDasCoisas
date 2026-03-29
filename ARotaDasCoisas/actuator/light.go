package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
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
			fmt.Println("\nErro na requisição do servidor: ", err)
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
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nDigite o ID da lâmpada: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Erro ao conectar no servidor: ", err)
		return
	}
	defer conn.Close()

	fmt.Println("\nConectado ao servidor.")

	actuator := Actuator{
		ID:   id,
		Type: "Light",
	}

	if err := json.NewEncoder(conn).Encode(actuator); err != nil {
		fmt.Println("\nErro ao cadastrar atuador: ", err)
		return
	}

	light.ID = id
	light.Type = "Light"

	go listenServer(conn)

	var on string
	for {
		fmt.Print("\033[H\033[2J")

		if !light.On {
			on = "Desligado"
		}
		if light.On {
			on = "Ligado"
		}

		fmt.Printf("%s (%s) = %s", light.Type, light.ID, on)
		time.Sleep(1 * time.Second)
	}
}
