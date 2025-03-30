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
	Id           string      `json:"ID"`
	Placa        string      `json:"placa"`
	Localizacao  Localizacao `json:"localizacao"`
	NivelBateria int         `json:"nivel_bateria"`
}

type Dados struct {
	Veiculos []Veiculo `json:"veiculos"`
}

// lÊ arquivo csv
// Retorna um slice de pontos e um erro (Se tiver erro)
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

	//Este polígono representa uma área geográfica válida onde os veículos podem estar

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

// Ler os dados do veiculo
func lerArquivoJson() {

	// Ler o arquivo JSON usando os.ReadFile
	bytes, err := os.ReadFile("/dados/dadosVeiculos.json")
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

// Função para salvar o JSON atualizado
func salvarArquivoJson() error {
	//dados -> é uma variavel global
	bytes, err := json.MarshalIndent(dados, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile("/dados/dadosVeiculos.json", bytes, 0644) //0644 é a permissão para alterar o nosso json
	return err
}

func main() {
	//ler os dados do veiculo
	lerArquivoJson()

	//coordenadas do mapa de feira
	polygon, err := readPolygon("/dados/MapaDeFeira.csv")
	if err != nil {
		fmt.Println("Erro ao ler arquivo CSV:", err)
		return
	}

	if len(polygon) == 0 {
		fmt.Println("Nenhum ponto encontrado no arquivo CSV")
		return
	}

	// Atualiza coordenadas de todos veiculos do arquivo json (para ser sempre rondomico)
	for i := range dados.Veiculos {
		randomCoord := randomPointInBoundingBox(polygon)
		dados.Veiculos[i].Localizacao.Latitude = randomCoord.Latitude
		dados.Veiculos[i].Localizacao.Longitude = randomCoord.Longitude
	}

	// Salva o JSON atualizado
	err = salvarArquivoJson()
	if err != nil {
		fmt.Println("Erro ao salvar arquivo JSON:", err)
		return
	}
	fmt.Println("Arquivo JSON atualizado com sucesso!")

	// Ler novamente o arquivo json para pegar as coordenadas atualizadas dos veiculos
	//lerArquivoJson()

	//Faz conexão com a nossa porta
	//conn -> representa nossa conexão/rede
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	// Lendo a variável de ambiente do docker compose (ID) para atrelar dados do json para cada conteiner
	veiculoID := os.Getenv("ID-VEICULO")
	if veiculoID == "" {
		fmt.Println("Erro: ID-VEICULO não definido")
		return
	}

	for _, veiculo := range dados.Veiculos {
		if veiculo.Id == veiculoID {
			if veiculo.NivelBateria <= 20 {
				//estrutura da mensagem: quem enviou + id + qual tipo de requisição
				//bb -> bateria baixa
				mensagem := "veiculo," + veiculo.Id + ",bb"                                //passa pro servidor informando que é um veiculo + o ID do veiculo
				fmt.Println("VEICULO " + veiculo.Placa + " com nível de bateria critico!") //envia mensagem ao servidor
				//manda o ID do veículo
				conn.Write([]byte(mensagem))
			}
		}

	}

	//for _, veiculo := range dados.Veiculos {
	//if veiculo.NivelBateria <= 20 {
	//randomCoord := randomPointInBoundingBox(polygon)
	//define mensagem
	//mensagem := fmt.Sprintf("VEICULO %s - Bateria: %d%% - Latitude: %f, Longitude: %f\n",
	//veiculo.Placa, veiculo.NivelBateria, randomCoord.Latitude, randomCoord.Longitude)
	//fmt.Println("Veículo enviado ao servidor:", mensagem) //envia mensagem
	//_, err := conn.Write([]byte(mensagem))
	//if err != nil {
	//fmt.Println("Erro ao enviar mensagem:", err)
	//return
	//}
	//}
	//}
}
