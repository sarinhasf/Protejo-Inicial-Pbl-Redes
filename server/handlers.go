package main

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

func handleVeiculo(sessao *SessaoCliente, mensagem string) {
	conn := sessao.Conn

	if strings.Contains(mensagem, "VEICULO CONECTADO") {
		pontosConns = append(pontosConns, conn)
		fmt.Println("Novo veiculo conectado!")

	} else if strings.HasPrefix(mensagem, "VEICULO |") {
		// primeira mensagem que o veiclo envia, com as informações do veiculo
		placa, lat, lon := trataInfo(mensagem)
		sessao.PlacaVeiculo = placa
		veiculosConns[placa] = conn
		fmt.Printf("Conexão registrada para o veículo %s\n", placa)

		closestPoint, distance := pegaPontoProximo(lat, lon)

		var pontoID = closestPoint.Id
		var filaPonto []string
		var nomePontoProx string

		for _, ponto := range dadosPontos.Pontos {
			if strings.EqualFold(ponto.Id, pontoID) {
				filaPonto = ponto.Fila
				nomePontoProx = ponto.Nome
				break
			}
		}

		if len(filaPonto) == 0 {
			msg := fmt.Sprintf("Melhor ponto para o veículo %s - Distância: %.2fKm - Fila: %d veículos \n", nomePontoProx, distance, len(filaPonto))
			conn.Write([]byte(msg))

			sessao.AguardandoResposta = true
			sessao.MelhorPontoID = pontoID
			sessao.MelhorPontoNome = nomePontoProx

		} else if veiculo, ok := getVeiculo(placa); ok {
			msg, melhorPonto := analiseTodosPontos(lat, lon, veiculo.BateryLevel, placa)
			conn.Write([]byte(msg))

			sessao.AguardandoResposta = true
			sessao.MelhorPontoID = melhorPonto.Ponto.Id
			sessao.MelhorPontoNome = melhorPonto.Ponto.Nome
		}

	} else if sessao.AguardandoResposta {
		resposta := strings.TrimPrefix(mensagem, "VEICULO")
		resposta = strings.TrimSpace(resposta)
		resposta = strings.ToLower(resposta)
		fmt.Printf("Resposta do veículo %s:%s\n", sessao.PlacaVeiculo, resposta)

		if resposta == "sim" {
			addFila(sessao.MelhorPontoID, sessao.PlacaVeiculo)
			confirmacao := fmt.Sprintf("Veículo %s adicionado à fila do ponto de recarga %s\n", sessao.PlacaVeiculo, sessao.MelhorPontoNome)
			conn.Write([]byte(confirmacao + "\n"))

			// Envia JSON da fila para o ponto
			sendFila(sessao.MelhorPontoID)

		} else {
			fmt.Printf("Veículo %s recusou entrar na fila\n", sessao.PlacaVeiculo)
		}

		sessao.AguardandoResposta = false
		return
	}
}

func handlePonto(sessao *SessaoCliente, mensagem string) {
	conn := sessao.Conn

	if strings.HasPrefix(mensagem, "PONTO DE RECARGA CONECTADO") {
		pontosConns = append(pontosConns, conn)
		fmt.Println("Novo ponto de recarga conectado!")

	} else if strings.HasPrefix(mensagem, "PONTO: Veiculo ") {
		fmt.Println(mensagem)
		placaRegex := regexp.MustCompile(`[A-Z]{3}[0-9][A-Z0-9][0-9]{2}`)
		placa := placaRegex.FindString(mensagem)
		fmt.Println("placa:", placa)

		// Regex para o número do ponto (após "Ponto ")
		pontoRegex := regexp.MustCompile(`Ponto (\d+)`)
		pontoMatch := pontoRegex.FindStringSubmatch(mensagem)
		var ponto string
		if len(pontoMatch) > 1 {
			ponto = pontoMatch[1]
		}
		ponto = strings.TrimSpace(ponto)
		fmt.Println("numero do ponto:", ponto)

		//removeFila(ponto, placa)
		//sendFila(ponto)

		//sendToVehicle(placa, mensagem)

	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // garante que a conexão será fechada ao final da função

	sessao := &SessaoCliente{
		Conn: conn,
	}

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

			if sessao.Tipo == "" {
				if strings.HasPrefix(mensagem, "VEICULO ") {
					sessao.Tipo = TipoCliente(TipoVeiculo)

				} else if strings.HasPrefix(mensagem, "PONTO") {
					sessao.Tipo = TipoCliente(TipoPonto)

				} else {
					fmt.Println("Tipo de cliente não reconhecido:", mensagem)
					return
				}
			}

			switch sessao.Tipo {
			case TipoCliente(TipoVeiculo):
				handleVeiculo(sessao, mensagem)
			case TipoCliente(TipoPonto):
				handlePonto(sessao, mensagem)
			}
		}
		//atualiza o buffer acumulado para manter apenas a última mensagem incompleta
		bufferAcumulado = mensagens[len(mensagens)-1]
	}
}
