#!/bin/bash

# Buildar as imagens
docker-compose build

docker-compose create # Criar os containers e mas não roda
#docker-compose up -d #Criar todos containers e já roda

# Mostrar todos os containers criados
docker ps -a
