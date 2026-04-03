# Windows + Go
GO=go
BIN_EXT=.exe

build: build-server build-client build-sensors build-actuators
	@echo Build completo.

build-server:
	$(GO) build -o server/server$(BIN_EXT) server/server.go

build-client:
	$(GO) build -o client/client$(BIN_EXT) client/client.go

build-sensors:
	$(GO) build -o sensor/gas/gas$(BIN_EXT) sensor/gas/gas.go
	$(GO) build -o sensor/humidity/humidity$(BIN_EXT) sensor/humidity/humidity.go
	$(GO) build -o sensor/luminosity/luminosity$(BIN_EXT) sensor/luminosity/luminosity.go
	$(GO) build -o sensor/smoke/smoke$(BIN_EXT) sensor/smoke/smoke.go
	$(GO) build -o sensor/temperature/temperature$(BIN_EXT) sensor/temperature/temperature.go

build-actuators:
	$(GO) build -o actuator/airConditioner/airConditioner$(BIN_EXT) actuator/airConditioner/airConditioner.go
	$(GO) build -o actuator/exhaustFan/exhaustFan$(BIN_EXT) actuator/exhaustFan/exhaustFan.go
	$(GO) build -o actuator/humidifier/humidifier$(BIN_EXT) actuator/humidifier/humidifier.go
	$(GO) build -o actuator/light/light$(BIN_EXT) actuator/light/light.go
	$(GO) build -o actuator/sprinkler/sprinkler$(BIN_EXT) actuator/sprinkler/sprinkler.go

clean:
	del /Q server\*.exe client\*.exe sensor\gas\*.exe sensor\humidity\*.exe sensor\luminosity\*.exe sensor\smoke\*.exe sensor\temperature\*.exe actuator\airConditioner\*.exe actuator\exhaustFan\*.exe actuator\humidifier\*.exe actuator\light\*.exe actuator\sprinkler\*.exe 2>nul || exit 0