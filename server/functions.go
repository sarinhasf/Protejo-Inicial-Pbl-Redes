package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)


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

func salvarDados(data DadosPontos) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter dados para JSON:", err)
		return
	}

	err = os.WriteFile("dadosPontos.json", bytes, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar no arquivo dadosPontos.json:", err)
		return
	}
}


//procura um veículo na lista de veículos cadastrados (dadosVeiculos.Veiculos) com base na placa 
func getVeiculo(placa string) (Veiculo, bool) {
	var veiculoFinal Veiculo
	controle := false

	for _, veiculo := range dadosVeiculos.Veiculos {
		if veiculo.Placa == placa {
			veiculoFinal = veiculo
			controle = true
		}
	}
	// retorna o veículo encontrado (do tipo Veiculo), 
	// e um booleano indicando se o veículo foi encontrado (true) ou não (false)
	return veiculoFinal, controle
}

func getPonto(id string) (PontoRecarga, bool) {
	var pontoFinal PontoRecarga
	controle := false

	for _, ponto := range dadosPontos.Pontos {
		if ponto.Id == id {
			pontoFinal = ponto
			controle = true
		}
	}
	return pontoFinal, controle
}


func sendFila(idPonto string) {
	for _, ponto := range dadosPontos.Pontos {
		if ponto.Id == idPonto {
			// cria um objeto contendo o id do ponto e a fila
			mensagemStruct := struct {
				IdPonto string   `json:"id_ponto"`
				Fila    []string `json:"fila"`
			}{
				IdPonto: ponto.Id,
				Fila:    ponto.Fila,
			}

			// converte objeto para json
			mensagemJson, err := json.Marshal(mensagemStruct)
			if err != nil {
				fmt.Println("Erro ao converter mensagem para JSON:", err)
				return
			}

			// envia json para o cliente
			_, err = pontosConns[0].Write(append(mensagemJson, '\n')) //\n indica o fim da mensagem
			if err != nil {
				fmt.Println("Erro ao enviar mensagem:", err)
				return
			}

			fmt.Printf("Mensagem enviada para o ponto %s: %s\n", idPonto, string(mensagemJson))
			break
		}
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

			// imprime a fila atualizada do ponto
			fmt.Printf("Fila atualizada do ponto %s: %v\n", idPonto, dadosPontos.Pontos[i].Fila)

			sendFila(idPonto) // Envia a fila atualizada para o ponto de recarga

			break
		}
	}
	if encontrado {
		salvarDados(dadosPontos) // Salva os dados atualizados no arquivo JSON
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
