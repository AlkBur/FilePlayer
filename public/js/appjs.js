"use strict";
let player;
window.onload = () => {
    if (!player) {
        player = new Player();
    }
};
class Player {
    constructor() {
        let self = this;
        let host = window.document.location.host;
        let isHttps = window.document.location.protocol === 'https:';
        let match = host.match(/^(.+):(\d+)$/);
        let defaultPort = isHttps ? 443 : 80;
        let port = match ? parseInt(match[2], 10) : defaultPort;
        let hostName = match ? match[1] : host;
        let httpProto = isHttps ? "https:" : "http:";
        this.url = httpProto + '//' + hostName + ':' + port + '/mp3/';
        this.audioElement = document.querySelector("#AudioPlayer");
        this.btnPlayElement = document.querySelector("#play-button");
        this.btnNextPlayElement = document.querySelector("#next-button");
        this.btnPreviousPlayElement = document.querySelector("#previous-button");
        this.btnStopElement = document.querySelector("#stop-button");
        this.btnPauseElement = document.querySelector("#pause-button");
        this.listPlaylistElement = document.querySelector(".playlist-list");
        this.listFilestElement = document.querySelector("#pl-files");
        this.canvasElement = document.querySelector("#visualizer");
        this.progressBarElement = document.querySelector("#progress-bar");
        this.progressElement = document.querySelector("#progress");
        this.tooltipElement = document.querySelector("#tooltip");
        this.trackElement = document.querySelector(".track");
        this.artistElement = document.querySelector(".artist");
        this.albumElement = document.querySelector(".album");
        this.volumeImputElement = document.querySelector(".cuteslider input");
        this.volumeImputElement.value = self.audioElement.volume.toString();
        this.progressBarElement.addEventListener("mousedown", function (e) {
            if (self.status == 1 || self.status == 2) {
                let x = e.pageX;
                let duration = (x * self.audioElement.duration) / document.body.clientWidth;
                self.audioElement.currentTime = duration;
            }
        });
        this.progressBarElement.addEventListener("mouseover", function (e) {
            let x = e.pageX;
            let duration = (x * self.audioElement.duration) / document.body.clientWidth;
            this.title = self.getTooltip(duration);
        });
        this.volumeImputElement.addEventListener("change", function (e) {
            let volume = Number(this.value);
            self.audioElement.volume = volume;
        });
        this.socket = new Socket();
        this.userSettings = { currentPlaylist: "", currentFile: "", currentFileID: -1 };
        this.PlayLists = {};
        this.btnPlayElement.addEventListener("click", function () {
            if (self.status == 2) {
                self.audioElement.play();
                return;
            }
            let pl = self.PlayLists[self.userSettings.currentPlaylist];
            if (!pl || pl.files.length == 0) {
                return;
            }
            if (self.userSettings.currentFileID < 0 || self.userSettings.currentFileID >= pl.files.length) {
                self.userSettings.currentFileID = 0;
            }
            self.playFile(self.userSettings.currentFileID);
        });
        this.btnStopElement.addEventListener("click", function () {
            self.audioElement.pause();
            self.audioElement.currentTime = 0;
        });
        this.btnPauseElement.addEventListener("click", function () {
            self.audioElement.pause();
        });
        this.btnNextPlayElement.addEventListener("click", function () {
            self.playNextFile();
        });
        this.btnPreviousPlayElement.addEventListener("click", function () {
            self.playPrevFile();
        });
        this.audioElement.addEventListener("play", function () {
            self.status = 1;
            self.btnPlayElement.classList.remove("paused");
        });
        this.audioElement.addEventListener("pause", function () {
            self.status = 2;
            self.btnPlayElement.classList.add("paused");
        });
        this.audioElement.addEventListener("ended", function () {
            self.status = 0;
            self.playNextFile();
        });
        this.audioElement.addEventListener("timeupdate", function () {
            let width = (self.audioElement.currentTime * document.body.clientWidth) / self.audioElement.duration;
            self.progressElement.style.width = width + "px";
            self.tooltipElement.innerHTML = self.getTooltip(this.currentTime);
        });
        this.socket.on('connect', function () {
            self.socket.send('get_settings', "User");
        }.bind(this));
        this.socket.on('settings', function (args) {
            let argsJSON = JSON.parse(args);
            self.userSettings.currentPlaylist = argsJSON.CurPlayList;
            self.userSettings.currentFile = argsJSON.File;
            self.userSettings.currentFileID = 0;
            for (let key in argsJSON.PlayLists) {
                let pl = argsJSON.PlayLists[key];
                if (!self.PlayLists[pl.ID]) {
                    self.PlayLists[pl.ID] = { name: pl.Name, files: [] };
                    self.addPlayListElem(pl.ID, argsJSON.PlayLists[pl.ID].Files.length);
                }
                if (key == self.userSettings.currentPlaylist) {
                    self.socket.send('get_playlist', key);
                }
            }
        }.bind(this));
        this.socket.on('playlist', function (args) {
            let argsJSON = JSON.parse(args);
            if (self.PlayLists[argsJSON.id].files.length == 0) {
                for (let i = 0; i < argsJSON.data.length; i++) {
                    self.PlayLists[argsJSON.id].files.push(argsJSON.data[i]);
                    self.addFilesRow(i, argsJSON.data[i]);
                }
            }
            else {
                console.log(self.PlayLists[argsJSON.id].files);
            }
            if (argsJSON.id = self.userSettings.currentPlaylist) {
                self.userSettings.currentFileID = self.getIDForUUID(self.userSettings.currentFile);
            }
        }.bind(this));
        this.socket.createWs();
        this.audio = new AudioAnalaser(this.audioElement);
        this.audio.init(this.canvasElement);
    }
    getTooltip(time) {
        let hours = Math.floor(time / 3600);
        time = time - hours * 3600;
        let minutes = Math.floor(time / 60);
        let seconds = Math.floor(time - minutes * 60);
        let data = { hours: hours, minutes: minutes, seconds: seconds };
        let display = "";
        if (data.hours !== 0) {
            display = data.hours.toString();
        }
        if (data.minutes < 10) {
            display += "0" + data.minutes.toString();
        }
        else {
            display += data.minutes.toString();
        }
        display += ":";
        if (data.seconds < 10) {
            display += "0" + data.seconds.toString();
        }
        else {
            display += data.seconds.toString();
        }
        return display;
    }
    playFile(id) {
        let pl = this.PlayLists[this.userSettings.currentPlaylist];
        if (!pl || pl.files.length == 0 || id >= pl.files.length) {
            return;
        }
        let table = this.listFilestElement;
        let curPlay = table.rows[this.userSettings.currentFileID];
        curPlay.classList.remove("active");
        curPlay = table.rows[id];
        curPlay.classList.add("active");
        if (this.userSettings.currentFileID != id) {
            this.userSettings.currentFileID = id;
            this.userSettings.currentFile = this.getUUIDForID(this.userSettings.currentFileID);
            this.socket.send('currentFileID', this.userSettings.currentFile);
        }
        let music = this.url + pl.files[id].ID;
        this.audioElement.src = music;
        this.audioElement.load();
        this.audioElement.play();
        this.albumElement.innerText = pl.files[id].Album;
        this.artistElement.innerText = pl.files[id].Artist;
        this.trackElement.innerText = pl.files[id].Title;
    }
    playNextFile() {
        let id = this.userSettings.currentFileID;
        id++;
        let pl = this.PlayLists[this.userSettings.currentPlaylist];
        if (!pl || pl.files.length == 0 || id >= pl.files.length || id < 0) {
            id = 0;
        }
        this.playFile(id);
    }
    playPrevFile() {
        let id = this.userSettings.currentFileID;
        id--;
        let pl = this.PlayLists[this.userSettings.currentPlaylist];
        if (!pl || pl.files.length == 0 || id >= pl.files.length || id < 0) {
            id = pl.files.length - 1;
        }
        this.playFile(id);
    }
    addPlayListElem(id, num) {
        let div = document.createElement('div');
        div.className = "playlist-item";
        div.dataset['id'] = id;
        this.listPlaylistElement.appendChild(div);
        let a = document.createElement('a');
        a.innerHTML = this.PlayLists[id].name;
        div.appendChild(a);
        let span = document.createElement('span');
        span.innerHTML = num.toString();
        span.id = "total-files-" + id;
        div.appendChild(span);
    }
    addFilesRow(i, data) {
        let self = this;
        let row = this.listFilestElement.insertRow(i);
        row.addEventListener('click', function (e) {
            let uid = this.dataset["id"];
            if (uid) {
                let id = self.getIDForUUID(uid);
                self.playFile(id);
            }
        });
        let cell1 = row.insertCell(0);
        cell1.className = "checkbox_file";
        let element1 = document.createElement("input");
        element1.type = "checkbox";
        cell1.appendChild(element1);
        cell1.addEventListener('click', function (e) {
            e.stopImmediatePropagation();
        }, false);
        row.dataset.id = data.ID;
        let cell2 = row.insertCell(1);
        cell2.innerHTML = (i + 1).toString();
        cell2.className = "num_file";
        let cell3 = row.insertCell(2);
        let element2 = document.createElement("div");
        element2.innerHTML = data.Artist + " - " + data.Title + " (" + data.Year + ")";
        cell3.appendChild(element2);
        let cell4 = row.insertCell(3);
        cell4.className = "time-file";
        let element3 = document.createElement("div");
        element3.innerHTML = '0:00';
        cell4.appendChild(element3);
    }
    getUUIDForID(id) {
        let pl = this.PlayLists[this.userSettings.currentPlaylist];
        if (!pl || pl.files.length == 0 || id >= pl.files.length) {
            return "";
        }
        return pl.files[id].ID;
    }
    getIDForUUID(uuid) {
        let pl = this.PlayLists[this.userSettings.currentPlaylist];
        if (!pl || pl.files.length == 0) {
            return 0;
        }
        for (let i = 0; i < pl.files.length; i++) {
            let f = pl.files[i];
            if (f.ID == uuid) {
                return i;
            }
        }
        return 0;
    }
}
