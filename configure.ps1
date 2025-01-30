$ENV:HELLDIVER2_DATA="D:/Program Files/Steam/steamapps/common/Helldivers 2/data"
$ENV:CGO_ENABLE=1
$ENV:GOBIN="C:/Users/Dekr0/go/bin"
$Env:GOOSE_DBSTRING="database"
$Env:GOOSE_MIGRATION_DIR="sql/schema"
$Env:GOOSE_DRIVER="sqlite3"
$dev = $true

function ExtractBankDev {
    param (
        $ArchiveId
    )
    if ($dev) {
        go run . -extract-bank="$ArchiveId" -data=$Env:HELLDIVER2_DATA
    }
}

function ParseBankXMLDev {
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

function RewriteTableArchiveSpreadSheet {
    rm log.txt
    if ($dev) {
        go run . -table-archive >> log.txt
    }
}

function RewriteTableArchiveAll {
    rm log.txt
    if ($dev) {
        go run . -table-archive-all >> log.txt
    }
}

function RewriteTableBankAll {
    rm log.txt
    if ($dev) {
        go run . -table-bank >> log.txt
    }
}

function RewriteTableSoundAssets {
    rm log.txt
    if ($dev) {
        go run . -table-sound-asset >> log.txt
    }
    CleanXML
}

function RewriteEverything {
    if ($dev) {
        go run . -table-archive-all >> log.txt
        go run . -table-bank >> log.txt
        go run . -table-sound-asset >> log.txt
    }
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
