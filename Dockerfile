# syntax=docker/dockerfile:1
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/park-systems ./cmd/park-systems

FROM gcr.io/distroless/static-debian12:nonroot
# Dev fixtures are baked so the dev profile runs out of the box. In prod, mount
# the real flags file and set PARK_FLAGS_FILE to override (see deploy/).
COPY --from=build /out/park-systems /park-systems
COPY fixtures /fixtures
EXPOSE 8080
ENTRYPOINT ["/park-systems"]
