package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"encoding/json" //pacote para manipulação de JSON

)


var (
	dadosVeiculos       DadosVeiculos
	dadosPontos         DadosPontos
	pontosConns []net.Conn // Lista de conexões dos pontos de recarga
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
	Id		   string
	Nome       string
	Fila 	   []string
	Carregando string
}

type DadosVeiculos struct {
	Veiculos []Veiculo        `json:"veiculos"`
}

type DadosPontos struct {
	Pontos []PontoRecarga        `json:"pontos"`
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
				// Se a mensagem começa com "VEICULO", processa como um veículo
				// Exemplo de mensagem: "VEICULO Placa1234 | Bateria: % | Latitude: lat | Longitude: long"
				mensagem = strings.TrimPrefix(mensagem, "VEICULO")
				parts := strings.Split(mensagem, "|")
				if len(parts) != 4 {
					fmt.Println("Mensagem inválida:", mensagem)
					return
				}

				placa := strings.TrimSpace(parts[0])
				// Remove o prefixo e sufixo
				// Exemplo: "Bateria: 80%" -> "80"
				bateria := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(parts[1], " Bateria: "), "% "))
				latitude := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(parts[2], " Latitude: "), " "))
				longitude := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(parts[3], " Longitude: "), " "))

				// sanitiza a longitude para evitar problemas de formatação
				if strings.Contains(longitude, "\n ") { // se a longitude contiver "\n ", remove o espaço em branco
					longitude = strings.TrimSpace(strings.TrimSuffix(parts[3], "\n"))
				}

				// Converte os valores para os tipos corretos
				bateriaInt, err := strconv.Atoi(bateria)
				if err != nil {
					fmt.Println("Erro ao converter bateria:", err)
					return
				}
				latitudeFloat, err := strconv.ParseFloat(latitude, 64)
				if err != nil {
					fmt.Println("Erro ao converter latitude:", err)
					return
				}
				longitudeFloat, err := strconv.ParseFloat(longitude, 64)
				if err != nil {
					fmt.Println("Erro ao converter longitude:", err)
					return
				}

				// Cria um novo veículo
				novoVeiculo := Veiculo{ // ****** OLHA AQUI DEPOIS ******
					Id:          placa, // ID do veículo (pode ser o mesmo que a placa) se for tirar o id ou de fato dar um id diferente da placa
					Placa:       placa,
					Location:    Location{Latitude: latitudeFloat, Longitude: longitudeFloat},
					BateryLevel: bateriaInt,
				}

				dadosVeiculos.Veiculos = append(dadosVeiculos.Veiculos, novoVeiculo) // adiciona o novo veículo à lista de veículos
				/*fmt.Println("Veículos armazenados atualmente:")
				for _, veiculo := range dados.Veiculos {
					fmt.Printf(" | %s | %d%% | %.6f | %.6f |\n", veiculo.Placa, veiculo.BateryLevel, veiculo.Location.Latitude, veiculo.Location.Longitude)
				}*/

				novasConns := []net.Conn{}
				mensagemParaPonto := fmt.Sprintf("%.6f,%.6f\n", novoVeiculo.Location.Latitude, novoVeiculo.Location.Longitude)
				for _, pontoConn := range pontosConns {
					_, err = pontoConn.Write([]byte(mensagemParaPonto))
					if err != nil {
						fmt.Println("Erro ao enviar mensagem para o ponto:", err, "FECHANDO CONEXÃO COM O PONTO")
						pontoConn.Close()
						continue
					}
					fmt.Println("Mensagem enviada ao ponto:", mensagemParaPonto)
					novasConns = append(novasConns, pontoConn)
				}
				pontosConns = novasConns

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

func addFila(idPonto string, idCarro string){
	for _, ponto := range dadosPontos.Pontos {
		if(ponto.Id == idPonto){
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
	leArquivoJson(); //lendo os arquivos do ponto

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
