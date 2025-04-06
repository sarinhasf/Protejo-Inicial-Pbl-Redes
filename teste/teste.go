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

func getVeiculo(id string) (Veiculo, bool) {
	var veiculoFinal Veiculo
	controle := false

	for _, veiculo := range dadosVeiculos.Veiculos {
		if veiculo.Placa == id {
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

				var pontoNome string = closestPoint.Nome
				//pega a fila do ponto
				var filaPonto []string
				//fmt.Printf("Nome do ponto +prox é %s\n", pontoNome)
				for _, ponto := range dadosPontos.Pontos {
					if strings.EqualFold(ponto.Nome, pontoNome) {
						filaPonto = ponto.Fila
						//numFila := len(filaPonto)
						//fmt.Printf("Ponto %s, sua fila é %d carro(s)\n", pontoNome, numFila)
					}
				}
				//caso a fila do ponto +prox esteja vazia, envia esse ponto
				if len(filaPonto) == 0 {
					fmt.Printf("Ponto de recarga mais próximo do veículo %s: %s - %.2fKm \n", placa, pontoNome, distance)

				} else { //se não faz a analise com todos os pontos
					veiculo, achou := getVeiculo(placa)
					if achou {
						analiseTodosPontos(lat, lon, veiculo.BateryLevel, placa)
					} else {
						fmt.Print("Não foi possível encontrar veiculo com essa placa.\n\n")
					}
				}

			} else if strings.HasPrefix(mensagem, "PONTO") { //
				time.Sleep(3 * time.Second)             //espera alguns segundos antes de enviar de fato a mensagem
				pontosConns = append(pontosConns, conn) // lista para armazenar as conexões dos pontos
				fmt.Println("Novo ponto de recarga conectado!")

			} else {
				fmt.Print("Aguardando nova requisção dos veiculos.\n\n")
			}
		}

		bufferAcumulado = mensagens[len(mensagens)-1] // limpa o buffer
		//defer conn.Close()

	}
}

func leArquivoJsonPonto() {
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

func addFila(idPonto string, idCarro string) {
	for _, ponto := range dadosPontos.Pontos {
		if ponto.Id == idPonto {
			ponto.Fila = append(ponto.Fila, idPonto)
		}
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

func main() {
	leArquivoJsonPonto()    //lendo os arquivos do ponto
	leArquivoJsonVeiculos() //lendo os dados do veiculo

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
