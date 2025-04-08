package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Lê o arquivo JSON e armazena os dados na variável global "dados"
func leArquivoJson(filename string) {
	// Verifica se o arquivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("Arquivo JSON não encontrado:", filename)
		return
	}
	// Ler o arquivo JSON usando os.ReadFile
	bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Erro ao abrir arquivo JSON:", err)
		return
	}

	// Passando dados do JSON para struct criada
	err = json.Unmarshal(bytes, &dados)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}
}

func salvarDadosVeiculos(data Dados) {
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

func leMapaFeira() []Point {
	polygon, err := readPolygon("MapaDeFeira.csv")
	if err != nil {
		fmt.Println("Erro ao ler arquivo CSV:", err)
		return nil
	}
	if len(polygon) == 0 {
		fmt.Println("Nenhum ponto encontrado no arquivo CSV")
		return nil
	}
	return polygon
}