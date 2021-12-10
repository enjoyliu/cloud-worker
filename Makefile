.PHONY: all
all: clean

.PHONY: gen
gen:
	protoc -I/usr/local/include -I. -Ithird_party -Ithird_party/googleapis \
		--gogo_out=plugins=grpc:. proto/*.proto

.PHONY: build
build:
	@echo "VERSION: $(VERSION), BUILD_TIME: $(BUILD_TIME), GIT_COMMIT_ID: $(GIT_COMMIT_ID)"
	@for target in $(TARGETS); do                                                      \
	  GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/$${target}                  			   \
	    -a -ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommitID=$(GIT_COMMIT_ID)'" \
	    $(CMD_DIR)/$${target};                                                         \
	done

.PHONY: clean
clean:
	if [ -d $(OUTPUT_DIR) ]; then rm -rf $(OUTPUT_DIR) ; fi