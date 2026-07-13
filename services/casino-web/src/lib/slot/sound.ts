"use client";

/**
 * Procedural Web Audio sound engine for a casino slot game.
 *
 * Pure browser Web Audio API + TypeScript. No external audio files, no React,
 * from oscillators and short noise buffers.
 *
 * The AudioContext is created lazily on the first sound call (browsers require
 * a user gesture) and routed through a single master GainNode so the entire
 * engine can be muted/unmuted in one place.
 */

type WebkitWindow = Window & {
  webkitAudioContext?: typeof AudioContext;
};

export class SoundEngine {
  /** Lazily-created AudioContext. Null until the first sound call. */
  private audioCtx: AudioContext | null = null;
  /** Master gain node; created together with the context. */
  private master: GainNode | null = null;
  /** Mute state. When true, master gain is driven to 0. */
  private muted = false;
  /** Target master volume when unmuted. */
  private readonly masterVolume = 0.5;

  /**
   * Lazily create and return the AudioContext, ensuring the master GainNode
   * exists and is connected to `destination`. All sound methods call this
   * first so the context is initialized on a user gesture.
   */
  private ctx(): AudioContext {
    if (this.audioCtx) return this.audioCtx;
    const w = window as WebkitWindow;
    const Ctor: typeof AudioContext | undefined =
      window.AudioContext ?? w.webkitAudioContext;
    if (!Ctor) {
      throw new Error("Web Audio API is not supported in this environment");
    }
    const ac = new Ctor();
    const master = ac.createGain();
    master.gain.value = this.muted ? 0 : this.masterVolume;
    master.connect(ac.destination);
    this.audioCtx = ac;
    this.master = master;
    return ac;
  }

  /** Current audio clock time, or 0 if the context has not been created yet. */
  private now(): number {
    return this.audioCtx ? this.audioCtx.currentTime : 0;
  }

  /** Mute or unmute the entire engine. Keeps the context running. */
  setMuted(b: boolean): void {
    this.muted = b;
    const ac = this.audioCtx;
    const master = this.master;
    if (!ac || !master) return;
    try {
      const t = this.now();
      master.gain.cancelScheduledValues(t);
      master.gain.setValueAtTime(master.gain.value, t);
      master.gain.linearRampToValueAtTime(
        b ? 0 : this.masterVolume,
        t + 0.02
      );
    } catch {
      // ignore
    }
  }

  /** Whether the engine is currently muted. */
  getMuted(): boolean {
    return this.muted;
  }

  /**
   * Play a single oscillator note with a quick attack and an exponential
   * decay envelope. Connects to `dest` (or the master gain by default).
   */
  private playNote(
    ac: AudioContext,
    freq: number,
    start: number,
    dur: number,
    type: OscillatorType,
    gain: number,
    dest?: AudioNode
  ): void {
    const osc = ac.createOscillator();
    const g = ac.createGain();
    osc.type = type;
    osc.frequency.setValueAtTime(freq, start);
    const peak = Math.max(gain, 0.0002);
    g.gain.setValueAtTime(0.0001, start);
    g.gain.exponentialRampToValueAtTime(peak, start + 0.01);
    g.gain.exponentialRampToValueAtTime(0.0001, start + dur);
    osc.connect(g);
    g.connect(dest ?? this.master ?? ac.destination);
    osc.start(start);
    osc.stop(start + dur + 0.02);
  }

  /** Generate a `dur`-second mono white-noise buffer. */
  private noiseBuffer(ac: AudioContext, dur: number): AudioBuffer {
    const length = Math.max(1, Math.floor(ac.sampleRate * dur));
    const buf = ac.createBuffer(1, length, ac.sampleRate);
    const data = buf.getChannelData(0);
    for (let i = 0; i < length; i++) {
      data[i] = Math.random() * 2 - 1;
    }
    return buf;
  }

