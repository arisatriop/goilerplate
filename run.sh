APP_ENV=local \
APP_DEBUG=true \
APP_TIMEZONE=Asia/Jakarta \
DBMS=psql \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_HOST=localhost \
DB_PORT=5432 \
DB_NAME=goilerplate \
ELASTIC_HOST=http://localhost \
ELASTIC_PORT=9200 \
REDIS_HOST=localhost \
REDIS_PORT=6379 \
REDIS_PASSWORD=secret \
CACHE_DRIVER=redis \
LOG_CHANNEL=elasticsearch \
PORT= \
nodemon --exec go run main.go --signal SIGTERM

