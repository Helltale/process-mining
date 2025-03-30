#!/bin/bash

set -e

GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}➡️  Запуск проекта...${NC}"

# Запускаем бэкенд в фоне
echo -e "${GREEN}🔧 Запускаем сервер (в фоне)...${NC}"
go run ./cmd/app/main.go serve &

# Ждём чуть-чуть, чтобы бэкенд стартовал
sleep 1

# Запускаем фронтенд
echo -e "${GREEN}🌐 Запускаем фронтенд...${NC}"
cd frontend
npm install
npm run dev
