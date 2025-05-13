export HD2DATA="/mnt/Program Files/Steam/steamapps/common/Helldivers 2/data"
export CGO_ENABLE=1
export GOOSE_DBSTRING="database_build_15016"
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

extractBnk() {
    local archiveId="${1:-}"
    go run . -extract-bank=$archiveId -data=$HD2DATA
}

cleanXML() {
    rm *.xml
}

cleanBnk() {
    rm *.bnk
}

cleanErrorBnk() {
    rm *.error
}

cleanAll() {
    cleanXML
    cleanBnk
    cleanErrorBnk
}

generate() {
    if [ -e $GOOSE_DBSTRING ]; then
        rm $GOOSE_DBSTRING
    fi
    if [ -e log.txt ]; then
        rm log.txt
    fi
    go run . -table-archive-all >> log.txt
    go run . -table-bank >> log.txt
    go run . -table-hierarchy-object >> log.txt
}

createView() {
    sql/view.sql | sqlite3 $GOOSE_DBSTRING
}
