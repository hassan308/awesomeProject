#!/bin/bash

# Bygg applikationen
echo "🔨 Bygger applikationen..."
go run build.go

if [ $? -eq 0 ]; then
    echo "✅ Bygget lyckades!"
    
    # Kopiera .env till rätt plats
    cp .env awesome-cv.env
    
    # Kör applikationen med explicit .env-sökväg
    echo "🚀 Startar applikationen..."
    ENV_FILE_PATH="./awesome-cv.env" ./awesome-cv.exe
else
    echo "❌ Bygget misslyckades!"
    exit 1
fi
