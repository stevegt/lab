all: v1_19_1 v1_21_3 server/server run


v1_19_1: 
	cd 1.19.1 && make

v1_21_3: 
	cd 1.21.3 && make

server/server: server/*.go
	cd server && go build

run:
	server/server
