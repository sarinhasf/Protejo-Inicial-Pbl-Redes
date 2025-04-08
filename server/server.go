package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

type ConnInfo struct {
	Conn          net.Conn
	Tipo          ConnTipo
	Identificador string // placa ou ID do ponto
	Estado        string // "aguardando_resposta", "aguardando_confirmacao", etc.
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Veiculo struct {
	Placa       string   `json:"placa"`
	Location    Location `json:"location"`
	BateryLevel int      `json:"nivel_bateria"`
	IdConta     int   `json:"conta_id"`
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

type TipoCliente string
type ConnTipo string

type SessaoCliente struct {
    Conn               net.Conn       // Representa a conexão de rede do cliente, do tipo net.Conn.
    Tipo               TipoCliente    // Define o tipo do cliente (veículo, ponto de recarga), do tipo TipoCliente (string).
    PlacaVeiculo       string         // Armazena a placa do veículo associada ao cliente, do tipo string.
    AguardandoResposta bool           // Indica se o cliente está aguardando uma resposta, do tipo booleano.
    MelhorPontoID      string         // Identificador do melhor ponto de recarga sugerido, do tipo string.
    MelhorPontoNome    string         // Nome do melhor ponto de recarga sugerido, do tipo string.
}

var (
	dadosVeiculos DadosVeiculos
	dadosPontos   DadosPontos
	pontosConns 	map[string]net.Conn // Lista de conexões dos pontos de recarga
	veiculosConns = map[string]net.Conn{}
	mutex         sync.Mutex					//evitar concorrencia nos arquivos
)

const (
	TipoVeiculo ConnTipo = "VEICULO"
	TipoPonto   ConnTipo = "PONTO"
)


func main() {
	leArquivoJsonVeiculos() //lendo os arquivos dos veiculos
	leArquivoJsonPontos()   //lendo os arquivos dos pontos

	pontosConns = make(map[string]net.Conn) //inicializa os pontosConns

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