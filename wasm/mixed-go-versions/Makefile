all: v1_21_3 server/server run

v1_21_3: 
	cd 1.21.3 && make
	# npx webpack

server/server: server/*.go
	cd server && go build

run:
	server/server
