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

function TestTableArchive() {
    go run . -table-archive >> log.txt
}

function TestTableBank() {
    go run . -table-bank >> log.txt
}

function TestTableAll() {
    go run . -table-archive
    go run . -table-bank
}
