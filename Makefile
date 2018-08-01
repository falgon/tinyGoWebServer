GO=go
DST=dst
OUTPUT=appserver
all: build

build:
	@mkdir -p $(DST)
	$(GO) build -o $(OUTPUT) src/index.go
	@mv $(OUTPUT) $(DST)
	@echo "Bins are in $(DST) :)"

serve:
	sudo ./
clean:
	@$(RM) -rf dst

get:
	$(GO) get github.com/gorilla/mux
