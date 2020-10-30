all: gen

.PHONY : gen

gen:
	protoc -I=. --gofast_out=plugins=grpc:. transporter.proto
