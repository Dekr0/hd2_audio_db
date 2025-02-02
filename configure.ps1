$ENV:HELLDIVER2_DATA="D:/Program Files/Steam/steamapps/common/Helldivers 2/data"
$ENV:CGO_ENABLE=1
$ENV:GOBIN="C:/Users/Dekr0/go/bin"
$Env:GOOSE_DBSTRING="database"
$Env:GOOSE_MIGRATION_DIR="sql/schema"
$Env:GOOSE_DRIVER="sqlite3"
$dev = $true

function Setup() {
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest
    wget -Uri "https://github.com/bnnm/wwiser/releases/download/v20241210/wwiser.pyz" -OutFile wwiser.pyz
    wget -Uri "https://github.com/bnnm/wwiser/releases/download/v20241210/wwnames.db3" wwnames.db3
    go get .
    sqlc generate
}

function ExtractBank {
    param (
        $ArchiveId
    )
    if ($dev) {
        go run . -extract-bank="$ArchiveId" -data=$Env:HELLDIVER2_DATA
    }
}

function ParseBankXML {
    param (
        $xml
    )
    if ($dev) {
        go run . -xmls="$xml"
    }
}

function CleanXML {
    Remove-Item './*.xml'
}

function CleanBnk {
    Remove-Item './*.bnk'
}

function CleanError {
    Remove-Item './*.error'
}

function CleanAll {
    CleanXML
    CleanBnk
    CleanError
}

function GenDatabase {
    if (Test-Path -Path database) {
        rm database
        goose up
    }
    if (Test-Path -Path log.txt) {
        rm log.txt
    }
    if ($dev) {
        go run . -table-archive-all >> log.txt
        go run . -table-bank >> log.txt
        go run . -table-hierarchy-object >> log.txt
    }
    CleanAll
    CreateView
}

function CreateView {
    Get-Content ./sql/view.sql | sqlite3.exe database
}

function UpdateLabelFolder {
    param (
        $json_file_folder
    )
    if ($dev) {
        go run . -import-label-folder="$json_file_folder"
    }
}
