$ENV:DATA="D:/Program Files/Steam/steamapps/common/Helldivers 2/data"
$ENV:CGO_ENABLE=1
$ENV:GOBIN="C:/Users/Dekr0/go/bin"
$Env:GOOSE_DBSTRING="build_15016"
$Env:GOOSE_MIGRATION_DIR="sql/schema"
$Env:GOOSE_DRIVER="sqlite3"

function config {
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest

    wget -Uri "https://github.com/bnnm/wwiser/releases/download/v20241210/wwiser.pyz" -OutFile wwiser.pyz
    wget -Uri "https://github.com/bnnm/wwiser/releases/download/v20241210/wwnames.db3" wwnames.db3

    go get .

    sqlc generate
}

function generate {
    if (Test-Path -Path $Env:GOOSE_DBSTRING) {
        mv $Env:GOOSE_DBSTRING $Env:GOOSE_DBSTRING_backup
    }
    goose up
    
    if (Test-Path -Path log.txt) {
        rm log.txt
    }

    go run . --insert_archive
    go run . --generate

    Get-Content ./sql/view.sql | sqlite3.exe $Env:GOOSE_DBSTRING
}

function extract_all_soundbank {
    param (
        $dest
    )
    go run . --extract_all_soundbank --dest $dest
}

function extract_soundbank {
    param (
        $dest
    )
    go run . --extract_soundbank --dest $dest
}

function extract_soundbank_db {
    param (
        $dest
    )
    go run . --extract_soundbank_db --dest $dest
}
