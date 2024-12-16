$ENV:HELLDIVER2_DATA="D:\Program Files\Steam\steamapps\common\Helldivers 2\data"
$ENV:CGO_ENABLE=1
$ENV:GOBIN="C:\Users\Dekr0\go\bin"
function ExtractBankDev {
    param (
        $ArchiveId
    )
    go run . -extract-bank="$ArchiveId" -data=$Env:HELLDIVER2_DATA
}

function ParseBankXMLDev {
    param (
        $xml
    )

    go run . -xmls="$xml"
}

function CleanXML {
    Remove-Item '.\*.xml'
}

function TestTableArchive {
    rm log.txt
    go run . -table-archive >> log.txt
}

function TestTableArchiveAll {
    rm log.txt
    go run . -table-archive-all >> log.txt
}

function TestTableBankAll {
    rm log.txt
    go run . -table-bank >> log.txt
}

function TestTableSoundAssets {
    rm log.txt
    go run . -table-sound-asset >> log.txt
    CleanXML
}

function CreateSoundView {
    Get-Content ./sql/view.sql | sqlite3.exe database
}

function UpdateLabel {
    param (
        $json_files
    )

    go run . -table-label="$json_files"
}

function DryRun {
    UpdateLabel -json_files "label/weapons_stratagems/safe/120mm.json;label/weapons_stratagems/safe/380mm.json;label/weapons_stratagems/safe/500kg.json"
}
