# ─── Stage 1: Build Angular aplikacije ──────────────────────────────────────
FROM node:20-alpine AS build
WORKDIR /app

# Layer caching: prvo zavisnosti
COPY package*.json ./
RUN npm ci

# Ostatak koda + produkcijski build (izlaz: dist/shop-frontend/browser)
COPY . .
RUN npm run build

# ─── Stage 2: Serviranje preko nginx-a ──────────────────────────────────────
FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist/shop-frontend/browser /usr/share/nginx/html
EXPOSE 80
