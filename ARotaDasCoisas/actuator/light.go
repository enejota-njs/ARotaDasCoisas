package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}
type Actuator struct {
	Conn net.Conn `json:conn`
	ID   string   `json:"id"`
	Type string   `json:"type"`
	On   bool     `json:"on"`
}

func clearTerminal() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func listenServer(actuator *Actuator, conn net.Conn) {
	decoder := json.NewDecoder(conn)
	request := Request{}

	for {
		if err := decoder.Decode(&request); err != nil {
			fmt.Println("\nErro na requisição do servidor: ", err)
			return
		}

		if request.Action == "on" {
			actuator.On = true
		}
		if request.Action == "off" {
			actuator.On = false
		}
	}
}

func main() {
	clearTerminal()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nDigite o ID da lâmpada: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	conn, err := net.Dial("tcp", "127.0.0.1:9000")
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

	go listenServer(&actuator, conn)

	var on string
	for {
		clearTerminal()

		if !actuator.On {
			on = "Desligado"
		}
		if actuator.On {
			on = "Ligado"
		}

		fmt.Printf("\n%s (%s) = %s", actuator.Type, actuator.ID, on)
		time.Sleep(1 * time.Second)
	}
}
