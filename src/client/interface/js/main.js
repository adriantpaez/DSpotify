app = new Vue({
    el: "#app",
    data: {
        artists: [],
        initialState: true,
        showArtists: false,
        upArtist: false,
        currentArtist: "",
        upSong: false,
        currentSongTitle: "",
        currentSongArtist: "",
    },
    methods: {
        getSongs(artist) {
            console.log("Find songs of " + artist.name);
            // TODO: Get artist songs from client server
            console.log("Songs of " + artist.name + " founded")
        },
        loadArtists() {
            console.log("Loading Artists");
            // TODO: Get artists from client server
            this.artists = artists;
            console.log("Artists loaded");
            this.showArtists = true;
        },
        // UPLOAD ARTIST
        uploadArtist() {
            this.upArtist = true;
            console.log("Uploading artis " + this.currentArtist);
            //TODO: Upload new artist
            console.log(name + " uploaded done");
            this.upArtist = false;
            this.initialState = true;
        },
        cancelUploadArtist() {
            console.log("Uploading artist cancelled");
            this.currentArtist = "";
            this.upArtist = false;
            this.initialState = true;
        },
        // UPLOAD SONG
        uploadSong() {
            this.upSong = true;
            console.log("Uploading song " + this.currentSongTitle);
            //TODO: Upload new song
            console.log(name + " uploaded done");
            this.upSong = false;
            this.initialState = true;
        },
        cancelUploadSong() {
            console.log("Uploading song cancelled");
            this.currentSongArtist = "";
            this.currentSongTitle = "";
            this.upSong = false;
            this.initialState = true;
        },
    },
    computed: {
        sortedArtists() {
            return this.artists.sort((a, b) => {
                if (a.name < b.name) {
                    return -1;
                } else {
                    return 1;
                }
            });
        }
    }
});