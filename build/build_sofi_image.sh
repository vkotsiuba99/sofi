set GOARCH=amd64
set GOOS=linux
go build -o sofi ../main.go

docker build --no-cache -t sofi:v0.0.1 .
docker tag sofi:v0.0.1
docker login
docker push sofi