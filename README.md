## Table of Content
- ![Introduction](#introduction)
- ![About CLI](#about-cli)
- ![About Labelling](#about-labelling) 
- ![About Database](#about-database)

## Introduction

- This repo contains a database that contains information about audio archive, audio sources,
voice line transcription and other game asset information for Helldiver 2. This is for users
that both making audio mod and wanting to standardize label for different audio assets. It's
under going continuous update.
- This repo contains a list of human-readable labelling for different audio assets in Helldiver 2.
It's under going continuous update.
- This repo contains a CLI (Command Line Interface) to resolve common problems users encounter
when they are patching audio sources. This is for if you're primarily on making audio mod.

## About CLI

### Install

- First, get the executable of this CLI in the ![release page]()
- Make sure that `hd_audio_db.db` is in the same folder the executable is in. Otherwise, the
executale won't know where to find data about all the audio assets

### Usage

#### Generate a listing of shared and "safe" audio source for an audio archive 

- If you want to look for what audio sources in an archive are shared with different other
archives, and what audio sources in an archive are completely independent (aka. safe to patch)
on their own, run the following in the command line at the folder where your exectuable is in,
replace `target_archive_id` with actual archie id.
```
hda.exe --gen_dep_archive_id=[target_archive_id]
```
- For example, `hda.exe --gen_dep_archive_id=6ad7cc21015a5f85` will generate two csv files,
one for shared audio sources, one for "safe" audio sources.
- REMARK! The asset names in the csv file that contains shared audio sources are not 100%
correct. They are suggestions. For accurcate location, use the archive id.
- There are asset names such as Purifier and Scorcher whose archive are exactly the same, i.e.,
both asset name Purifier and Scorcher points to audio archive `c51f160470b34ae3`. Thus, you
need to use the audio archive ID and audio source ID to pin point which asset name is actually
relative to this audio source.
- REMARK! There are audio sources that are implicitly shared between audio archive. The CLI won't
be able to catch this right now unless someone label them and put it in the database.
- What does "implicitly shared" means? Take Eagle Strafing Run and Gatling Barrage as examples.
Eagle Strafing Run only has very number of audio source if you unpack it. It's actually use around
20 audio sources in the Gatling Barrage audio archive for its projectile impact SFX but those audio
sources id don't show up when you unpack Eagle Strafing Run's audio archive. 

#### Audio source overwrite check

- If you want to look for whether the audio sources you want to patch are shared with other archives,
run the folliwng in the command line at the folder where your exectuable is in, replace the
`input_file_path` with actual file path of the file you want to in your computer.
```
hda.exe --overwrite_check=[input_file_path]
```
- The input file must follow the following format. Otherwise, it will fail.
- First line must be the archive id you're patching.
- The rest of the line must be a valid audio source id that exists in the archive you're patching.
    - An empty line is allowed.
- Input file example (For Autocannon)
```
6ad7cc21015a5f85
624808124
262499303
64932800
641902294
1037720523
567398501
640606470
399599120
858271035
789524327
713403673
891973340
```
- It will generate a file that contains a list of affected audio sources and the audio archive ID they're
in. Example output is given in the following
```
{
    "shared_audio_sources": [
        {
            "audio_source_id": "624808124",
            "affecting_audio_archive_ids": [
                "2a9fb8f5d0576eae",
                "6ca87637eaeb5923",
                "6ad7cc21015a5f85",
                "e0807831d6f456c8",
                "7f37db9b767844c2",
                "e75f556a740e00c9",
                "c0052e5b38e18c33",
                "f7646e79610c124d",
                "a37891d879cb3b1d",
                "530bb611a14b6ce3",
                "cf1acde501ccfa1b",
                "ea6967f8565a2d76"
            ],
            "affecting_audio_archive_names": [
                "RL-77 Airburst Rocket Launcher",
                "AC-8 Autocannon",
                "FLAM-40 Flamethrower",
                "GR-8 Recoilless Rifle",
                "LAS-98 Laser Cannon",
                "MLS-4X Commando",
                "APW-1 Anti-materiel Rifle",
                "FAF-14 Spear",
                "R-36 Eruptor",
                "ARC-3 Arc Thrower",
                "content/audio/destruction",
                "RS-422 Railgun"
            ]
        },
        {
            "audio_source_id": "262499303",
            "affecting_audio_archive_ids": [
                "2a9fb8f5d0576eae",
                "6ca87637eaeb5923",
                "6ad7cc21015a5f85",
                "e0807831d6f456c8",
                "7f37db9b767844c2",
                "e75f556a740e00c9",
                "c0052e5b38e18c33",
                "f7646e79610c124d",
                "a37891d879cb3b1d",
                "530bb611a14b6ce3",
                "cf1acde501ccfa1b",
                "ea6967f8565a2d76"
            ],
```

## About Labelling

- If you want to know what audio sources that are completely independent on their own (aka.
safe to patch), go to ![here](https://github.com/Dekr0/hd2_audio_db/tree/main/label) to look at them.

## About Database

- If you want to directly interact the database with SQL, or other external programming language.
The data are stored in the SQLite database. If you want to perform complex querying in the database,
you can download ![sqlite3](https://www.sqlite.org/download.html) CLI tool or this sqlite3 ![SQLite3 GUI tool](https://sqlitebrowser.org/) to do so.
- The schema definition is located in ![here](https://github.com/Dekr0/hd2_audio_db/tree/main/sql/schema).
