export DATA="/mnt/d/Program Files/Steam/steamapps/common/Helldivers 2/data"
export CGO_ENABLE=1
export GOOSE_DBSTRING="build_15314"
export GOOSE_MIGRATION_DIR="sql/schema"
export GOOSE_DRIVER="sqlite3"

config() {
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest

    wget "https://github.com/bnnm/wwiser/releases/download/v20241210/wwiser.pyz"
    wget "https://github.com/bnnm/wwiser/releases/download/v20241210/wwnames.db3"

    go get .

    sqlc generate
}

generate() {
    if [ -e $GOOSE_DBSTRING ]; then
        mv $GOOSE_DBSTRING "${GOOSE_DBSTRING}_backup"
    fi
    goose up

    if [ -e log.txt ]; then
        rm log.txt
    fi

    go run . --insert_archive
    go run . --generate

    sqlite3 $GOOSE_DBSTRING < sql/view.sql 
}

extract_all_soundbank() {
    go run . --extract_all_soundbank --dest $1
}

extract_soundbank() {
    go run . --extract_soundbank --dest $1
}

extract_soundbank_db() {
    go run . --extract_soundbank_db --dest $1
}
