FROM node:22-alpine
WORKDIR /app
# Ne pas définir NODE_ENV pour permettre l'installation de toutes les deps
COPY package*.json ./
RUN npm ci --legacy-peer-deps
COPY . .
EXPOSE 3001
ENV NEXT_TELEMETRY_DISABLED=1
CMD ["npm","run","dev"]
