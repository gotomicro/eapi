APP_NAME:=ego-gen-api
APP_PATH:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SCRIPT_PATH:=$(APP_PATH)/scripts
COMPILE_OUT:=$(APP_PATH)/bin/$(APP_NAME)

run:export EGO_DEBUG=true
run:
	@cd $(APP_PATH) && egoctl run

build:
	@echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>making build app<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
	@chmod +x $(SCRIPT_PATH)/build/*.sh
	@cd $(APP_PATH) && $(SCRIPT_PATH)/build/gobuild.sh $(APP_NAME) $(COMPILE_OUT)

runtmpls:
	@go run main.go gen -p testdata/bff -d "github.com/gin-gonic/gin" -t testdata/tmpls/ --resFuncs "JSONOK,JSONListPage"
