package main

import (
	//"bufio"
	"encoding/csv"
	"fmt"

	//"math"
	"net"
	"os"
	"strconv"
	"strings"

	//"time"
	"encoding/json" //pacote para manipulação de JSON
)

// struct para armazenar os pontos de recarga
type ChargePoint struct {
	Latitude  float64
	Longitude float64
	Nome      string
}

// struct para armazenar pagamentos
type Pagamentos struct {
	IdPonto string
	Valor   float64
}

// struct para armazenar contas de usuario
type ContaUser struct {
	Id         string
	Pagamentos []Pagamentos
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Veiculo struct {
	Placa       string   `json:"placa"`
	Location    Location `json:"location"`
	BateryLevel int      `json:"batery_level"`
	IdConta     string   `json:"conta_id"`
}

// struct para armazenar Dados das contas
type DadosContas struct {
	Contas []ContaUser `json:"contas"`
}

type DadosVeiculos struct {
	Veiculos []Veiculo `json:"veiculos"`
}

var dadosContas DadosContas
var dadosVeiculos DadosVeiculos

func leArquivoJsonContas() {
	bytes, err := os.ReadFile("contasUsuarios.json")
	if err != nil {
		fmt.Println("Erro ao abrir arquivo JSON:", err)
		return
	}

	err = json.Unmarshal(bytes, &dadosContas)
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

func readChargingPoints(filename string) ([]ChargePoint, error) {
	file, err := os.Open(filename) // abre o arquivo
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file) // cria um leitor de csv
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var chargePoints []ChargePoint

	// Itera sobre as linhas do arquivo CSV e extrai os pontos de recarga
	// A primeira linha é o cabeçalho, então começamos a partir da segunda linha
	for i := 1; i < len(records); i++ {
		rawData := records[i][0] //item 0 da linha i
		if strings.HasPrefix(rawData, "POINT") {
			rawData = strings.TrimPrefix(rawData, "POINT (")
			rawData = strings.TrimSuffix(rawData, ")")
			parts := strings.Split(rawData, " ")
			if len(parts) != 2 {
				continue
			}

			lat, _ := strconv.ParseFloat(parts[1], 64)
			lon, _ := strconv.ParseFloat(parts[0], 64)
			nome := records[i][1]

			chargePoints = append(chargePoints, ChargePoint{Latitude: lat, Longitude: lon, Nome: nome})
		}
	}
	return chargePoints, nil
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

func getContaUsuario(id string) (ContaUser, bool) {
	var contaFinal ContaUser
	controle := false

	for _, conta := range dadosContas.Contas {
		if conta.Id == id {
			contaFinal = conta
			controle = true
		}
	}
	return contaFinal, controle
}

func salvarDadosPontos() {
	bytes, err := json.MarshalIndent(dadosContas, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter dadosContas para JSON:", err)
	}

	err = os.WriteFile("contasUsuarios.json", bytes, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar no arquivo contasUsuarios.json:", err)
	}
}

func salvarDadosVeiculos() {
	bytesVeiculos, err := json.MarshalIndent(dadosVeiculos, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter dadosVeiculos para JSON:", err)
	}

	err = os.WriteFile("dadosVeiculos.json", bytesVeiculos, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar no arquivo dadosVeiculos.json:", err)
	}
}

// passa o Id do veiculo, o Id do ponto e o valor
func efetivarPagamento(idVeiculo string, idPonto string, valor float64) {
	veiculo, achou := getVeiculo(idVeiculo)
	contaId := veiculo.IdConta

	if achou {
		contaVeiculo, achou2 := getContaUsuario(contaId)

		if achou2 {
			novoPagamento := Pagamentos{
				IdPonto: idPonto,
				Valor:   valor,
			}
			contaVeiculo.Pagamentos = append(contaVeiculo.Pagamentos, novoPagamento)

			// Salva no arquivo contasUsuarios.json
			salvarDadosPontos()
			fmt.Println("Pagamento registrado com sucesso.")
		} else {
			fmt.Printf("Conta do Veiculo %s não encontrada!", veiculo.Placa)
		}

	} else {
		fmt.Println("Veiculo não encontrado!")
	}
}

func main() {
	leArquivoJsonContas()
	leArquivoJsonVeiculos()

	//Faz conexão
	//conn -> representa nossa conexão/rede
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	//defer conn.Close()

	//Envia mensagem
	mensagem := "PONTO DE RECARGA CONECTADO\n " //tem que terminar com \n se não o servidor não processa
	fmt.Println("Registro de Ponto de recarga conectado ao servidor:", mensagem)

	_, error := conn.Write([]byte(mensagem))
	if error != nil {
		fmt.Println("Erro ao enviar mensagem de registro ao servidor:", err)
		return
	}

}
