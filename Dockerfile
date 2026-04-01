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

# Image finale : Chromium (léger, dépôts Debian) pour la génération de PDF
FROM debian:bullseye-slim

ENV DEBIAN_FRONTEND=noninteractive

# Chromium + dépendances minimales en un seul RUN (moins de couches, moins de RAM au build)
RUN apt-get update && apt-get install -y \
    chromium \
    fonts-liberation \
    libnss3 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libgbm1 \
    libasound2 \
    ca-certificates \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN mkdir -p /app/uploads/template /app/tmp /app/config /app/docs

COPY --from=builder /app/app .
COPY --from=builder /app/config ./config
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/.env ./.env

ENV CHROME_PATH=/usr/bin/chromium
ENV PORT=5000

EXPOSE 5000

CMD ["./app"]
