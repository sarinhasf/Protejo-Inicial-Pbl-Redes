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
)

// struct para armazenar os pontos de recarga
type ChargePoint struct {
	Latitude  float64
	Longitude float64
	Nome      string
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

func main() {
	//lê os pontos de recarga do arquivo csv
	/*points, err := readChargingPoints("MapaDeFeira.csv")
	if err != nil {
		fmt.Println("Error reading csv:", err)
		return
	}*/

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