  /** Quick upward "whoosh" played when the reels start spinning. */
  spinStart(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      const dur = 0.4;
      const src = ac.createBufferSource();
      src.buffer = this.noiseBuffer(ac, dur);
      const bp = ac.createBiquadFilter();
      bp.type = "bandpass";
      bp.Q.value = 1.2;
      bp.frequency.setValueAtTime(400, t);
      bp.frequency.exponentialRampToValueAtTime(2500, t + dur);
      const g = ac.createGain();
      g.gain.setValueAtTime(0.0001, t);
      g.gain.exponentialRampToValueAtTime(0.3, t + 0.08);
      g.gain.exponentialRampToValueAtTime(0.0001, t + dur);
      src.connect(bp);
      bp.connect(g);
      g.connect(this.master ?? ac.destination);
      src.start(t);
      src.stop(t + dur + 0.05);
    } catch {
      // ignore
    }
  }

  /** Short mechanical "thunk" played when a reel stops. */
  reelStop(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      this.playNote(ac, 120, t, 0.12, "sine", 0.4);
      this.playNote(ac, 800, t, 0.03, "square", 0.06);
    } catch {
      // ignore
    }
  }

  /** Pleasant ascending 3-note chime for small wins. */
  winSmall(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      const notes = [659.25, 880, 1318.5]; // E5, A5, E6
      const step = 0.1;
      notes.forEach((f, i) => {
        this.playNote(ac, f, t + i * step, 0.18, "triangle", 0.25);
      });
    } catch {
      // ignore
    }
  }

  /** Bigger ascending 5-note arpeggio with a shimmer layer for big wins. */
  winBig(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      const notes = [523.25, 659.25, 783.99, 1046.5, 1318.5]; // C5 E5 G5 C6 E6
      const step = 0.1;
      notes.forEach((f, i) => {
        const start = t + i * step;
        this.playNote(ac, f, start, 0.16, "triangle", 0.28);
        // detuned shimmer an octave up, low gain
        this.playNote(ac, f * 2, start, 0.16, "sine", 0.07);
      });
    } catch {
      // ignore
    }
  }

  /** Cascade of short metallic "clinks" for coin payouts. */
  coinDrop(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      const count = 7;
      for (let i = 0; i < count; i++) {
        const offset = (i / count) * 0.5 + Math.random() * 0.05;
        const base = 1200 + Math.random() * 200;
        this.playNote(ac, base, t + offset, 0.05, "sine", 0.12);
        this.playNote(ac, base * 1.5, t + offset, 0.05, "sine", 0.08);
      }
    } catch {
      // ignore
    }
  }

  /** UI click for buttons. */
  click(): void {
    try {
      const ac = this.ctx();
      this.playNote(ac, 600, ac.currentTime, 0.04, "square", 0.12);
    } catch {
      // ignore
    }
  }

  /** Subtle tick for bet changes. */
  tick(): void {
    try {
      const ac = this.ctx();
      this.playNote(ac, 900, ac.currentTime, 0.03, "triangle", 0.08);
    } catch {
      // ignore
    }
  }

  /** Short "pop" when winning symbols are removed during a tumble. */
  tumblePop(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      // bright blip with quick pitch drop
      const osc = ac.createOscillator();
      const g = ac.createGain();
      osc.type = "triangle";
      osc.frequency.setValueAtTime(1400, t);
      osc.frequency.exponentialRampToValueAtTime(500, t + 0.08);
      g.gain.setValueAtTime(0.0001, t);
      g.gain.exponentialRampToValueAtTime(0.18, t + 0.005);
      g.gain.exponentialRampToValueAtTime(0.0001, t + 0.1);
      osc.connect(g);
      g.connect(this.master ?? ac.destination);
      osc.start(t);
      osc.stop(t + 0.12);
    } catch {
      // ignore
    }
  }

  /** Whoosh + chime when a multiplier bomb is collected. */
  bomb(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      // descending whoosh
      const src = ac.createBufferSource();
      src.buffer = this.noiseBuffer(ac, 0.3);
      const bp = ac.createBiquadFilter();
      bp.type = "bandpass";
      bp.Q.value = 0.8;
      bp.frequency.setValueAtTime(2000, t);
      bp.frequency.exponentialRampToValueAtTime(300, t + 0.3);
      const g = ac.createGain();
      g.gain.setValueAtTime(0.0001, t);
      g.gain.exponentialRampToValueAtTime(0.2, t + 0.03);
      g.gain.exponentialRampToValueAtTime(0.0001, t + 0.3);
      src.connect(bp);
      bp.connect(g);
      g.connect(this.master ?? ac.destination);
      src.start(t);
      src.stop(t + 0.32);
      // bright bell
      this.playNote(ac, 1568, t, 0.3, "sine", 0.16); // G6
      this.playNote(ac, 2093, t + 0.05, 0.3, "sine", 0.1); // C7
    } catch {
      // ignore
    }
  }

  /** Magical sparkle sweep when free spins are triggered. */
  freeSpinsTrigger(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      const notes = [523.25, 587.33, 659.25, 783.99, 880, 1046.5]; // C5 D5 E5 G5 A5 C6
      const step = 0.12;
      notes.forEach((f, i) => {
        this.playNote(ac, f, t + i * step, 0.16, "triangle", 0.22);
      });
      // high sparkle blips scattered over ~1s
      for (let i = 0; i < 12; i++) {
        const freq = 2500 + Math.random() * 1500;
        const start = t + Math.random() * 1.0;
        this.playNote(ac, freq, start, 0.04, "sine", 0.05);
      }
    } catch {
      // ignore
    }
  }

  /** Celebratory fanfare for jackpots. */
  jackpot(): void {
    try {
      const ac = this.ctx();
      const t = ac.currentTime;
      // sustained low root (C3 + C4) underneath
      this.playNote(ac, 130.81, t, 1.4, "sine", 0.18); // C3
      this.playNote(ac, 261.63, t, 1.4, "triangle", 0.1); // C4
      // bright melody building energy then resolving
      const melody: Array<[number, number]> = [
        [523.25, 0.0], // C5
        [659.25, 0.15], // E5
        [783.99, 0.3], // G5
        [1046.5, 0.45], // C6
        [783.99, 0.65], // G5
        [1046.5, 0.8], // C6
        [1318.5, 0.95], // E6
        [1567.98, 1.1], // G6
      ];
      melody.forEach(([f, off]) => {
        const start = t + off;
        this.playNote(ac, f, start, 0.2, "triangle", 0.25);
        this.playNote(ac, f * 2, start, 0.2, "sine", 0.06);
      });
    } catch {
      // ignore
    }
  }
}

/** Shared singleton instance for the app. */
export const sound = new SoundEngine();
