# DesignMyPDF Backend

## Déploiement avec Docker

Ce projet peut être facilement déployé à l'aide de Docker et Docker Compose.

### Prérequis

- Docker
- Docker Compose

### Variables d'environnement

Créez un fichier `.env` à la racine du projet avec les variables suivantes :

```
# Paramètres du serveur
PORT=5000

# Connexion à la base de données (déjà configurée dans docker-compose.yml)
DATABASE_URL=postgres://postgres:postgres@postgres:5432/designmypdf

# Backblaze B2 (recommandé — aligné sur frontend/services/backblazeService.ts)
# Premier champ = Application Key ID (souvent préfixe 005…), pas seulement l’ID compte sauf clé maître.
BACKBLAZE_KEY_ID=
BACKBLAZE_APP_KEY=
BACKBLAZE_BUCKET_NAME=

# Ancien nommage (toujours supporté si BACKBLAZE_* vides)
# B2_ACCOUNT_ID doit contenir la même valeur que BACKBLAZE_KEY_ID (Application Key ID).
# B2_ACCOUNT_ID=
# B2_APPLICATION_KEY=
# B2_BUCKET_NAME=

# Variables d'authentification (si nécessaire)
JWT_SECRET=your_jwt_secret
```

### Construction et démarrage

Pour construire et démarrer l'application :

```bash
# Construire l'image Docker
docker-compose build

# Démarrer les services
docker-compose up -d
```

L'API sera disponible à l'adresse : http://localhost:5000

### Génération de PDF

L'application utilise Chromium (headless) pour générer des PDFs à partir de templates HTML. Cette fonctionnalité est déjà incluse dans l'image Docker.

### Structure des dossiers

- `/uploads` : Stockage temporaire des PDFs générés avant leur upload sur Backblaze B2
- `/tmp` : Fichiers temporaires
- `/config` : Fichiers de configuration
- `/api` : Routes et handlers API
- `/pkg` : Packages core de l'application
- `/utils` : Utilitaires divers

### Logs

Les logs de l'application sont accessibles via :

```bash
docker-compose logs -f backend
```
