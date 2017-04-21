BUILD_PATH=$(PWD)
OUTPUT_PATH=$(PWD)/bin
SERVER_BIN=serdis-server
AGENT_BIN=serdis-agent


build-server:
	cd $(BUILD_PATH)/server && go build -o $(OUTPUT_PATH)/$(SERVER_BIN) server.go controller.go

build-agent:
	cd $(BUILD_PATH)/agent && go build -o $(OUTPUT_PATH)/$(AGENT_BIN) agent.go
