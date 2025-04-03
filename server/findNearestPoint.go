package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
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
	const R = 6378 // raio da Terra em quilômetros

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

	return R * c //distancia
}

func getClosestPoint(lat, lon float64, points []ChargePoint) (closestPoint ChargePoint, distância float64) {
	//var closestPoint ChargePoint
	minDistance := math.MaxFloat64

	for _, point := range points {
		distance := calculateDistance(lat, lon, point.Latitude, point.Longitude)
		if distance < minDistance {
			minDistance = distance
			closestPoint = point
		}
	}
	return closestPoint, minDistance
}

func pegaPontoProximo(latitudeCarro, longitudeCarro float64) (closestPoint ChargePoint, distancia float64) {
	//lê os pontos de recarga do arquivo csv
	points, err := readChargingPoints("MapaDeFeira.csv")
	if err != nil {
		fmt.Println("Error reading csv:", err)
		return
	}

	closestPoint, distance := getClosestPoint(latitudeCarro, longitudeCarro, points)

	return closestPoint, distance
}
