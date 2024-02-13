up:
	docker compose stop && docker compose up --build --wait --wait-timeout 60
	docker compose alpha watch
