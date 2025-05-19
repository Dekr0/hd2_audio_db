build_win() {
    rm -r build_win 
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o main.exe
    mkdir build_win
    mv main.exe build_win
    cp .env build_win
    cp database build_win

    cd build_win

    wget https://github.com/bnnm/wwiser/releases/download/v20240526/wwiser.pyz
    wget https://github.com/bnnm/wwiser/releases/download/v20240526/wwnames.db3 
}

build_linux() {
    rm -r build_linux
    mkdir build_linux
    go build . -o main
    mv main build/linux

    cp .env build_linux
    cp database build_linux

    cd build_linux

    wget https://github.com/bnnm/wwiser/releases/download/v20240526/wwiser.pyz
    wget https://github.com/bnnm/wwiser/releases/download/v20240526/wwnames.db3 
}
