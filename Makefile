ENTRY=cmd/main.go
TARGET=aurora

COMPILER=go build
FLAGS=CGO_ENABLED=1

all: $(TARGET)

aurora:
	$(FLAGS) go build -o build/$(TARGET) $(ENTRY)

tests:
	go test ./...

clean:
	rm -f build/$(TARGET)
