# Buildar as imagens
docker-compose build

# Criar os containers mas não iniciar
docker-compose create

# Caso queira já iniciar: descomente a linha abaixo
# docker-compose up -d

# Mostrar todos os containers
docker ps -a

# Inicia conteiner de Service
docker-compose start server

# Inicia todos pontos
docker-compose start ponto ponto2 ponto3 ponto4 ponto5

# Inicia veiculo 1
#docker-compose start veiculo