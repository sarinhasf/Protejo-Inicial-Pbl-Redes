package main

import (
	"fmt"
	"strconv"
	"strings"
)

func trataInfo(mensagem string) (plaque string, lat, lon float64) {
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
		//Id:          placa, // ID do veículo (pode ser o mesmo que a placa) se for tirar o id ou de fato dar um id diferente da placa
		Placa:       placa,
		Location:    Location{Latitude: latitudeFloat, Longitude: longitudeFloat},
		BateryLevel: bateriaInt,
	}

	dadosVeiculos.Veiculos = append(dadosVeiculos.Veiculos, novoVeiculo) // adiciona o novo veículo à lista de veículos
	/*fmt.Println("Veículos armazenados atualmente:")
	for _, veiculo := range dados.Veiculos {
		fmt.Printf(" | %s | %d%% | %.6f | %.6f |\n", veiculo.Placa, veiculo.BateryLevel, veiculo.Location.Latitude, veiculo.Location.Longitude)
	}*/

	return novoVeiculo.Placa, novoVeiculo.Location.Latitude, novoVeiculo.Location.Longitude
}