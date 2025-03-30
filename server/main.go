package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

// Definindo estrutura com os dados dos veiculos
var dados Dados

type Localizacao struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Veiculo struct {
	Id           string      `json:"ID"`
	Placa        string      `json:"placa"`
	Localizacao  Localizacao `json:"localizacao"`
	NivelBateria int         `json:"nivel_bateria"`
}

type PontoRecarga struct {
	Nome      string
	Latitude  float64
	Longitude float64
}

type Dados struct {
	Veiculos []Veiculo `json:"veiculos"`
	Pontos   []PontoRecarga 
}



func lerArquivoJson() {

	// Ler o arquivo JSON usando os.ReadFile
	bytes, err := os.ReadFile("dados/dadosVeiculos.json")
	if err != nil {
		fmt.Println("Erro ao abrir arquivo JSON:", err)
		return
	}

	// Passando dados do JSON para struct criada dados
	err = json.Unmarshal(bytes, &dados)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}
}

func getVeiculo(id string) (Veiculo, bool){
	var veiculoFinal Veiculo 	// Estrutura de dados do veiculo
	controle := false

	for _, veiculo := range dados.Veiculos {
		if veiculo.Id == id {
			veiculoFinal = veiculo
			controle = true
		}
	}
	return veiculoFinal, controle
}


func handleConnection(conn net.Conn) { //conn -> conexão
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

	//Verificar se a mensagem é do veiculo/ponto (essa sinalização é feia antes da virgula)
	partes := strings.Split(mensagem, ",") // Divide a string recebida 
	tipo := partes[0] //sinaliza se a mensagem é do veiculo ou do ponto
	id := partes[1]
	requisicao := partes[2]

	// Lendo a variável de ambiente do docker compose para pegar o ID do veiculo
	veiculoID := os.Getenv("ID-VEICULO")
	if veiculoID == "" {
		fmt.Println("Erro: ID-VEICULO não definido")
		return
	}

	if(tipo == "veiculo"){ //se for veiculo
		if(veiculoID == id){

			veiculoEncontrado, controle := getVeiculo(veiculoID)
			if(controle){
				fmt.Println("A placa do Veiculo com bateria baixa é:", veiculoEncontrado)
			} else {
				fmt.Println("Veículo não encontrado com ID:", veiculoID)
			}

			if(requisicao == "bb") { //se a requisição for do tipo bateria baixa
				

			}
		}

	}

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
