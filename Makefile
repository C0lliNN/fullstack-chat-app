.PHONY: install start-web start-service start-all

start-web:
	cd web && npm start

start-service:
	MONGO_URI=mongodb://localhost:27017 MONGO_DATABASE=chatapp go run cmd/main.go

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
