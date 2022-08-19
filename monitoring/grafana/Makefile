GRAFANA_ADDR ?= localhost:3000
GRAFANA_USERNAME ?= admin
GRAFANA_PASSWORD ?= prom-operator

RISINGWAVE_OVERVIEW_DASHBOARD = risingwave-overview.json

generate:
	cd jsonnet && make
	cp jsonnet/target/*.json .

update: generate
	@echo "updating the Grafana with $(RISINGWAVE_OVERVIEW_DASHBOARD)..."
	@curl -X POST -H 'Content-Type: application/json' \
	-d '{"dashboard": $(shell jq -c . risingwave-overview.json), "overwrite": true}'  \
	"http://$(GRAFANA_USERNAME):$(GRAFANA_PASSWORD)@$(GRAFANA_ADDR)/api/dashboards/db"