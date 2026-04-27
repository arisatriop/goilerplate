### To run using nodemon
# export APP_ENV=local
# nodemon -e go --signal SIGINT --exec "go run cmd/server/main.go" 
mkdir -p ./storage/logs && nodemon -e go --signal SIGINT --exec "sh -c 'go run cmd/server/main.go 2>&1 | tee ./storage/logs/app.log'"



## To run using air
# Option 1: You can run simply by typing "air" command in terminal if you have configured the .air.toml
# Option 2: Or you can run directly using the command below
# air --build.cmd "go build -o tmp/server ./cmd/server" --build.bin "tmp/server"
