create-image:
	docker build build/all-in-one-ubuntu -t all-in-one-ubuntu

build:
	go build -o sofi main.go

deploy:
	cp sofi /usr/local/bin