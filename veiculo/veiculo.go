package main

import (
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

func main() {
	leArquivoJson("dadosVeiculos.json")
	polygon := leMapaFeira()

	// Lendo a variável de ambiente do docker compose
	veiculoID := os.Getenv("PLACA")
	if veiculoID == "" {
		fmt.Println("Erro: PLACA não definida")
		return
	}

	//Faz conexão
	//conn -> representa nossa conexão/rede
	conn, err := net.Dial("tcp", "server:8080")
	//conn, err := net.Dial("tcp", "10.65.133.231:8080")

	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}

	//Envia mensagem para servidor
	mensagem := "VEICULO CONECTADO\n " //tem que terminar com \n se não o servidor não processa
	fmt.Printf("Registro de Veiculo %s conectado ao servidor.\n", veiculoID)
	_, error := conn.Write([]byte(mensagem))
	if error != nil {
		fmt.Println("Erro ao enviar mensagem de registro ao servidor:", err)
		return
	}

	for {
		for i, veiculo := range dados.Veiculos { //Itera entre todos dados para pegar os dados desse veiculo especifico
			if veiculo.Placa == veiculoID {
				fmt.Printf("\nO nível atual de bateria do veículo %s é: %d.\n", veiculoID, veiculo.NivelBateria)
				fmt.Println("\n======================================================================================")

				if veiculo.NivelBateria <= 20 {

					randomCoord := randomPointInBoundingBox(polygon)
					//define mensagem
					mensagem := fmt.Sprintf("VEICULO | Placa %s | Bateria: %d%% | Latitude: %f | Longitude: %f \n",
						veiculo.Placa, veiculo.NivelBateria, randomCoord.Latitude, randomCoord.Longitude)
					fmt.Println("Mensagem Encaminhada ao Servidor:")
					fmt.Println(mensagem)
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

							time.Sleep(20 * time.Second) //espera alguns segundos antes de atualizar
							dados.Veiculos[i].NivelBateria = 100   //atualiza nivel de bateria
							//salvarDadosVeiculos(dados)
							break                        //sai do for de receber mensagem
						}

					}

				} else {
					fmt.Printf("\nA bateria de veículo não está crítica.\n")
					time.Sleep(1 * time.Minute) //espera alguns segundos antes de verificar bateria novamente
					//break
				}

			}
		}
	}

}
