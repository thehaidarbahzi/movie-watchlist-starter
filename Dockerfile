# ---- build stage ----
FROM golang:1.22-alpine AS build

# go-sqlite3 needs CGO + a C compiler.
RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/filmlog .

# ---- runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=build /out/filmlog .

# SQLite is stored outside the binary so it can be mounted as a volume.
ENV PORT=9994
ENV DB_PATH=/app/data/filmlog.db
RUN mkdir -p /app/data
VOLUME ["/app/data"]

EXPOSE 9994
CMD ["./filmlog"]
