.PHONY: build vet test test-cover cover-html fmt

# Compila el binario en ./bin.
build:
	CGO_ENABLED=0 go build -o bin/go-api ./cmd

# Análisis estático.
vet:
	go vet ./...

# Escaneo de vulnerabilidades (govulncheck) con una toolchain parcheada.
# Requiere: go install golang.org/x/vuln/cmd/govulncheck@latest
vuln:
	GOTOOLCHAIN=go1.26.6 govulncheck ./...

# Ejecuta todos los tests.
test:
	go test ./...

# Ejecuta los tests y muestra la cobertura agregada de los paquetes internos.
# La cobertura se mide sobre ./internal/... (cmd/main.go queda fuera por ser
# solo el punto de entrada de wiring, como es habitual).
test-cover:
	go test -coverpkg ./internal/... -coverprofile coverage.out ./...
	go tool cover -func coverage.out

# Genera el reporte HTML de cobertura en ./coverage.html.
cover-html:
	go test -coverpkg ./internal/... -coverprofile coverage.out ./...
	go tool cover -html coverage.out

# Formatea el código.
fmt:
	go fmt ./...
