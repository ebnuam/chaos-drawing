// Drawing Chaos Online — Audio Synthesizer (Web Audio API)
// Efek suara UI & musik lo-fi sintetis orisinal tanpa unduhan aset eksternal.

(function (window) {
  'use strict';

  class SoundManager {
    constructor() {
      this.ctx = null;
      this.musicGain = null;
      this.sfxGain = null;
      this.masterGain = null;

      this.musicVolume = 0.4;
      this.sfxVolume = 0.7;
      this.isMuted = false;
      this.effectsReduced = false;

      this.currentMusicTimer = null;
      this.currentMusicTrack = null;
      this.isInitialized = false;

      // Muat preferensi dari localStorage
      this.loadSettings();
    }

    loadSettings() {
      try {
        const saved = JSON.parse(localStorage.getItem('dc_audio_settings') || '{}');
        if (typeof saved.musicVolume === 'number') this.musicVolume = saved.musicVolume;
        if (typeof saved.sfxVolume === 'number') this.sfxVolume = saved.sfxVolume;
        if (typeof saved.isMuted === 'boolean') this.isMuted = saved.isMuted;
        if (typeof saved.effectsReduced === 'boolean') this.effectsReduced = saved.effectsReduced;
      } catch (e) {
        // Abaikan jika localStorage tidak tersedia
      }
    }

    saveSettings() {
      try {
        localStorage.setItem('dc_audio_settings', JSON.stringify({
          musicVolume: this.musicVolume,
          sfxVolume: this.sfxVolume,
          isMuted: this.isMuted,
          effectsReduced: this.effectsReduced
        }));
      } catch (e) {
        // Abaikan
      }
    }

    init() {
      if (this.isInitialized) return;
      const AudioCtx = window.AudioContext || window.webkitAudioContext;
      if (!AudioCtx) return;

      this.ctx = new AudioCtx();

      // Master -> Destination
      this.masterGain = this.ctx.createGain();
      this.masterGain.gain.setValueAtTime(this.isMuted ? 0 : 1, this.ctx.currentTime);
      this.masterGain.connect(this.ctx.destination);

      // SFX -> Master
      this.sfxGain = this.ctx.createGain();
      this.sfxGain.gain.setValueAtTime(this.sfxVolume, this.ctx.currentTime);
      this.sfxGain.connect(this.masterGain);

      // Music -> Master
      this.musicGain = this.ctx.createGain();
      this.musicGain.gain.setValueAtTime(this.musicVolume, this.ctx.currentTime);
      this.musicGain.connect(this.masterGain);

      this.isInitialized = true;
    }

    ensureContext() {
      if (!this.isInitialized) this.init();
      if (this.ctx && this.ctx.state === 'suspended') {
        this.ctx.resume();
      }
    }

    setMuted(muted) {
      this.isMuted = !!muted;
      if (this.masterGain && this.ctx) {
        this.masterGain.gain.setTargetAtTime(this.isMuted ? 0 : 1, this.ctx.currentTime, 0.05);
      }
      this.saveSettings();
    }

    setMusicVolume(vol) {
      this.musicVolume = Math.max(0, Math.min(1, vol));
      if (this.musicGain && this.ctx) {
        this.musicGain.gain.setTargetAtTime(this.musicVolume, this.ctx.currentTime, 0.05);
      }
      this.saveSettings();
    }

    setSfxVolume(vol) {
      this.sfxVolume = Math.max(0, Math.min(1, vol));
      if (this.sfxGain && this.ctx) {
        this.sfxGain.gain.setTargetAtTime(this.sfxVolume, this.ctx.currentTime, 0.05);
      }
      this.saveSettings();
    }

    setEffectsReduced(reduced) {
      this.effectsReduced = !!reduced;
      this.saveSettings();
    }

    // ================= SFX PRESETS =================

    playClick() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const osc = this.ctx.createOscillator();
      const gain = this.ctx.createGain();

      osc.type = 'triangle';
      osc.frequency.setValueAtTime(440, this.ctx.currentTime);
      osc.frequency.exponentialRampToValueAtTime(880, this.ctx.currentTime + 0.04);

      gain.gain.setValueAtTime(0.3, this.ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, this.ctx.currentTime + 0.05);

      osc.connect(gain);
      gain.connect(this.sfxGain);

      osc.start();
      osc.stop(this.ctx.currentTime + 0.05);
    }

    playPop() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const osc = this.ctx.createOscillator();
      const gain = this.ctx.createGain();

      osc.type = 'sine';
      osc.frequency.setValueAtTime(320, this.ctx.currentTime);
      osc.frequency.exponentialRampToValueAtTime(640, this.ctx.currentTime + 0.08);

      gain.gain.setValueAtTime(0.4, this.ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, this.ctx.currentTime + 0.09);

      osc.connect(gain);
      gain.connect(this.sfxGain);

      osc.start();
      osc.stop(this.ctx.currentTime + 0.09);
    }

    playJoin() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const notes = [392, 523.25]; // G4, C5
      notes.forEach((freq, i) => {
        const osc = this.ctx.createOscillator();
        const gain = this.ctx.createGain();
        const t = this.ctx.currentTime + i * 0.09;

        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, t);

        gain.gain.setValueAtTime(0.35, t);
        gain.gain.exponentialRampToValueAtTime(0.001, t + 0.14);

        osc.connect(gain);
        gain.connect(this.sfxGain);

        osc.start(t);
        osc.stop(t + 0.15);
      });
    }

    playLeave() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const notes = [440, 330]; // A4, E4
      notes.forEach((freq, i) => {
        const osc = this.ctx.createOscillator();
        const gain = this.ctx.createGain();
        const t = this.ctx.currentTime + i * 0.09;

        osc.type = 'triangle';
        osc.frequency.setValueAtTime(freq, t);

        gain.gain.setValueAtTime(0.3, t);
        gain.gain.exponentialRampToValueAtTime(0.001, t + 0.12);

        osc.connect(gain);
        gain.connect(this.sfxGain);

        osc.start(t);
        osc.stop(t + 0.13);
      });
    }

    playStart() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const notes = [261.63, 329.63, 392, 523.25]; // C4, E4, G4, C5 (arpeggio)
      notes.forEach((freq, i) => {
        const osc = this.ctx.createOscillator();
        const gain = this.ctx.createGain();
        const t = this.ctx.currentTime + i * 0.08;

        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, t);

        gain.gain.setValueAtTime(0.4, t);
        gain.gain.exponentialRampToValueAtTime(0.001, t + 0.2);

        osc.connect(gain);
        gain.connect(this.sfxGain);

        osc.start(t);
        osc.stop(t + 0.22);
      });
    }

    playCorrect() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const notes = [523.25, 659.25, 783.99, 1046.5]; // C5, E5, G5, C6
      notes.forEach((freq, i) => {
        const osc = this.ctx.createOscillator();
        const gain = this.ctx.createGain();
        const t = this.ctx.currentTime + i * 0.07;

        osc.type = 'triangle';
        osc.frequency.setValueAtTime(freq, t);

        gain.gain.setValueAtTime(0.45, t);
        gain.gain.exponentialRampToValueAtTime(0.001, t + 0.25);

        osc.connect(gain);
        gain.connect(this.sfxGain);

        osc.start(t);
        osc.stop(t + 0.26);
      });
    }

    playWrong() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const osc = this.ctx.createOscillator();
      const gain = this.ctx.createGain();

      osc.type = 'sawtooth';
      osc.frequency.setValueAtTime(180, this.ctx.currentTime);
      osc.frequency.linearRampToValueAtTime(110, this.ctx.currentTime + 0.2);

      gain.gain.setValueAtTime(0.25, this.ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, this.ctx.currentTime + 0.25);

      osc.connect(gain);
      gain.connect(this.sfxGain);

      osc.start();
      osc.stop(this.ctx.currentTime + 0.26);
    }

    playTick() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const osc = this.ctx.createOscillator();
      const gain = this.ctx.createGain();

      osc.type = 'sine';
      osc.frequency.setValueAtTime(800, this.ctx.currentTime);

      gain.gain.setValueAtTime(0.15, this.ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, this.ctx.currentTime + 0.03);

      osc.connect(gain);
      gain.connect(this.sfxGain);

      osc.start();
      osc.stop(this.ctx.currentTime + 0.03);
    }

    playTimeUp() {
      if (this.isMuted) return;
      this.ensureContext();
      if (!this.ctx) return;

      const notes = [440, 415, 392, 349.2];
      notes.forEach((freq, i) => {
        const osc = this.ctx.createOscillator();
        const gain = this.ctx.createGain();
        const t = this.ctx.currentTime + i * 0.1;

        osc.type = 'sawtooth';
        osc.frequency.setValueAtTime(freq, t);

        gain.gain.setValueAtTime(0.3, t);
        gain.gain.exponentialRampToValueAtTime(0.001, t + 0.18);

        osc.connect(gain);
        gain.connect(this.sfxGain);

        osc.start(t);
        osc.stop(t + 0.2);
      });
    }

    // ================= SINTESIS MUSIK LATAR =================

    startMusic(type = 'lobby') {
      this.stopMusic();
      this.currentMusicTrack = type;
      if (this.isMuted) return;

      this.ensureContext();
      if (!this.ctx) return;

      // Nada melodi santai dan menyenangkan (pentatonik C)
      const lobbyNotes = [
        261.63, 0, 329.63, 392.00, 0, 440.00, 392.00, 329.63,
        261.63, 329.63, 0, 392.00, 440.00, 0, 523.25, 0
      ];

      const playNotes = [
        329.63, 392.00, 440.00, 523.25, 440.00, 392.00, 329.63, 0,
        261.63, 293.66, 329.63, 392.00, 329.63, 293.66, 261.63, 0
      ];

      const sequence = type === 'play' ? playNotes : lobbyNotes;
      let step = 0;
      const stepDuration = 240; // ms

      const playStep = () => {
        if (!this.currentMusicTrack || this.isMuted || !this.ctx) return;

        const freq = sequence[step % sequence.length];
        step++;

        if (freq > 0) {
          const osc = this.ctx.createOscillator();
          const gain = this.ctx.createGain();
          const filter = this.ctx.createBiquadFilter();

          filter.type = 'lowpass';
          filter.frequency.setValueAtTime(800, this.ctx.currentTime);

          osc.type = 'triangle';
          osc.frequency.setValueAtTime(freq, this.ctx.currentTime);

          gain.gain.setValueAtTime(0.08, this.ctx.currentTime);
          gain.gain.exponentialRampToValueAtTime(0.0001, this.ctx.currentTime + 0.2);

          osc.connect(filter);
          filter.connect(gain);
          gain.connect(this.musicGain);

          osc.start();
          osc.stop(this.ctx.currentTime + 0.22);
        }

        this.currentMusicTimer = setTimeout(playStep, stepDuration);
      };

      playStep();
    }

    stopMusic() {
      if (this.currentMusicTimer) {
        clearTimeout(this.currentMusicTimer);
        this.currentMusicTimer = null;
      }
      this.currentMusicTrack = null;
    }
  }

  window.SoundManager = new SoundManager();
})(window);

