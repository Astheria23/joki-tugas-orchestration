.PHONY: help dev-web build-orchestrator compose-up compose-down

help:
	@echo "Joki Tugas Orchestrator Monorepo"
	@echo ""
	@echo "  make dev-web              - Run React frontend development server"
	@echo "  make build-orchestrator   - Build Orchestrator service binary"
	@echo "  make compose-up           - Start infrastructure (MongoDB) + services + React web"
	@echo "  make compose-down         - Stop compose stack"

dev-web:
	cd web && npm run dev

build-orchestrator:
	CGO_ENABLED=0 go build -o ./bin/orchestrator ./service/orchestrator/cmd/main.go

compose-up:
	docker compose -f deploy/compose/docker-compose.yml up -d

compose-down:
	docker compose -f deploy/compose/docker-compose.yml down
