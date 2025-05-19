## Introduction

- This repo contains a database that contains information about audio archive, 
audio sources, voice line transcription and other game asset information for 
Helldiver 2. This is for users that both making audio mod and wanting to 
standardize label for different audio assets. It's under going continuous update.
- This repo contains a list of human-readable labeling for different audio 
assets in Helldiver 2. It's under going continuous update.

## Usage

### Development Setup

- This procedure should only run once.

#### Windows

- Open `./configure.ps1` with an editor, and assign the following
    - `$ENV:HELLDIVER2_DATA` with absolute path Helldivers 2 game data folder
    - `$ENV:GOBIN` with absolute path of Go bin folders (It contains all 
    executable programs installed using `go install`.)
        - Usually it's something like `C:/Users/YourUserName/go/bin`
- Execute `. ./configure.ps1` in the terminal to source all commands used in 
this go project.
- Execute `Setup` in the terminal to install all the necessary go dependencies 
for this go project.

### Database Generation 

#### Windows

- Make sure to execute `. ./configure.ps1` in the terminal first.
- Execute `GenDatabase` in the terminal to generate a brand new database.
- If there's any error happen at any point of the generation, please check 
`log.txt`.

### Other Utility Tools

- Work in Progress, or take a look of the source code and shell script.

## About Labeling

- If you want to know what audio sources that are completely independent on 
their own (aka. safe to patch), go to 
[here](https://github.com/Dekr0/hd2_audio_db/tree/redesign/label) to look at them.

### Contributing

- If you want to contribute to audio source labeling in any sort of form (correction, 
adding new labels for an existed label file, or adding new label file for a new 
audio archive), please open an issue / pull request with all the labels you have 
written down (either in plain text or in a file) as well as your prefer online name 
for crediting. If you want to contribute to audio source labeling for long terms, 
I will grant you access to this repository but please be responsible and do not 
do anything dangerous.
- When writing labels, please following the general format all files use in 
[here](https://github.com/Dekr0/hd2_audio_db/tree/redesign/label). For labeling audio 
sources relative to weapons and stratagems, please following the naming convention 
commonly used in about gun soudn design (https://youtu.be/_J56n496u6k?si=10nkfjoSItHGmTGE).

## About Database

- If you want to directly interact the database with SQL, or other external 
programming language. The data are stored in the SQLite database. If you want to 
perform complex querying in the database, you can download 
[sqlite3](https://www.sqlite.org/download.html) CLI tool or this sqlite3 
[SQLite3 GUI tool](https://sqlitebrowser.org/) to do so.
- The schema definition is located in 
[here](https://github.com/Dekr0/hd2_audio_db/tree/redesign/sql/schema).

### Understanding the Database schema

- Rework In Progress
