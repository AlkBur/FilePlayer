"use strict";
const maxInterval = 1000 * 60 * 5;
const minInterval = 1000;
class Socket {
    constructor() {
        this.url = '/ws';
        this.isConnected = false;
        this.setFun = {};
        this.interval = minInterval;
    }
    createWs() {
        let host = window.document.location.host;
        let isHttps = window.document.location.protocol === 'https:';
        let match = host.match(/^(.+):(\d+)$/);
        let defaultPort = isHttps ? 443 : 80;
        let port = match ? parseInt(match[2], 10) : defaultPort;
        let hostName = match ? match[1] : host;
        let wsProto = isHttps ? "wss:" : "ws:";
        let wsUrl = wsProto + '//' + hostName + ':' + port + this.url;
        this.ws = new WebSocket(wsUrl);
        this.ws.addEventListener('message', this.onMessage.bind(this), false);
        this.ws.addEventListener('error', this.onError.bind(this), false);
        this.ws.addEventListener('close', this.timeoutThenCreateNew.bind(this), false);
        this.ws.addEventListener('open', this.onOpen.bind(this), false);
    }
    timeoutThenCreateNew(event) {
        if (event.wasClean) {
            console.log('Соединение закрыто чисто');
        }
        else {
            console.log('Обрыв соединения');
        }
        console.log('Код: ' + event.code + ' причина: ' + event.reason);
        this.ws.removeEventListener('message', this.onMessage.bind(this), false);
        this.ws.removeEventListener('error', this.onError.bind(this), false);
        this.ws.removeEventListener('close', this.timeoutThenCreateNew.bind(this), false);
        this.ws.removeEventListener('open', this.onOpen.bind(this), false);
        if (this.isConnected) {
            this.isConnected = false;
            this.emit('disconnect', null);
        }
        if (!this.isConnected) {
            setTimeout(this.createWs.bind(this), this.interval);
            if (this.interval < maxInterval) {
                this.interval += minInterval;
            }
        }
    }
    onMessage(ev) {
        let msg = JSON.parse(ev.data);
        this.emit(msg.name, msg.args);
    }
    onOpen() {
        this.isConnected = true;
        this.emit('connect', null);
        this.interval = minInterval;
    }
    onError(error) {
        console.log("Ошибка " + error.message);
        this.emit('error', null);
    }
    send(name, args) {
        console.log("Send: " + name + "; args: " + args);
        this.ws.send(JSON.stringify({
            name: name,
            args: args,
        }));
    }
    emit(name, args) {
        console.log("emit Socket: " + name);
        let fnArr = this.setFun[name];
        if (fnArr) {
            for (let fn of fnArr) {
                fn(args);
            }
        }
    }
    on(name, func) {
        if (!this.setFun[name]) {
            this.setFun[name] = new Array();
        }
        this.setFun[name].push(func);
    }
}
