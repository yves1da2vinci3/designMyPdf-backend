FROM golang:1.22-bullseye AS builder

WORKDIR /app

# Installer swag pour la génération de la documentation Swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copier d'abord seulement go.mod et go.sum
COPY go.mod go.sum ./

# Télécharger les dépendances et les mettre en cache
RUN go mod download

# Copier le reste du code source
COPY . .

# Générer la documentation Swagger
RUN swag init

# Générer les fichiers vendor pour assurer que toutes les dépendances sont disponibles localement
RUN go mod vendor

# Compiler l'application avec vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o app .

# Image finale avec Chrome pour la génération de PDF
FROM debian:bullseye-slim

# Installer Chrome et les dépendances nécessaires
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    ca-certificates \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libwayland-client0 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxkbcommon0 \
    libxrandr2 \
    xdg-utils \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Installer Google Chrome
RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - \
    && echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google.list \
    && apt-get update \
    && apt-get install -y google-chrome-stable --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Créer les répertoires pour les uploads et les fichiers temporaires
RUN mkdir -p /app/uploads/template /app/tmp

WORKDIR /app

# Copier le binaire compilé depuis l'étape précédente
COPY --from=builder /app/app .

# Copier les fichiers de configuration et docs
COPY --from=builder /app/config ./config
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/.env ./.env

# Exposer le port utilisé par l'application
EXPOSE 5000

# Définir les variables d'environnement nécessaires
ENV PORT=5000
ENV CHROME_PATH=/usr/bin/google-chrome
ENV GO111MODULE=on

# Commande pour démarrer l'application
CMD ["./app"] 