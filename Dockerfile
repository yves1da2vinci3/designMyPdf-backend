# --- ÉTAPE 1 : BUILDER ---
    FROM golang:1.22-bullseye AS builder

    WORKDIR /app
    
    # Installation de swag
    RUN go install github.com/swaggo/swag/cmd/swag@latest
    
    # Cache des dépendances
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copie du code et génération swagger
    COPY . .
    RUN swag init
    
    # Compilation statique optimisée
    # On retire les symboles de debug (-s -w) pour alléger le binaire
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app .
    
    # --- ÉTAPE 2 : FINAL ---
    FROM debian:bullseye-slim
    
    ENV DEBIAN_FRONTEND=noninteractive
    
    # Installation de Chromium et des libs de base
    RUN apt-get update && apt-get install -y --no-install-recommends \
        chromium \
        fonts-liberation \
        ca-certificates \
        && apt-get clean \
        && rm -rf /var/lib/apt/lists/*
    
    WORKDIR /app
    
    # Création des répertoires de travail
    RUN mkdir -p /app/uploads/template /app/tmp /app/config /app/docs
    
    # Copie des fichiers nécessaires depuis le builder
    COPY --from=builder /app/app .
    COPY --from=builder /app/config ./config
    COPY --from=builder /app/docs ./docs
    
    # NOTE : Ne JAMAIS copier le .env ici. 
    # Coolify injecte les variables d'environnement au lancement du container.
    
    ENV CHROME_PATH=/usr/bin/chromium
    ENV PORT=5000
    
    EXPOSE 5000
    
    CMD ["./app"]