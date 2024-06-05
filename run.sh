APP_ENV=local \
APP_DEBUG=true \
DBMS=psql \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_HOST=localhost \
DB_PORT=5432 \
DB_NAME=goilerplate \
PORT= \
ELASTIC_HOST=http://localhost \
ELASTIC_PORT=9200 \
nodemon --exec go run main.go --signal SIGTERM

