GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o hda.exe
mv hda.exe ./build
cp ./.env ./build
cp ./hd_audio_db.db ./build
