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

// procura um veículo na lista de veículos cadastrados (dadosVeiculos.Veiculos) com base na placa
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
			conn, ok := pontosConns[idPonto]
			if !ok {
				fmt.Printf("Erro: Conexão para o ponto %s não encontrada no mapa!\n", idPonto)
				return
			}
			_, err = conn.Write(append(mensagemJson, '\n')) //\n indica o fim da mensagem
			//fmt.Printf("\nEnviado Fila Atualizada para o ponto %s: %s\n", idPonto, string(mensagemJson))
			if err != nil {
				fmt.Println("Erro ao enviar mensagem:", err)
				return
			}

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
			placaTratada := strings.TrimPrefix(placaVeiculo, "Placa ")
			dadosPontos.Pontos[i].Fila = append(ponto.Fila, placaTratada)
			encontrado = true

			// imprime a fila atualizada do ponto
			fmt.Printf("\nFila atualizada do ponto %s, adicionando o veículo: %v\n", idPonto, dadosPontos.Pontos[i].Fila)

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

func removeFila(idPonto string, placaVeiculo string) {
	mutex.Lock()
	defer mutex.Unlock()

	encontrado := false
	placaRemovida := false

	for i, ponto := range dadosPontos.Pontos {
		if strings.TrimSpace(ponto.Id) == strings.TrimSpace(idPonto) {
			encontrado = true

			// Filtra a fila removendo a placa
			novaFila := []string{}
			//fmt.Println("fila:", ponto.Fila)
			for _, placa := range ponto.Fila {
				// Remover o prefixo "Placa "
				placaLimpa := strings.TrimPrefix(strings.TrimSpace(placa), "Placa ")
				if placaLimpa != strings.TrimSpace(placaVeiculo) {
					novaFila = append(novaFila, placa)
				} else {
					placaRemovida = true
				}
			}

			dadosPontos.Pontos[i].Fila = novaFila

			if placaRemovida {
				fmt.Printf("Veículo com placa %s removido da fila do ponto %s.\n", placaVeiculo, idPonto)
				//sendFila(idPonto)
				salvarDados(dadosPontos)
			} else {
				fmt.Printf("Placa %s não encontrada na fila do ponto %s.\n", placaVeiculo, idPonto) //
			}

			break
		}
	}
	if !encontrado {
		fmt.Printf("Erro: Ponto de recarga com ID %s não encontrado.\n", idPonto)
	}
}
