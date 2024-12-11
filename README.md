## Introduction

- This repo contains a database that contains information about audio archive, 
audio sources, voice line transcription and other game asset information for 
Helldiver 2. This is for users that both making audio mod and wanting to 
standardize label for different audio assets. It's under going continuous update.
- This repo contains a list of human-readable labeling for different audio 
assets in Helldiver 2. It's under going continuous update.

## About Labeling

- If you want to know what audio sources that are completely independent on 
their own (aka. safe to patch), go to 
![here](https://github.com/Dekr0/hd2_audio_db/tree/main/label) to look at them.

### Contributing

- If you want to contribute to audio source labeling in any sort of form (correction, 
adding new labels for an existed label file, or adding new label file for a new 
audio archive), please open an issue / pull request with all the labels you have 
written down (either in plain text or in a file) as well as your prefer online name 
for crediting. If you want to contribute to audio source labeling for long terms, 
I will grant you access to this repository but please be responsible and do not 
do anything dangerous.
- When writing labels, please following the general format all files use in 
![here](https://github.com/Dekr0/hd2_audio_db/tree/main/label). For labeling audio 
sources relative to weapons and stratagems, please following the naming convention 
commonly used in about gun soudn design (https://youtu.be/_J56n496u6k?si=10nkfjoSItHGmTGE).

## About Database

- If you want to directly interact the database with SQL, or other external 
programming language. The data are stored in the SQLite database. If you want to 
perform complex querying in the database, you can download 
![sqlite3](https://www.sqlite.org/download.html) CLI tool or this sqlite3 
![SQLite3 GUI tool](https://sqlitebrowser.org/) to do so.
- The schema definition is located in 
![here](https://github.com/Dekr0/hd2_audio_db/tree/main/sql/schema).

### Understanding the Database schema

- `game_archive` table contain all records of all existence Helldivers 2 game 
archive in the game data folder.
- `soundbank` table contain all records of all Wwise Soundbanks in all Helldivers
 2 game archives. Here are something you need to keep in mind:
    - A single game archive can contain one or more than one Wwise Soundbanks. 
    - A Wwise Soundbank can appear only in one game archive, or in multiple game 
    archives. For example, HMG.
- `hirearchy_object` table contain all hierarchy objects of all Wwise Soundbanks 
in all Helldivers 2 game archives.
- `sound` table contain all sound objects of all Wwise Soundbanks in all 
Helldivers 2 game archives. Here's something you need to keep in mind:
    - There are two different types of ID. One is `wwise_short_id`. This type of 
    ID is the ID you will see in the UI of audio modding tool. Another one is 
    `wwise_object_id`. This type of ID is not visible in the UI of audio modding 
    tool. They're only visible when you explore Wwiser XML file of a Soundbank. 
    `wwise_short_id` is a one of the properties inside of a Sound object.
    - Two sound objects with the same `wwise_object_id` does not mean they will 
    have the `wwise_short_id`. `wwise_object_id` seems to be a type of ID that 
    keep track of an object in the hierarchy level. In contrast, 
    `wwise_short_id` is a type of ID that more focus on keeping track of a raw 
    audio source, since sound object is a container for a raw audio source.
    - Despite a sound object is in a given hierarchy of a Soundbank, this doesn't 
    means it will in the media header section of that Soundbank. And, it most 
    likely won't show up in the audio modding tool's UI.
