### To run using nodemon
# export APP_ENV=local
nodemon -e go --signal SIGINT --exec "go run cmd/server/main.go" 



### To run using air
## Option 1: You can run simply by typing "air" command in terminal if you have configured the .air.toml
## Option 2: Or you can run directly using the command below
# air --build.cmd "go build -o tmp/server ./cmd/server" --build.bin "tmp/server"
