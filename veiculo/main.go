package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

//Definindo estrutura com os dados dos veiculos
var dados Dados

//Criando structs apartir do JSON
type Localizacao struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Veiculo struct {
	Placa        string      `json:"placa"`
	Localizacao  Localizacao `json:"localizacao"`
	NivelBateria int         `json:"nivel_bateria"`
}

type Dados struct {
	Veiculos        []Veiculo `json:"veiculos"`
}

func lerArquivoJson(){
	
	// Ler o arquivo JSON usando os.ReadFile 
	bytes, err := os.ReadFile("dados.json")
	if err != nil {
		fmt.Println("Erro ao abrir arquivo JSON:", err)
		return
	}

	// Passando dados do JSON para struct criada
	err = json.Unmarshal(bytes, &dados)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}
}

func main() {

	lerArquivoJson();

	//Faz conexão
	//conn -> representa nossa conexão/rede
	conn, err := net.Dial("tcp", "server:8080")

	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	for _, veiculo := range dados.Veiculos {
		if(veiculo.NivelBateria <= 30){
			//define mensagem
			mensagem := "VEICULO " + veiculo.Placa + " com bível de bateria critico!" 
			conn.Write([]byte(mensagem))
			fmt.Println("Veículo enviado ao servidor:", mensagem) //envia mensagem
		}
	}
}
