.PHONY: install start-web start-service start-all

start-web:
	cd web && npm start

start-service:
	go run cmd/main.go

