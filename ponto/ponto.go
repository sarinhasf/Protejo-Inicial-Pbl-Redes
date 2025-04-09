package main

import (
	"encoding/json" //pacote para manipulação de JSON
	"fmt"
	"net"
	"os"
	"sync"
	"time" //pacote para manipulação de tempo
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
	Id         int
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
	IdConta     int      `json:"conta_id"`
}

// struct para armazenar Dados das contas
type DadosContas struct {
	Contas []ContaUser `json:"contas"`
}

type DadosVeiculos struct {
	Veiculos []Veiculo `json:"veiculos"`
}

type PontoRecarga struct {
	Id         string
	Nome       string
	Fila       []string
	Carregando string
}

type DadosPontos struct {
	Pontos []PontoRecarga `json:"pontos"`
}

// Estrutura para o histórico do ponto
type Historico struct {
	Carro  string `json:"carro"`
	Status string `json:"status"`
}

var dadosContas DadosContas
var dadosVeiculos DadosVeiculos
var dadosPontos DadosPontos

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

func getContaUsuario(id int) (ContaUser, bool) {
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

func processarFila(idPonto string, filaSlice []string) {
	mutex.Lock()         //bloqueia acesso concorrente
	defer mutex.Unlock() //libera depois da execução da função

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
	time.Sleep(30 * time.Second) // Simula o tempo de carregamento (30 segundos para teste)

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
	fmt.Printf("\nFila atualizada do ponto %s, retirando o Veículo %s: %v\n\n", idPonto, carro, filaSlice)

	//Atualizando porcentagem do veiculo
	veiculo, achou := getVeiculo(carro)
	if achou {
		leArquivoJsonVeiculos()
		dadosVeiculos.Veiculos[veiculo.IdConta-1].BateryLevel = 100
		salvarDadosVeiculos(dadosVeiculos)
	} else {
		fmt.Printf("Veiculo com a placa %s não encontrado.", carro)
	}
}

func salvarDadosVeiculos(data DadosVeiculos) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter dados para JSON:", err)
		return
	}

	err = os.WriteFile("dadosVeiculos.json", bytes, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar no arquivo dadosPontos.json:", err)
		return
	}
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
	leArquivoJsonContas()

	veiculo, achou := getVeiculo(idVeiculo)
	contaId := veiculo.IdConta

	if achou {
		contaVeiculo, achou2 := getContaUsuario(contaId)

		if achou2 {
			novoPagamento := Pagamentos{
				IdPonto: idPonto,
				Valor:   valor,
			}
			//contaVeiculo.Pagamentos = append(contaVeiculo.Pagamentos, novoPagamento)
			dadosContas.Contas[contaVeiculo.Id-1].Pagamentos = append(dadosContas.Contas[contaVeiculo.Id-1].Pagamentos, novoPagamento)

			// Salva no arquivo contasUsuarios.json
			salvarDadosContas()
			fmt.Printf("Pagamento registrado com sucesso do Veículo %s.", veiculo.Placa)
		} else {
			fmt.Printf("Conta do Veiculo %s não encontrada!", veiculo.Placa)
		}

	} else {
		fmt.Println("Veiculo não encontrado!")
	}
}

func calculaPrecoRecarga(nivelBateria int) float64 {
	bateriaAtual := float64(nivelBateria) 
	falta := 100.0 - bateriaAtual  //bateria desejada - atual (quanto falta em %)
	var precoPorKWh float64 = 0.50  //valor da tarifa                                         
	var capacidadeBateria float64 = 100.0              
	faltaPorc := falta/100.0                       
	energiaCarregada := faltaPorc * capacidadeBateria   //o resultado disso é o quanto de energia sera carregado o carro     
	var precoTotal float64 = energiaCarregada * precoPorKWh

	return precoTotal
}

func main() {
	leArquivoJsonContas()
	leArquivoJsonVeiculos()
	//leArquivoJsonPontos()

	//Faz conexão
	//conn -> representa nossa conexão/rede
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	//defer conn.Close()

	// Lendo a variável de ambiente do docker compose
	pontoID := os.Getenv("ID-PONTO")
	if pontoID == "" {
		fmt.Println("Erro: PLACA não definida")
		return
	}

	//Envia mensagem
	mensagem := "PONTO DE RECARGA CONECTADO," + pontoID + "\n" //tem que terminar com \n se não o servidor não processa
	fmt.Printf("Registro de Ponto de recarga %s conectado ao servidor.\n", pontoID)

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
			IdPonto    string   `json:"id_ponto"`
			Fila       []string `json:"fila"`
			Carregando bool     `json:"carregando"`
		}

		err = json.Unmarshal(buffer[:n], &mensagemRecebida)
		if err != nil {
			fmt.Println("Erro ao decodificar mensagem recebida:", err)
			continue
		}

		// exibe a fila recebida
		fmt.Printf("\nFila do ponto %s atualizada: %v\n", mensagemRecebida.IdPonto, mensagemRecebida.Fila)

		// Processa o primeiro carro da fila recebida
		processarFila(mensagemRecebida.IdPonto, mensagemRecebida.Fila)

		placa := mensagemRecebida.Fila[0] //pega o primeiro elemento da fila

		veiculoSearch, achou := getVeiculo(placa) //pega o veiculo pela placa (placa)

		if achou {
			precoRecarga := calculaPrecoRecarga(veiculoSearch.BateryLevel)
			preco := fmt.Sprintf("MENSAGEM DO PONTO: Veiculo %s carregado no Ponto %s - Valor da Recarga: R$ %.2f\n", veiculoSearch.Placa, mensagemRecebida.IdPonto, precoRecarga)
			efetivarPagamento(veiculoSearch.Placa, mensagemRecebida.IdPonto, precoRecarga)

			// Envia o preço da recarga de volta ao servidor e confirma o registro do pagamento
			_, err = conn.Write([]byte(preco))
			if err != nil {
				fmt.Println("Erro ao enviar mensagem:", err)
			}

		} else {
			fmt.Println("\nNão encontrato veiculo com a placa: ", placa)
		}

	}
}
