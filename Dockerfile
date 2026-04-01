# --- ÉTAPE 1 : BUILDER (Compilation de l'application Go) ---
    FROM golang:1.22-bullseye AS builder

    WORKDIR /app
    
    # 1. Installer l'outil Swagger (swag)
    RUN go install github.com/swaggo/swag/cmd/swag@latest
    
    # 2. Gérer les dépendances (utilisons le cache Docker au maximum)
    COPY go.mod go.sum ./
    RUN go mod download
    
    # 3. Copier le reste du code source
    COPY . .
    
    # 4. Générer la documentation Swagger
    RUN swag init
    
    # 5. Compiler le binaire de manière statique
    # -ldflags="-s -w" permet de réduire la taille du binaire final
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app .
    
    # --- ÉTAPE 2 : FINAL (Image légère pour l'exécution) ---
    FROM debian:bullseye-slim
    
    # Éviter les prompts interactifs d'apt
    ENV DEBIAN_FRONTEND=noninteractive
    
    # Installation de Chromium et des certificats CA (nécessaires pour HTTPS/PDF)
    # On ne liste que le strict nécessaire, Debian gère les dépendances de Chromium
    RUN apt-get update && apt-get install -y --no-install-recommends \
        chromium \
        fonts-liberation \
        ca-certificates \
        && apt-get clean \
        && rm -rf /var/lib/apt/lists/*
    
    WORKDIR /app
    
    # Création des dossiers nécessaires au runtime
    RUN mkdir -p /app/uploads/template /app/tmp /app/config /app/docs
    
    # Copier uniquement ce qui est nécessaire depuis le builder
    COPY --from=builder /app/app .
    COPY --from=builder /app/config ./config
    COPY --from=builder /app/docs ./docs
    COPY --from=builder /app/.env ./.env
    
    # Variables d'environnement pour l'application
    ENV CHROME_PATH=/usr/bin/chromium
    ENV PORT=5000
    ENV GO111MODULE=on
    
    # Exposer le port
    EXPOSE 5000
    
    # Lancement de l'application
    CMD ["./app"]