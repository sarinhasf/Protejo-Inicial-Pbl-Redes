package main

import (
	"fmt"
	"net"
	"os"
	"encoding/json"
	"strings"
	"time"
	"sync"
)

var (
	dadosVeiculos DadosVeiculos
	dadosPontos   DadosPontos
	pontosConns   []net.Conn // Lista de conexões dos pontos de recarga
	mutex sync.Mutex //evitar concorrencia nos arquivos
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Veiculo struct {
	Placa       string   `json:"placa"`
	Location    Location `json:"location"`
	BateryLevel int      `json:"nivel_bateria"`
	IdConta		string   `json:"conta_id"`
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

func getVeiculo(placa string) (Veiculo, bool) {
	var veiculoFinal Veiculo
	controle := false

	for _, veiculo := range dadosVeiculos.Veiculos {
		if veiculo.Placa == placa {
			veiculoFinal = veiculo
			controle = true
		}
	}
	return veiculoFinal, controle
}

func getPonto(nome string) (PontoRecarga, bool) {
	var pontoFinal PontoRecarga
	controle := false

	for _, ponto := range dadosPontos.Pontos {
		if ponto.Nome == nome {
			pontoFinal = ponto
			controle = true
		}
	}
	return pontoFinal, controle
}

func leArquivoJsonPontos() {
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

func leArquivoJsonVeiculos() {
	bytes, err := os.ReadFile("dadosVeiculos.json")
	if err != nil {
		fmt.Println("Erro ao abrir arquivo JSON:", err)
		return
	}

	err = json.Unmarshal(bytes, &dadosVeiculos)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}
}

func addFila(idPonto string, placaVeiculo string) {
	mutex.Lock()         //bloqueia acesso concorrente
	defer mutex.Unlock() //libera depois da execução

	encontrado := false

	for i, ponto := range dadosPontos.Pontos {
		if strings.TrimSpace(ponto.Id) == strings.TrimSpace(idPonto) {
			dadosPontos.Pontos[i].Fila = append(ponto.Fila, placaVeiculo)
			encontrado = true
			break
		}
	}

	if encontrado {
		salvarDadosPontos()
	} else {
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

func salvarDadosPontos() {
	bytes, err := json.MarshalIndent(dadosPontos, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter dadosContas para JSON:", err)
	}

	err = os.WriteFile("dadosPontos.json", bytes, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar no arquivo dadosPontos.json:", err)
	}
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

				var pontoID string = closestPoint.Id
				var filaPonto []string //pega a fila do ponto +prox
				var nomePontoProx string
				for _, ponto := range dadosPontos.Pontos {
					if strings.EqualFold(ponto.Id, pontoID) {
						filaPonto = ponto.Fila
						nomePontoProx = ponto.Nome
					}
				}

				var melhorPontoId string
				var melhorPontoNome string

				//caso a fila do ponto +prox esteja vazia, envia esse ponto
				if len(filaPonto) == 0 {
					melhorPontoNome = nomePontoProx
					melhorPontoId = pontoID

					fmt.Printf("Ponto de recarga mais próximo do veículo %s: %s - %.2fKm \n", placa, nomePontoProx, distance)

					mensagem := fmt.Sprintf("Ponto de recarga mais próximo: %s - Distância: %.2fKm\n", nomePontoProx, distance)
					_, err := conn.Write([]byte(mensagem)) //envia mensagem
					if err != nil {
						fmt.Println("Erro ao enviar mensagem:", err)
						return
					}

				} else { //se não faz a analise com todos os pontos
					veiculo, achou := getVeiculo(placa) //pega veiculo

					if achou {

						mensagem, melhorPonto := analiseTodosPontos(lat, lon, veiculo.BateryLevel, placa) //eniva a melhor escolha
						melhorPontoNome = melhorPonto.Ponto.Nome
						melhorPontoId = melhorPonto.Ponto.Id

						_, err := conn.Write([]byte(mensagem)) //envia mensagem
						if err != nil {
							fmt.Println("Erro ao enviar mensagem:", err)
							return
						}

					} else {
						fmt.Print("Não foi possível encontrar veiculo com essa placa.\n\n")
					}
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
				if strings.ToLower(resposta) == "sim" {
					// Adiciona o veículo à fila do ponto de recarga
					addFila(melhorPontoId, placa)

					confirmacao := fmt.Sprintf("Veículo %s adicionado à fila do ponto de recarga %s\n", placa, melhorPontoNome)
					fmt.Println(confirmacao)

					// Envia a confirmação para o veículo
					_, err := conn.Write([]byte(confirmacao))
					if err != nil {
						fmt.Println("Erro ao enviar confirmação para o veículo:", err)
						return
					}

				} else {
					fmt.Printf("Infelizmente o veiculo não aceitou entrar na fila do ponto %s", melhorPontoNome)
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
	leArquivoJsonVeiculos() //lendo os arquivos dos veiculos
	leArquivoJsonPontos()   //lendo os arquivos dos pontos

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
