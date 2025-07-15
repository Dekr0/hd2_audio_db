export DATA="/mnt/d/Program Files/Steam/steamapps/common/Helldivers 2/data"
export CGO_ENABLE=1
export GOOSE_DRIVER="sqlite3"

VERSION=15314

config() {
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest

    wget "https://github.com/bnnm/wwiser/releases/download/v20241210/wwiser.pyz"
    wget "https://github.com/bnnm/wwiser/releases/download/v20241210/wwnames.db3"

    go get .

    sqlc generate
}

goose_up_complete() {
    export GOOSE_DBSTRING="build_${VERSION}"
    export GOOSE_MIGRATION_DIR="sql/schema/complete"

    if [ -e $GOOSE_DBSTRING ]; then
        mv $GOOSE_DBSTRING "${GOOSE_DBSTRING}_backup"
    fi
    goose up
}

goose_up_id() {
    export GOOSE_DBSTRING="id_${VERSION}"
    export GOOSE_MIGRATION_DIR="sql/schema/id"

    if [ -e $GOOSE_DBSTRING ]; then
        mv $GOOSE_DBSTRING "${GOOSE_DBSTRING}_backup"
    fi
    goose up
}

reset_log() {
    if [ -e log.txt ]; then
        rm log.txt
    fi
}

generate() {
    goose_up_complete

    reset_log

    go run . --insert_archive
    go run . --generate

    sqlite3 $GOOSE_DBSTRING < sql/view.sql 
}

export_id() {
    goose_up_id
    export BUILD="build_${VERSION}_backup"
    reset_log
    go run . --export_id
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
