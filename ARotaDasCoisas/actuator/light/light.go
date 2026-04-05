package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}
type Actuator struct {
	Conn net.Conn `json:"-"`
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

func readId(reader *bufio.Reader) string {
	for {
		clearTerminal()
		fmt.Print("\nDigite o ID da lâmpada: ")
		idStr, _ := reader.ReadString('\n')
		idStr = strings.TrimSpace(idStr)

		_, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Println("\nDigite apenas números")
			reader = bufio.NewReader(os.Stdin)
			fmt.Println("\nPressione ENTER para tentar novamente")
			reader.ReadString('\n')
			continue
		}

		return idStr
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	id := readId(reader)

	var actuator Actuator
	var conn net.Conn
	var err error

	for {
		conn, err = net.Dial("tcp", "server:9000")
		if err != nil {
			fmt.Println("\nErro ao conectar no servidor: ", err)
			time.Sleep(1 * time.Second)
			continue
		}

		actuator = Actuator{
			ID:   id,
			Type: "Lâmpada",
		}

		if err = json.NewEncoder(conn).Encode(actuator); err != nil {
			fmt.Println("\nErro ao cadastrar atuador: ", err)
			conn.Close()
			continue
		}

		var response Response
		if err = json.NewDecoder(conn).Decode(&response); err != nil {
			fmt.Println("\nErro na resposta do servidor: ", err)
			conn.Close()
			continue
		}

		if response.Status == "error" {
			fmt.Println("\n", response.Error)
			fmt.Println("\nPressione ENTER para tentar novamente")
			reader.ReadString('\n')
			id = readId(reader)
			conn.Close()
			continue
		}

		if response.Status == "success" {
			break
		}
	}

	var on string
	clearTerminal()
	fmt.Println("\nConectado ao servidor")

	if !actuator.On {
		on = "Desligado"
	}
	if actuator.On {
		on = "Ligado"
	}

	fmt.Printf("\n- %s (%s) = %s", actuator.Type, actuator.ID, on)

	decoder := json.NewDecoder(conn)
	request := Request{}

	for {
		if err = decoder.Decode(&request); err != nil {
			clearTerminal()
			fmt.Println("\nDesconectado ao servidor")
			return
		}

		if request.Action == "on" {
			actuator.On = true
		}
		if request.Action == "off" {
			actuator.On = false
		}

		clearTerminal()
		fmt.Println("\nConectado ao servidor")

		if !actuator.On {
			on = "Desligado"
		}
		if actuator.On {
			on = "Ligado"
		}

		fmt.Printf("\n- %s (%s) = %s", actuator.Type, actuator.ID, on)
	}

	conn.Close()
}
