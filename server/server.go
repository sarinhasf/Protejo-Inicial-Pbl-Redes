package main

import (
	"fmt"
	"net"
	"os"

	//"strconv"
	"encoding/json" //pacote para manipulação de JSON
	"strings"
	"time"
)

var (
	dadosVeiculos DadosVeiculos
	dadosPontos   DadosPontos
	pontosConns   []net.Conn // Lista de conexões dos pontos de recarga
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Veiculo struct {
	Id          string   `json:"id"`
	Placa       string   `json:"placa"`
	Location    Location `json:"location"`
	BateryLevel int      `json:"batery_level"`
}

type PontoRecarga struct {
	Id         string
	Nome       string
	Fila       []string
	Carregando string
}

type DadosVeiculos struct {
	Veiculos []Veiculo `json:"veiculos"`
}

type DadosPontos struct {
	Pontos []PontoRecarga `json:"pontos"`
}

/*func getVeiculo(id string) (Veiculo, bool){
	var veiculoFinal Veiculo
	controle := false

	for _, veiculo := range dados.Veiculos {
		if veiculo.Id == id {
			veiculoFinal = veiculo
			controle = true
		}
	}
	return veiculoFinal, controle
}*/

func leArquivoJson() {
	bytes, err := os.ReadFile("dadosPontos.json")
	if err != nil {
		fmt.Println("Erro ao abrir arquivo JSON:", err)
		return
	}

	err = json.Unmarshal(bytes, &dadosPontos)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}
}

func addFila(idPonto string, placaVeiculo string) {
	encontrado := false

	for i, ponto := range dadosPontos.Pontos {
		if strings.TrimSpace(ponto.Id) == strings.TrimSpace(idPonto) {
			dadosPontos.Pontos[i].Fila = append(ponto.Fila, placaVeiculo)
			fmt.Printf("Veículo %s adicionado à fila do ponto %s\n", placaVeiculo, ponto.Nome)
			encontrado = true
			break
		}
	}

	if !encontrado {
		fmt.Printf("Erro: Ponto de recarga com ID %s não encontrado\n", idPonto)
	}
}

func removeFila(idPonto string, idCarro string) {
	for i, ponto := range dadosPontos.Pontos {
		if ponto.Id == idPonto {
			for j, carro := range ponto.Fila {
				if carro == idCarro {
					// Remove o carro da fila
					dadosPontos.Pontos[i].Fila = append(ponto.Fila[:j], ponto.Fila[j+1:]...)
					return
				}
			}
		}
	}
}

func salvarDados(data DadosPontos) {
	file, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile("dadosPontos.json", file, 0644)
}

func handleConnection(conn net.Conn) {
	bufferAcumulado := "" // buffer para armazenar dados recebidos

	for { // loop infinito para receber mensagens continuamente

		//criando buffer para receber dados/mensagens da nossa conexão
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer) //n -> número de bytes lidos
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Conexão encerrada pelo cliente.")
				return
			}
			// Se ocorrer um erro diferente de EOF, exibe a mensagem de erro
			fmt.Println("Erro ao ler mensagem:", err)
			return
		}
		// adiciona os dados recebidos ao buffer acumulado
		bufferAcumulado += string(buffer[:n]) //pega apenas os bytes válidos, evitando partes vazias

		//cria uma lista de mensagens separadas pelo \n
		mensagens := strings.Split(bufferAcumulado, "\n")
		// processa todas as mensagens completas
		for i := 0; i < len(mensagens)-1; i++ {
			mensagem := strings.TrimSpace(mensagens[i])
			if mensagem == "" {
				continue
			}

			// processa a mensagem recebida e envia confirmacao
			// feita para receber veiculo, ponto de recarga nao reconhce
			if strings.HasPrefix(mensagem, "VEICULO") {
				placa, lat, lon := trataInfo(mensagem)

				//calcula o ponto de recarga mais próximo do veículo
				closestPoint, distance := pegaPontoProximo(lat, lon)
				fmt.Printf("Ponto de recarga mais próximo do veículo %s: ID %s - Distância %.2fKm \n", placa, closestPoint.Id, distance)

				//envia a ponto mais próximo para o veículo
				mensagem := fmt.Sprintf("Ponto de recarga mais próximo: ID %s - Distância: %.2fKm\n", closestPoint.Id, distance)
				_, err := conn.Write([]byte(mensagem)) //envia mensagem
				if err != nil {
					fmt.Println("Erro ao enviar mensagem:", err)
					return
				}

				// Lê a resposta do veículo
				buffer2 := make([]byte, 1024)
				n, err := conn.Read(buffer2)
				if err != nil {
					fmt.Println("Erro ao receber resposta do veículo:", err)
					return
				}
				resposta := strings.TrimSpace(string(buffer2[:n]))
				fmt.Printf("Resposta do veículo %s: %s\n", placa, resposta)

				// Verifica se a resposta é "sim"
				if strings.ToLower(resposta) == "sim" { // caso o usuário digite "sim", "Sim" ou "SIM"
					// Adiciona o veículo à fila do ponto de recarga
					addFila(closestPoint.Id, placa)

					confirmacao := fmt.Sprintf("Veículo %s adicionado à fila do ponto de recarga %s\n", placa, closestPoint.Id)
					fmt.Println(confirmacao)

					// Envia a confirmação para o veículo
					_, err := conn.Write([]byte(confirmacao))
					if err != nil {
						fmt.Println("Erro ao enviar confirmação para o veículo:", err)
						return
					}
				}

			} else if strings.HasPrefix(mensagem, "PONTO") { //
				time.Sleep(3 * time.Second)             //espera alguns segundos antes de enviar de fato a mensagem
				pontosConns = append(pontosConns, conn) // lista para armazenar as conexões dos pontos
				fmt.Println("Novo ponto de recarga conectado!")

			} else {
				fmt.Print("Aguardando nova requisição dos veiculos.\n\n")
			}
		}

		bufferAcumulado = mensagens[len(mensagens)-1] // limpa o buffer
		//defer conn.Close()

	}
}

func main() {
	leArquivoJson() //lendo os arquivos do ponto

	//Verificação se o servidor iniciou corretamente
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Print("Servidor rodando na porta 8080...\n\n")

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
