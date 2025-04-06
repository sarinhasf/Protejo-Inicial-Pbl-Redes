package main

import (
	"encoding/json" //pacote para manipulação de JSON
	"fmt"           //pacote para formatação de strings
	"net"           //pacote para comunicação em rede
	"os"            //pacote para manipulação de arquivos
	"strings"
	"time"
)

// Definindo estrutura com os dados dos veiculos
var dados Dados
var veiculoID string
var polygon []Point

// struct para armazenar coordenadas
type Point struct {
	Latitude  float64
	Longitude float64
}

// Criando structs apartir do JSON
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
	Veiculos []Veiculo `json:"veiculos"`
}

// Lê o arquivo JSON e armazena os dados na variável global "dados"
func leArquivoJson(filename string) {
	// Verifica se o arquivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("Arquivo JSON não encontrado:", filename)
		return
	}
	// Ler o arquivo JSON usando os.ReadFile
	bytes, err := os.ReadFile(filename)
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
	leArquivoJson("dadosVeiculos.json")

	polygon, err := readPolygon("MapaDeFeira.csv")
	if err != nil {
		fmt.Println("Erro ao ler arquivo CSV:", err)
		return
	}
	if len(polygon) == 0 {
		fmt.Println("Nenhum ponto encontrado no arquivo CSV")
		return
	}

	// Lendo a variável de ambiente do docker compose
	veiculoID := os.Getenv("PLACA")
	if veiculoID == "" {
		fmt.Println("Erro: PLACA não definida")
		return
	}

	//Faz conexão
	//conn -> representa nossa conexão/rede
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}

	for { //criando for infinito para manter conexão
		for _, veiculo := range dados.Veiculos {
			fmt.Println("\n==========================================================================")
			if veiculo.Placa == veiculoID && veiculo.NivelBateria <= 20 {
				//randomCoord := randomPointInBoundingBox(polygon)
				//define mensagem
				//mensagem := fmt.Sprintf("VEICULO %s | Bateria: %d%% | Latitude: %f | Longitude: %f \n",
				//	veiculo.Placa, veiculo.NivelBateria, randomCoord.Latitude, randomCoord.Longitude)
				mensagem := fmt.Sprintf("VEICULO %s | Bateria: %d%% | Latitude: %f | Longitude: %f \n",
					veiculo.Placa, veiculo.NivelBateria, -12.260784, -38.980637)
				//fmt.Println(mensagem)
				fmt.Println("Veículo enviado ao servidor:", mensagem)
				time.Sleep(5 * time.Second) //espera alguns segundos antes de enviar de fato a mensagem

				_, err := conn.Write([]byte(mensagem)) //envia mensagem
				if err != nil {
					fmt.Println("Erro ao enviar mensagem:", err)
					return
				}

				// lê resposta do servidor
				buffer := make([]byte, 1024) //cria buffer para receber dados
				n, err := conn.Read(buffer)
				if err != nil {
					fmt.Println("Erro ao receber mensagem do servidor:", err)
					return
				}
				mensagem2 := string(buffer[:n])
				fmt.Println(mensagem2) //exibe mensagem recebida

				if strings.Contains(mensagem2, "Ponto de recarga mais próximo:") {
					fmt.Println("Deseja entrar na fila(S/N)?")
					var reserva string
					fmt.Scanln(&reserva) //lê resposta do usuário

					_, err := conn.Write([]byte(reserva + "\n")) //envia resposta
					if err != nil {
						fmt.Println("Erro ao enviar resposta ao servidor:", err)
						return
					}
					//fmt.Println("Resposta enviada ao servidor:", reserva)

					// lê resposta do servidor
					buffer := make([]byte, 1024) //cria buffer para receber dados
					n, err := conn.Read(buffer)
					if err != nil {
						fmt.Println("Erro ao receber mensagem do servidor:", err)
						return
					}
					fmt.Println(string(buffer[:n])) //exibe mensagem recebida
				}
			}
			time.Sleep(1 * time.Minute)
		}
	}
}
