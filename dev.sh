testTableArchive() {
    go run . -table-archive
}

testTableBank() {
    go run . -table-bank -data="/mnt/d/Program Files/Steam/steamapps/common/Helldivers 2/data"
}

testTableAll() {
    testTableArchive
    testTableBank
}

testExtractBank() {
    go run . -extract-bank=$1
}
