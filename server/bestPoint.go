package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// struct para armazenar os pontos de recarga
type ChargePoint struct {
	Id        string
	Latitude  float64
	Longitude float64
	Nome      string
}

type PontoInfo struct {
	Ponto     string
	PontoID     string
	Distancia float64
	Fila      []string
}

func readChargingPoints(filename string) ([]ChargePoint, error) {
	leArquivoJsonPontos()

	file, err := os.Open(filename) // abre o arquivo CSV
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
			id := records[i][1]

			chargePoints = append(chargePoints, ChargePoint{
				Id:        id,
				Latitude:  lat,
				Longitude: lon,
			})

		}
	}
	//for _, ponto := range chargePoints {
	//fmt.Printf(" | %s | %.6f | %.6f |\n", ponto.Id, ponto.Latitude, ponto.Longitude)
	//}

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

func analiseTodosPontos(lat float64, lon float64, bateria int, placa string) (string, PontoInfo) {

	leArquivoJsonPontos() //lendo os arquivos do ponto

	var pontosOrdenados []PontoInfo

	//lê os pontos de recarga do arquivo csv
	points, err := readChargingPoints("MapaDeFeira.csv")
	if err != nil {
		fmt.Println("Error reading csv:", err)
		//return
	}

	for _, point := range points {
		dist := calculateDistance(lat, lon, point.Latitude, point.Longitude)
		pontoEncontrado, controle := getPonto(point.Id)
		if controle {
			p := PontoInfo{
				Ponto:      pontoEncontrado.Nome,
				PontoID:    pontoEncontrado.Id,
				Distancia:  dist,
				Fila:       pontoEncontrado.Fila,
			}
			pontosOrdenados = append(pontosOrdenados, p)

		} else {
			fmt.Printf("Ponto não encontrado!\n")
			break
		}
	}

	//comparar em relação ao custo -> o custo é uma variavel relacionando o tempo de distancia e o tempo médio de fila
	sort.Slice(pontosOrdenados, func(i, j int) bool {
		var tempoTotal1 float64
		var tempoTotal2 float64

		if len(pontosOrdenados[i].Fila) != 0 {
			for _, c := range pontosOrdenados[i].Fila {
				carro, achou := getVeiculo(c)
				if achou {
					tempoTotal1 += calcularTempoCargaHoras(carro.BateryLevel)
				}
			}
		}else {
			tempoTotal1 = 0.0
		}

		if len(pontosOrdenados[j].Fila) != 0 {
			for _, v := range pontosOrdenados[j].Fila {
				carro, achou := getVeiculo(v)
				if achou {
					tempoTotal2 += calcularTempoCargaHoras(carro.BateryLevel)
				}
			}
		}else {
			tempoTotal2 = 0.0
		}

		custoI := tempoDistancia(pontosOrdenados[i].Distancia) + tempoTotal1
		custoJ := tempoDistancia(pontosOrdenados[j].Distancia) + tempoTotal2
		return custoI < custoJ
	})

	//Tendo a fila ordenada agora pegamos o primeiro elemento
	melhor := pontosOrdenados[0]

	mensagem := fmt.Sprintf("Melhor ponto para o veículo %s: %s - Distância: %.2fKm - Fila: %d veículos\n", placa, melhor.Ponto, melhor.Distancia, len(melhor.Fila))
	return mensagem, melhor
}

func tempoDistancia(dist float64) float64 {
	//considerando que todos carros rodem em uma media de 60km/h
	horas := dist / 60
	return horas
}

// Calcula o tempo de carregamento em horas
func calcularTempoCargaHoras(nivelBateria int) float64 {
	//transforma a bateria de int pra float
	var nivelInicial float64 = float64(nivelBateria)

	//Presumindo que todos carregadores tem a potencia de 150 kW (sendo um carregador rápido - nível 3)
	var potenciaCarregador float64 = 150
	//E que a capacidade total da bateria de todo carro elétrico seja de 100 kWh
	var kwhBateria float64 = 100

	if nivelInicial >= 100 {
		fmt.Print("Este carro ja esta 100%% carregado.\n\n")
	}

	// Energia total a carregar (em kWh)
	energiaRestante := kwhBateria * ((100 - nivelInicial) / 100)

	// Separar a carga em duas fases:
	// 1. Até 80% (carga rápida)
	// 2. De 80% a 100% (carga lenta)

	// Quantos % ainda faltam até 80%
	ate80 := 80 - nivelInicial
	if ate80 < 0 {
		ate80 = 0
	}

	// Energia da fase rápida (até 80%)
	energiaFase1 := kwhBateria * (ate80 / 100)

	// Energia da fase lenta (80% a 100%)
	energiaFase2 := energiaRestante - energiaFase1

	// Tempo da fase 1: usando potência cheia
	tempoFase1 := energiaFase1 / potenciaCarregador

	// Tempo da fase 2: usando potência reduzida (~40 kW)
	potenciaReduzida := math.Min(potenciaCarregador, 40)
	tempoFase2 := 0.0
	if energiaFase2 > 0 {
		tempoFase2 = energiaFase2 / potenciaReduzida
	}

	// Tempo total em horas
	tempoTotalHoras := tempoFase1 + tempoFase2

	return tempoTotalHoras
}
