FROM node:20-alpine AS base

FROM base AS frontend-deps
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci

FROM base AS backend-deps
WORKDIR /app
COPY backend/go.mod backend/go.sum* ./
RUN go mod download

FROM base AS frontend-build
WORKDIR /app
COPY --from=frontend-deps /app/node_modules ./node_modules
COPY frontend/ ./
RUN npm run build

FROM base AS backend-build
WORKDIR /app
COPY --from=backend-deps /app/go.mod ./go.mod
COPY --from=backend-deps /app/go.sum ./go.sum
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM base AS frontend-runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=frontend-build /app/.next/standalone ./
COPY --from=frontend-build /app/.next/static ./
COPY --from=frontend-build /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]

FROM scratch AS backend-runner
WORKDIR /app
COPY --from=backend-build /server .
COPY backend/configs/.env.production ./
EXPOSE 8080
CMD ["/server"]
