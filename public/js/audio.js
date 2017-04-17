"use strict";
class AudioAnalaser {
    constructor(audio) {
        this.audio = audio;
        let self = this;
        this.disable = false;
        this.capHeight = 2,
            this.capStyle = "#fff",
            this.capYPositionArray = [];
        this.context = new AudioContext();
        this.node = this.context.createScriptProcessor(2048, 1, 1);
        this.analyser = this.context.createAnalyser();
        this.analyser.smoothingTimeConstant = 0.3;
        this.analyser.fftSize = 512;
        this.frequencyData = new Uint8Array(this.analyser.frequencyBinCount);
        this.audio.addEventListener("canplay", function () {
            if (!self.source) {
                self.source = self.context.createMediaElementSource(self.audio);
                self.source.connect(self.analyser);
                self.analyser.connect(self.node);
                self.node.connect(self.context.destination);
                self.source.connect(self.context.destination);
                self.node.onaudioprocess = function () {
                    self.analyser.getByteFrequencyData(self.frequencyData);
                    if (!self.audio.paused) {
                        self.update();
                    }
                };
            }
        });
    }
    init(contener) {
        this.ctx = this.createCanvas(contener);
        this.canvas = this.ctx.canvas;
    }
    createCanvas(contener) {
        let context;
        context = contener.getContext('2d');
        return context;
    }
    update() {
        if (!this.disable)
            this.drawEqualizer();
    }
    drawEqualizer() {
        let array = this.frequencyData, ctx = this.ctx, x = 0, h = 0, cwidth = this.canvas.width, cheight = this.canvas.height;
        let barWidth = 6;
        let spaceWidth = 6;
        let numBars = Math.round(cwidth / (barWidth + spaceWidth));
        barWidth = (cwidth - spaceWidth * (numBars - 1)) / numBars;
        let step = Math.round(array.length / numBars);
        if (this.capYPositionArray.length < numBars) {
            for (let i = this.capYPositionArray.length; i < numBars; i++) {
                this.capYPositionArray.push(cheight);
            }
        }
        ctx.clearRect(0, 0, cwidth, cheight);
        let magnitude = 0;
        for (let i = 0; i < numBars; i++) {
            magnitude += array[i * step];
        }
        magnitude = magnitude / numBars;
        let scaling = cheight / 2 / magnitude;
        for (let i = 0; i < numBars; i++) {
            let value = array[i * step];
            x = i * (barWidth + spaceWidth);
            h = value * scaling;
            if (h > cheight) {
                h = cheight;
            }
            ctx.fillStyle = this.capStyle;
            let hWhiteBlock = cheight - h;
            if (value == 0) {
                hWhiteBlock -= this.capHeight * 2;
            }
            if (hWhiteBlock < this.capYPositionArray[i]) {
                this.capYPositionArray[i] = hWhiteBlock;
            }
            else {
                hWhiteBlock = this.capYPositionArray[i];
                this.capYPositionArray[i]++;
            }
            ctx.fillRect(x, hWhiteBlock, barWidth, this.capHeight);
            ctx.fillStyle = "hsl( " + Math.round(i * 360 / numBars) + ", 100%, 50% )";
            ctx.fillRect(x, cheight - h + this.capHeight, barWidth, h - this.capHeight * 2);
        }
    }
}
