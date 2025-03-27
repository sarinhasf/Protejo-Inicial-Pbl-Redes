package main

import (
	"encoding/json"
	"fmt"

	//"io/ioutil"
	"net"
	"os"
	"strings"
)

// Estrutura para armazenar dados
type Data struct {
	Veiculos        []string `json:"veiculos"`
	PontosDeRecarga []string `json:"pontos_de_recarga"`
}

func carregarDados() Data {
	file, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Println("Erro ao ler JSON:", err)
		return Data{}
	}
	var data Data
	json.Unmarshal(file, &data)
	return data
}

func salvarDados(data Data) {
	file, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile("data.json", file, 0644)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	//criando buffer para receber dados/mensagens da nossa conexão
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Erro ao ler mensagem:", err)
		return
	}

	//Caso não tenha dado erro, exibimos a mensagem
	mensagem := strings.TrimSpace(string(buffer[:n]))
	fmt.Println("Recebido:", mensagem)

	//define mensagem
	/*mensagem = fmt.Sprintln("oi helena")
	_, err = conn.Write([]byte(mensagem))
	if err != nil {
		fmt.Println("Erro ao enviar mensagem:", err)
		return
	}*/

	//dados := carregarDados()
	//if strings.HasPrefix(mensagem, "VEICULO:") {
	//	id := strings.Split(mensagem, ":")[1]
	//	dados.Veiculos = append(dados.Veiculos, id)
	//	fmt.Println("Veículo registrado:", id)
	//} else if strings.HasPrefix(mensagem, "PONTO:") {
	//	id := strings.Split(mensagem, ":")[1]
	//	dados.PontosDeRecarga = append(dados.PontosDeRecarga, id)
	//	fmt.Println("Ponto de recarga registrado:", id)
	//}

	//salvarDados(dados)
}

func main() {
	//Verificação se o servidor iniciou corretamente
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor rodando na porta 8080...")

	for {
		//conn -> conexão TCP
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}

		//O uso de go antes da chamada cria uma goroutine, ou seja, executa a função de forma assíncrona
		//Ou seja, criamos uma thread
		//Isso permite que o servidor continue aceitando novas conexões sem precisar esperar o
		//processamento de uma conexão terminar
		go handleConnection(conn) //passa a conxeão para nossa função
	}
}
