package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"encoding/json" //pacote para manipulação de JSON
	"time"          //pacote para manipulação de tempo
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

// Estrutura para o histórico do ponto
type Historico struct {
	Carro  string `json:"carro"`
	Status string `json:"status"`
}

var dadosContas DadosContas
var dadosVeiculos DadosVeiculos
var mutex sync.Mutex //evitar concorrencia nos arquivos

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

func salvarDadosContas() {
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

func processarFila(idPonto string, filaSlice []string) {
	// Verifica se a fila não está vazia
	if len(filaSlice) == 0 {
		fmt.Println("Fila vazia, nada para processar.")
		return
	}

	// Pega o primeiro carro da fila
	carro := filaSlice[0]

	// Cria o histórico com status "carregando"
	historico := Historico{
		Carro:  carro,
		Status: "carregando",
	}

	// Salva o histórico no arquivo JSON
	filename := fmt.Sprintf("historicodoponto%s.json", idPonto)
	salvarHistorico(filename, historico)

	// Aguarda um minuto
	fmt.Printf("Carro %s está carregando...\n", carro)
	time.Sleep(15 * time.Second) // Simula o tempo de carregamento (15 segundos para teste)

	// Atualiza o status para "carregado"
	historico.Status = "carregado"
	salvarHistorico(filename, historico)
	fmt.Printf("Carro %s foi carregado.\n", carro)

	// le e imprime o arquivo JSON de histórico
	conteudo, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Erro ao ler arquivo de histórico:", err)
		return
	}
	fmt.Printf("Conteúdo do arquivo %s: \n%s \n", filename, string(conteudo))

	// Remove o carro da fila
	filaSlice = filaSlice[1:] // Remove o primeiro carro da fila
	fmt.Printf("Fila atualizada do ponto %s: %v\n", idPonto, filaSlice)

}

func salvarHistorico(filename string, historico Historico) {
	bytes, err := json.MarshalIndent(historico, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter histórico para JSON:", err)
		return
	}

	err = os.WriteFile(filename, bytes, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar o arquivo histórico:", err)
	}
}

// passa o Id do veiculo, o Id do ponto e o valor
func efetivarPagamento(idVeiculo string, idPonto string, valor float64) {
	mutex.Lock()         //bloqueia acesso concorrente
	defer mutex.Unlock() //libera depois da execução da função

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
			salvarDadosContas()
			fmt.Println("Pagamento registrado com sucesso.")
		} else {
			fmt.Printf("Conta do Veiculo %s não encontrada!", veiculo.Placa)
		}

	} else {
		fmt.Println("Veiculo não encontrado!")
	}
}

func calculaPrecoRecarga(nivelBateria int) float64 {
	bateria := float64(nivelBateria)                                         // nível da bateria do carro
	var precoPorKWh float64 = 0.5                                            // preço médio por kWh
	var capacidadeBateria float64 = 50.0                                     // capacidade média de um carro elétrico
	var KWPorPorcentagem float64 = capacidadeBateria / 100.0                 // kw por %
	var bateriaEmKW float64 = KWPorPorcentagem * bateria                     // bateria em kW
	var precoTotal float64 = (capacidadeBateria - bateriaEmKW) * precoPorKWh // Preço total da recarga

	return precoTotal
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

	buffer := make([]byte, 1024) // buffer para receber dados do servidor
	// Loop para receber mensagens do servidor
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Erro ao receber mensagem do servidor:", err)
			break
		}

		// decodifica a mensagem recebida
		// Exemplo de mensagem recebida: {"id_ponto": "1", "fila": ["carro1", "carro2"]}
		var mensagemRecebida struct {
			IdPonto string   `json:"id_ponto"`
			Fila    []string `json:"fila"`
		}

		err = json.Unmarshal(buffer[:n], &mensagemRecebida)
		if err != nil {
			fmt.Println("Erro ao decodificar mensagem recebida:", err)
			continue
		}

		// exibe a fila recebida
		fmt.Printf("Fila do ponto %s atualizada: %v\n", mensagemRecebida.IdPonto, mensagemRecebida.Fila)

		// Processa o primeiro carro da fila recebida
		processarFila(mensagemRecebida.IdPonto, mensagemRecebida.Fila)

		precoRecarga := calculaPrecoRecarga(dadosVeiculos.Veiculos[0].BateryLevel)
		preco := fmt.Sprintf("PONTO: Veiculo %s - R$ %.2f\n", dadosVeiculos.Veiculos[0].Placa, precoRecarga)

		// Envia o preço da recarga de volta ao servidor
		_, err = conn.Write([]byte(preco))
		if err != nil {
			fmt.Println("Erro ao enviar mensagem:", err)
		}

		efetivarPagamento(dadosVeiculos.Veiculos[0].Placa, mensagemRecebida.IdPonto, precoRecarga)
	}
}
