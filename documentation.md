# Update Archive Table

- For each `csv` file, first column always starts with a tag that is assigned to 
an array of game archive IDs.
- For each `csv` file, the name of that `csv` file is the category that is belong 
to all game archive IDs shown in that `csv` file.
- Since an game archive can contain multiple Wwise Soundbank (e.g. e75f556a740e00c9),
, a game archive can have more than one tag, and can have more than one category.

# Update Wwise Soundbank and Wwise Hierarchy Object Table

- Select all rows in the `helldiver_game_archive`, and obtain `game_archive_id` 
for each row.
- Join the path, specified by `HELLDIVER_DATA` (location of Helldivers 2 game 
directory) environmental variable, with the `game_archive_id`, to obtain the 
file path of a game archive with associated with `game_archive_id`.
- Parse ToC file of a game archive, and extract all Wwise Soundbanks and Wwise 
Dependencies.
    - There are might be some game archives that are no longer in the Helldivers 
    2 game data directory but still listed in the google spreadsheet.
- For each Wwise Soundbank, write its record into the `helldiver_soundbank`.
- For each Wwise Soundbank, output its binary content, input into wwiser to 
generate a XML file.
- Parse through the XML file, and return a `CAkWwiseBank` struct that encapsulate 
all information (media index, hirearchy, objects [encapsulated by `CAkObject` 
interface], etc.) in a given Wwise Soundbank.
- For a given Wwise Soundbank, transfer its objects into the hashmap that stores 
all objects from every single Wwise Soundbank used in Helldivers 2 uniquely.
    - If an object isn't in the hashmap, create a wrapper around this object. 
    This wrapper contains a hashmap. This hashmap contains primary keys of records 
    for Wwise Sounbanks. If a Wwise Soundbank contain this object, the primary key 
    of its record will in this hashmap
    - If an object is in the hashmap, store primary key of its record.
