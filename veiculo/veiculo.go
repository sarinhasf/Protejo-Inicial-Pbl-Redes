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
	IdConta      int         `json:"conta_id"`
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

	//Envia mensagem
	mensagem := "VEICULO CONECTADO\n " //tem que terminar com \n se não o servidor não processa
	fmt.Printf("Registro de Veiculo %s conectado ao servidor.\n", veiculoID)
	_, error := conn.Write([]byte(mensagem))
	if error != nil {
		fmt.Println("Erro ao enviar mensagem de registro ao servidor:", err)
		return
	}
	fmt.Println("\n======================================================================================")

	for _, veiculo := range dados.Veiculos {
		if veiculo.Placa == veiculoID {
			fmt.Printf("\nO nível atual de bateria do veículo %s é: %d.\n", veiculoID, veiculo.NivelBateria)

			if veiculo.NivelBateria <= 20 {
				randomCoord := randomPointInBoundingBox(polygon)
				//define mensagem
				mensagem := fmt.Sprintf("VEICULO | Placa %s | Bateria: %d%% | Latitude: %f | Longitude: %f \n",
					veiculo.Placa, veiculo.NivelBateria, randomCoord.Latitude, randomCoord.Longitude)
				fmt.Println("Mensagem Encaminhada ao Servidor:")
				fmt.Println(mensagem)
				//fmt.Println("Veículo enviado ao servidor:", mensagem)
				time.Sleep(5 * time.Second) //espera alguns segundos antes de enviar de fato a mensagem

				_, err := conn.Write([]byte(mensagem)) //envia mensagem
				if err != nil {
					fmt.Println("Erro ao enviar mensagem:", err)
					return
				}

				for { //cria loop para pegar as informações
					buffer := make([]byte, 1024) // cria buffer para receber dados
					n, err := conn.Read(buffer)
					if err != nil {
						if err.Error() == "EOF" { // Conexão encerrada pelo servidor
							fmt.Println("Conexão encerrada pelo servidor.")
							break
						}
						fmt.Println("Erro ao receber mensagem do servidor:", err)
						continue
					}

					mensagemRecebida := string(buffer[:n])
					fmt.Println("\nMensagem recebida do servidor sobre o melhor ponto:")
					fmt.Println(mensagemRecebida) //exibe mensagem recebida

					if strings.Contains(mensagemRecebida, "Melhor ponto para o veículo") {
						fmt.Println("Deseja entrar na fila(S/N)?")
						var reserva string
						fmt.Scanln(&reserva) //lê resposta do usuário
						//reserva := "sim"
						fmt.Printf("O veículo respondeu que %s.", reserva)

						_, err := conn.Write([]byte("VEICULO " + reserva + "\n")) //envia resposta
						if err != nil {
							fmt.Println("Erro ao enviar resposta ao servidor:", err)
							return
						} else {
							fmt.Printf("\nResposta Enviada ao Servidor: [%s].\n", reserva) //exibe mensagem recebida
						}

						// lê resposta do servidor
						buffer := make([]byte, 1024) //cria buffer para receber dados
						n, err := conn.Read(buffer)
						if err != nil {
							fmt.Println("Erro ao receber mensagem do servidor:", err)
							return
						}
						fmt.Println(string(buffer[:n])) //exibe mensagem recebida

					} else if strings.Contains(mensagemRecebida, "PONTO: Veiculo") {
						mensagemRecebida = strings.TrimPrefix(mensagemRecebida, "PONTO: ")
						fmt.Println(mensagemRecebida)
						break
					}
				}

			} else {
				fmt.Printf("\nA bateria de veículo não está crítica.\n")
				break
			}
		}
	}

}
