docker-up:
	docker compose -f docker-compose.yaml up -d

docker-build:
	docker compose -f docker-compose.yaml build

docker-logs:
	docker compose -f docker-compose.yaml logs -f

docker-down:
	docker compose -f docker-compose.yaml down || true

mock:
	mockgen -destination internal/repository/mock/query_mock.go github.com/DanRulev/vocabot.git/internal/repository QueryI
	mockgen -destination internal/service/mock/repository_mock.go github.com/DanRulev/vocabot.git/internal/service RepositoryI
	mockgen -destination internal/service/mock/clients_mock.go github.com/DanRulev/vocabot.git/internal/service APII
	mockgen -destination internal/bot/mock/service_mock.go github.com/DanRulev/vocabot.git/internal/bot ServiceI

test:
	go test -v -cover  ./...

run:
	go run ./cmd/main.go

.PHONY: docker-up docker-build docker-logs docker-down mock test run