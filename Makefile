clean:
	rm pb -rfv

gen:
	mkdir pb
	protoc -I proto/ --go_out=paths=source_relative:pb proto/*.proto

build:
	go build -o pproto

run:
	go build -o pproto
	./pproto
