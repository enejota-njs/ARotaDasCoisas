package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		fmt.Println("Erro ao conectar no servidor: ", err)
		return
	}
	defer conn.Close()

	fmt.Println("Conectado ao servidor.")

	input := bufio.NewReader(os.Stdin)
	reader := bufio.NewReader(conn)

	for {
		fmt.Println("\n|----------- MENU -----------|")
		fmt.Println("| [ 1 ] - Listar sensores    |")
		fmt.Println("| [ 2 ] - Verificar sensores |")
		fmt.Println("| [ 3 ] - Selecionar sensor  |")
		fmt.Println("|----------------------------|")

		fmt.Print("\nSelecione uma opção: ")

		option, _ := input.ReadString('\n')
		option = strings.TrimSpace(option)

		fmt.Println("")

		switch option {
		case "1":
			fmt.Fprint(conn, "1|\n")
		case "2":
			fmt.Fprint(conn, "2|\n")
			go func() {
				for {
					fmt.Print("\nAperte [ S ] para voltar: ")

					optionS, _ := input.ReadString('\n')
					optionS = strings.TrimSpace(optionS)

					if optionS == "s" || optionS == "S" {
						return
					}
				}
			}()
		case "3":
			fmt.Print("\nDigite o ID do sensor: ")

			id, _ := input.ReadString('\n')

			fmt.Fprint(conn, "3|"+id+"\n")
		default:
			fmt.Println("Opção inválida.")
			continue
		}

		for {
			result, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Conexão encerrada: ", err)
				return
			}

			if strings.TrimSpace(result) == "ok" {
				break
			}

			fmt.Print(result)
		}
	}
}
