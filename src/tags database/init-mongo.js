db = db.getSiblingDB('dspotify');
db.createCollection("artist", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["name"],
            properties: {
                name: {
                    bsonType: "string",
                    description: "Name of the artist"
                },
            }
        }
    }
});
db.artist.createIndex({"name": 1}, {unique: true});
db.createCollection("song", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["title", "artist", "blocks"],
            properties: {
                title: {
                    bsonType: "string",
                    description: "Song title"
                },
                artist: {
                    bsonType: "objectId",
                    description: "Artist of the song"
                },
                blocks: {
                    bsonType: "int",
                    description: "Number of block to split song"
                }
            }
        }
    }
});
db.song.createIndex({title: 1, artist: 1}, {unique: true});
