create_db: ## Creates db for this project.
	docker run --name=db -e POSTGRES_PASSWORD='123456' -p 5432:5432 -d --rm postgres
	sleep 2
	docker exec -it db createdb -U postgres kitchennerd_db

run: ## Runs the project.
	go run ./cmd/main.go
