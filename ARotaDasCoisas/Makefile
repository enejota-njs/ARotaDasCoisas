IP=server-ards

# ==========================================
# 1. CONSTRUINDO AS IMAGENS
# ==========================================

build:
	docker compose build

build-server:
	docker compose build server

build-client:
	docker compose build client

build-gas:
	docker compose build gas

build-humidity:
	docker compose build humidity

build-luminosity:
	docker compose build luminosity

build-smoke:
	docker compose build smoke

build-temperature:
	docker compose build temperature

build-air-conditioner:
	docker compose build air_conditioner

build-exhaust-fan:
	docker compose build exhaust_fan

build-humidifier:
	docker compose build humidifier

build-light:
	docker compose build light

build-sprinkler:
	docker compose build sprinkler

# ==========================================
# 2. EXECUTANDO O ECOSSISTEMA
# ==========================================

run-server:
	docker compose up server

run-client:
	docker compose run --rm client ./client_bin $(IP)

# --- Rodar Sensores ---
run-gas:
	docker compose run --rm gas ./sensor_bin $(IP)

run-humidity:
	docker compose run --rm humidity ./sensor_bin $(IP)

run-luminosity:
	docker compose run --rm luminosity ./sensor_bin $(IP)

run-smoke:
	docker compose run --rm smoke ./sensor_bin $(IP)

run-temperature:
	docker compose run --rm temperature ./sensor_bin $(IP)

# --- Rodar Atuadores ---
run-air-conditioner:
	docker compose run --rm air_conditioner ./actuator_bin $(IP)

run-exhaust-fan:
	docker compose run --rm exhaust_fan ./actuator_bin $(IP)

run-humidifier:
	docker compose run --rm humidifier ./actuator_bin $(IP)

run-light:
	docker compose run --rm light ./actuator_bin $(IP)

run-sprinkler:
	docker compose run --rm sprinkler ./actuator_bin $(IP)