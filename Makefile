build:
	go build -o sofi main.go

deploy:
	cp sofi /usr/local/bin