$ENV:HELLDIVER2_DATA="D:\Program Files\Steam\steamapps\common\Helldivers 2\data"
$ENV:CGO_ENABLE=1
$ENV:GOBIN="C:\Users\Dekr0\go\bin"
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
    Remove-Item '.\*.xml'
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
        go run . -table-archive-all > log.txt
        go run . -table-sound-asset > log.txt
    }
}

function CreateView {
    Get-Content ./sql/view.sql | sqlite3.exe database
}

function UpdateLabel {
    param (
        $json_files
    )
    if ($dev) {
        go run . -table-label="$json_files"
    }
}

function UpdateLabelFolder {
    param (
        $json_file_folder
    )
    if ($dev) {
        go run . -table-label-folder="$json_file_folder"
    }
}
