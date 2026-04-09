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

// Estrutura de Resposta do servidor
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// Estrutura de Requisição do servidor
type Request struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

// Estrutura do Atuador (Sprinkler)
type Actuator struct {
	Conn net.Conn `json:"-"`
	ID   string   `json:"id"`
	Type string   `json:"type"`
	On   bool     `json:"on"`
}

// Função para limpar o terminal dependendo do sistema operacional
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

// Função para ler e validar o ID do atuador inserido pelo usuário
func readId(reader *bufio.Reader) string {
	for {
		clearTerminal()
		fmt.Print("\nDigite o ID do sprinkler: ")
		idStr, _ := reader.ReadString('\n')
		idStr = strings.TrimSpace(idStr)

		// Verifica se o ID digitado contém apenas números
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

/*
	Função principal:

- Lê o ID do atuador
- Recebe o IP do servidor via argumento e conecta (realiza o cadastro)
- Fica aguardando requisições (Request) do servidor para alterar seu estado
*/
func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	serverIP := os.Args[1] // Recebe IP do servidor

	reader := bufio.NewReader(os.Stdin)
	id := readId(reader)

	var actuator Actuator
	var conn net.Conn
	var err error

	for {
		conn, err = net.Dial("tcp", serverIP+":9000") // Cria conexão com servidor
		if err != nil {
			fmt.Println("\nErro ao conectar no servidor: ", err)
			time.Sleep(1 * time.Second)
			continue
		}

		actuator = Actuator{
			ID:   id,
			Type: "Sprinkler",
		} // Atuador criado

		// Envia os dados do atuador para o servidor (Cadastro)
		if err = json.NewEncoder(conn).Encode(actuator); err != nil {
			fmt.Println("\nErro ao cadastrar atuador: ", err)
			conn.Close()
			continue
		}

		// Aguarda a resposta do servidor sobre o cadastro
		var response Response
		if err = json.NewDecoder(conn).Decode(&response); err != nil {
			fmt.Println("\nErro na resposta do servidor: ", err)
			conn.Close()
			continue
		}

		// Trata erro caso o servidor recuse o cadastro
		if response.Status == "error" {
			fmt.Println("\n", response.Error)
			fmt.Println("\nPressione ENTER para tentar novamente")
			reader.ReadString('\n')
			id = readId(reader)
			conn.Close()
			continue
		}

		// Verifica se o cadastro foi concluído com sucesso
		if response.Status == "success" {
			break
		}
	}

	var on string
	clearTerminal()
	fmt.Println("\nConectado ao servidor")

	// Define a string de exibição baseada no estado inicial do atuador
	if !actuator.On {
		on = "Desligado"
	}
	if actuator.On {
		on = "Ligado"
	}

	fmt.Printf("\n- %s (%s) = %s", actuator.Type, actuator.ID, on)

	decoder := json.NewDecoder(conn)
	request := Request{}

	// Loop infinito: Fica verificando requisições e alterando o estado do atuador
	for {
		if err = decoder.Decode(&request); err != nil {
			clearTerminal()
			fmt.Println("\nDesconectado ao servidor")
			return
		}

		// Muda estado do atuador baseado na requisição recebida
		if request.Action == "on" {
			actuator.On = true
		}
		if request.Action == "off" {
			actuator.On = false
		}

		clearTerminal()
		fmt.Println("\nConectado ao servidor")

		// Atualiza a string de exibição para o novo estado
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
