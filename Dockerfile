# --- Build stage ---
    FROM golang:1.22 AS build
    WORKDIR /app
    
    # Cache de dependencias
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copia del código
    COPY . .
    
    # Build estático
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/admira ./cmd/server
    
    # --- Runtime stage ---
    FROM gcr.io/distroless/base-debian12:nonroot
    ENV PORT=8080
    EXPOSE 8080
    USER nonroot:nonroot
    COPY --from=build /bin/admira /admira
    ENTRYPOINT ["/admira"]
    