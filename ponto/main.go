package main

import (
	"fmt"
	"net"
)

func main() {
	//Faz conexão
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	//Envia mensagem
	mensagem := "PONTO:5678"
	conn.Write([]byte(mensagem))

	fmt.Println("Ponto de recarga enviado ao servidor:", mensagem)
}
