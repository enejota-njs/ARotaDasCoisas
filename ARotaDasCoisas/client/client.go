package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Erro ao conectar: ", err)
		return
	}
	defer conn.Close()

	fmt.Println("Conectado ao servidor.")

	reader := bufio.NewReader(conn)

	for {
		values, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Conexão encerrada:", err)
			return
		}

		fmt.Println(values)
	}
}
