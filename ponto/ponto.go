package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6378 // raio da Terra em metros

	// Converte coordenadas para radianos
	lat1Rad, lon1Rad := lat1*math.Pi/180, lon1*math.Pi/180
	lat2Rad, lon2Rad := lat2*math.Pi/180, lon2*math.Pi/180
	// Diferença de latitude e longitude
	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	// Fórmula de Haversine para calcular a distância entre dois pontos
	// https://en.wikipedia.org/wiki/Haversine_formula
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlon/2)*math.Sin(dlon/2) // Sen²((lat2 - lat1) / 2) + cos(lat1) * cos(lat2) * sen²((lon2 - lon1) / 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))                                                              // 2 * atan2(sqrt(a), sqrt(1 - a))
	// c representa o fator de escala para calcular a distância entre dois pontos usando a fórmula de Haversine

	return R * c
}

func findClosestPoint(lat, lon float64, points []ChargePoint) ChargePoint {
	var closestPoint ChargePoint
	minDistance := math.MaxFloat64

	for _, point := range points {
		distance := calculateDistance(lat, lon, point.Latitude, point.Longitude)
		if distance < minDistance {
			minDistance = distance
			closestPoint = point
		}
	}
	return closestPoint
}

func main() {
	//lê os pontos de recarga do arquivo csv
	points, err := readChargingPoints("MapaDeFeira.csv")
	if err != nil {
		fmt.Println("Error reading csv:", err)
		return
	}

	//Faz conexão
	//conn -> representa nossa conexão/rede
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	//Envia mensagem
	mensagem := "PONTO DE RECARGA CONECTADO\n " //tem que terminar com \n se não o servidor não processa
	_, error := conn.Write([]byte(mensagem))
	if error != nil {
		fmt.Println("Erro ao enviar mensagem de registro ao servidor:", err)
		return
	}
	fmt.Println("Registro de Ponto de recarga enviado ao servidor:", mensagem)

	reader := bufio.NewReader(conn)
	for {
		fmt.Println("Aguardando servidor enviar coordenadas do veículo...")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler mensagem do servidor:", err)
			break
		}
		fmt.Println("Mensagem recebida do servidor:", input)

		//confirma que recebeu as cordenadas
		input = strings.TrimSpace(input)
		fmt.Println("Coordenadas recebidas:", input)

		// Envia confirmação ao servidor
		/*message := "CONFIRMACAO PONTO recebeu coordenadas do veículo"
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Erro ao enviar confirmação:", err)
			return
		}
		fmt.Println("Mensagem enviada ao servidor:", message)*/

		parts := strings.Split(input, ",")
		if len(parts) != 2 {
			fmt.Println("Mensagem inválida do servidor:", input)
			continue
		}
		fmt.Println("Coordenadas recebidas:", parts[0], parts[1])

		lat, err1 := strconv.ParseFloat(parts[0], 64)
		lon, err2 := strconv.ParseFloat(parts[1], 64)
		// Verifica se houve erro na conversão
		if err1 != nil || err2 != nil {
			fmt.Println("Erro ao converter coordenadas:", err1, err2)
			continue
		}
		fmt.Println("Coordenadas convertidas:", lat, lon)

		closestPoint := findClosestPoint(lat, lon, points)
		message := fmt.Sprintf("Ponto de recarga mais próximo: %s\n", closestPoint.Nome)

		fmt.Println("Enviando resposta ao servidor:", message)
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Erro ao enviar mensagem ao servidor:", err)
			break
		}
		time.Sleep(3 * time.Second)
	}
}
