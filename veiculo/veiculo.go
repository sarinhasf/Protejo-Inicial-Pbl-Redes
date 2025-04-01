package main

import (
	//"bufio"         //pacote para leitura de dados
	"encoding/csv"  //pacote para manipulação de CSV
	"encoding/json" //pacote para manipulação de JSON
	"fmt"           //pacote para formatação de strings
	"math/rand"     //pacote para gerar números aleatórios
	"net"           //pacote para comunicação em rede
	"os"            //pacote para manipulação de arquivos
	"strconv"       //conversão de tipos
	"strings"       //manipulação de strings
)

// Definindo estrutura com os dados dos veiculos
var dados Dados

// struct para armazenar coordenadas
type Point struct {
	Latitude  float64
	Longitude float64
}

// Criando structs apartir do JSON
type Localizacao struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Veiculo struct {
	Placa        string      `json:"placa"`
	Localizacao  Localizacao `json:"localizacao"`
	NivelBateria int         `json:"nivel_bateria"`
}

type Dados struct {
	Veiculos []Veiculo `json:"veiculos"`
}

// lÊ arquivo csv
// Retorna um slice de pontos e um erro
func readPolygon(filename string) ([]Point, error) {
	file, err := os.Open(filename) //abre o arquivo
	if err != nil {
		return nil, err
	}
	defer file.Close() //fecha o arquivo

	reader := csv.NewReader(file)    //cria um leitor de csv
	records, err := reader.ReadAll() //lê todos os registros do arquivo
	//records -> matriz de strings
	if err != nil {
		return nil, err
	}

	//Pega o segundo registro e o primeiro elemento
	//Remove os prefixos e sufixos
	rawData := records[1][0]
	rawData = strings.TrimPrefix(rawData, "POLYGON ((")
	rawData = strings.TrimSuffix(rawData, "))")

	var polygon []Point                    //cria um slice de pontos
	coords := strings.Split(rawData, ", ") //separa as coordenadas por vírgula
	for _, coord := range coords {
		parts := strings.Split(coord, " ") //separa as partes da coordenada
		if len(parts) != 2 {               //se não tiver duas partes,
			continue //pula para a próxima iteração
		}
		longitude, _ := strconv.ParseFloat(parts[0], 64) //converte a string para float64
		latitude, _ := strconv.ParseFloat(parts[1], 64)
		polygon = append(polygon, Point{Latitude: latitude, Longitude: longitude})
	}

	return polygon, nil
}

// Gera um ponto aleatório dentro de um polígono
// Retorna um ponto
func randomPointInBoundingBox(polygon []Point) Point {
	minLat, maxLat := polygon[0].Latitude, polygon[0].Latitude
	minLon, maxLon := polygon[0].Longitude, polygon[0].Longitude

	for _, p := range polygon {
		if p.Latitude < minLat {
			minLat = p.Latitude
		}
		if p.Latitude > maxLat {
			maxLat = p.Latitude
		}
		if p.Longitude < minLon {
			minLon = p.Longitude
		}
		if p.Longitude > maxLon {
			maxLon = p.Longitude
		}
	}

	for {
		//Gera um ponto aleatório dentro de uma bounding box
		//bounding box -> caixa que envolve um objeto
		lat := minLat + rand.Float64()*(maxLat-minLat)
		lon := minLon + rand.Float64()*(maxLon-minLon)
		p := Point{Latitude: lat, Longitude: lon}
		if isPointInsidePolygon(p, polygon) {
			return p //retorna o ponto se estiver dentro do polígono
		}
	}
}

// Verifica se um ponto está dentro do polígono
// Retorna um booleano (true se estiver dentro, false caso contrário)
func isPointInsidePolygon(point Point, polygon []Point) bool {
	inside := false
	j := len(polygon) - 1
	for i := 0; i < len(polygon); i++ {
		if (polygon[i].Latitude > point.Latitude) != (polygon[j].Latitude > point.Latitude) &&
			point.Longitude < (polygon[j].Longitude-polygon[i].Longitude)*(point.Latitude-polygon[i].Latitude)/(polygon[j].Latitude-polygon[i].Latitude)+polygon[i].Longitude {
			inside = !inside
		}
		j = i
	}
	return inside
}

func leArquivoJson() {
	// Ler o arquivo JSON usando os.ReadFile
	bytes, err := os.ReadFile("dadosVeiculos.json")
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

func main() {
	leArquivoJson()

	polygon, err := readPolygon("MapaDeFeira.csv")
	if err != nil {
		fmt.Println("Erro ao ler arquivo CSV:", err)
		return
	}

	if len(polygon) == 0 {
		fmt.Println("Nenhum ponto encontrado no arquivo CSV")
		return
	}

	// Lendo a variável de ambiente do docker compose
	veiculoID := os.Getenv("PLACA")
	if veiculoID == "" {
		fmt.Println("Erro: PLACA não definida")
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

	for _, veiculo := range dados.Veiculos {
		if veiculo.Placa == veiculoID && veiculo.NivelBateria <= 20 {
			randomCoord := randomPointInBoundingBox(polygon)
			//define mensagem
			mensagem := fmt.Sprintf("VEICULO %s | Bateria: %d%% | Latitude: %f | Longitude: %f \n",
				veiculo.Placa, veiculo.NivelBateria, randomCoord.Latitude, randomCoord.Longitude)
			fmt.Println("Veículo enviado ao servidor:", mensagem)
			_, err := conn.Write([]byte(mensagem)) //envia mensagem
			if err != nil {
				fmt.Println("Erro ao enviar mensagem:", err)
				return
			}

			//reader := bufio.NewReader(conn)          //cria um leitor de dados
			//response, err := reader.ReadString('\n') //lê a resposta do servidor
			/*if err != nil {
				fmt.Println("Erro ao ler resposta do servidor:", err)
				return
			}
			fmt.Println("Resposta do servidor:", response)*/
		}
	}
}
