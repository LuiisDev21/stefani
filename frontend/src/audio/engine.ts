/**
 * The audio engine: one graph, built once, fed by the player's own `<audio>`
 * element.
 *
 * The chain is deliberately boring — gain, ten filters, a spatial branch, a
 * limiter — and never rebuilt while playing. Effects are switched by ramping
 * gains, so nothing clicks and no node is ever orphaned. The only thing that
 * gets connected and disconnected is the spatial branch, because an HRTF
 * panner costs real CPU and has no business running while it is muted.
 */

import { sumResponseDb } from "./dsp";
import {
  AudioSettings,
  EQ_BANDS,
  EQ_FREQUENCIES,
  equalPowerGains,
  orbitPosition,
} from "./settings";

type AnyPanner = PannerNode & {
  setPosition?: (x: number, y: number, z: number) => void;
};

const SMOOTH = 0.03;
const ORBIT_STEP = 0.04;
const ORBIT_LOOKAHEAD = 0.15;

function ramp(param: AudioParam, value: number, ctx: AudioContext, tau = SMOOTH): void {
  const now = ctx.currentTime;
  param.cancelScheduledValues(now);
  param.setTargetAtTime(value, now, tau);
}

export interface EqChain {
  input: GainNode;
  output: GainNode;
  filters: BiquadFilterNode[];
}

/**
 * Builds the preamp and the ten filters that make up the EQ. Kept apart from
 * the rest of the graph so an offline render can measure exactly the chain the
 * player uses.
 */
export function buildEqChain(ctx: BaseAudioContext, settings: AudioSettings): EqChain {
  const input = ctx.createGain();
  const eqOn = settings.eq.enabled;
  input.gain.value = Math.pow(10, (eqOn ? settings.eq.preamp : 0) / 20);

  const filters: BiquadFilterNode[] = [];
  let previous: AudioNode = input;
  EQ_FREQUENCIES.forEach((frequency, index) => {
    const filter = ctx.createBiquadFilter();
    if (index === 0) {
      filter.type = "lowshelf";
    } else if (index === EQ_FREQUENCIES.length - 1) {
      filter.type = "highshelf";
    } else {
      filter.type = "peaking";
      filter.Q.value = 1.1;
    }
    filter.frequency.value = frequency;
    filter.gain.value = eqOn ? (settings.eq.bands[index] ?? 0) : 0;
    previous.connect(filter);
    previous = filter;
    filters.push(filter);
  });

  const output = ctx.createGain();
  previous.connect(output);
  return { input, output, filters };
}

export interface EngineStatus {
  supported: boolean;
  hrtf: boolean;
  reason?: string;
}

export class AudioEngine {
  private ctx?: AudioContext;
  private element?: HTMLAudioElement;
  private status: EngineStatus = { supported: false, hrtf: false };

  private source?: MediaElementAudioSourceNode;
  private inputGain?: GainNode;
  private filters: BiquadFilterNode[] = [];
  private eqOut?: GainNode;
  private dryGain?: GainNode;
  private wetGain?: GainNode;
  private limiter?: DynamicsCompressorNode;
  private analyser?: AnalyserNode;
  private panner?: AnyPanner;
  private convolver?: ConvolverNode;
  private reverbIn?: GainNode;
  private airFilter?: BiquadFilterNode;
  private sideExtra?: GainNode;

  /** Mirrors the ten filters with the target gains so the curve can be read
   *  back: an AudioParam driven by setTargetAtTime does not report its
   *  automated value through .value, and the graph must not be touched just to
   *  draw a line. */
  private probe?: BiquadFilterNode[];

  private spatialRouted = false;
  private orbitTimer?: ReturnType<typeof setInterval>;
  private scheduledUntil = 0;
  private settings: AudioSettings | null = null;

  /** True once the graph exists and the element is wired to the context. */
  get attached(): boolean {
    return Boolean(this.source);
  }

  getStatus(): EngineStatus {
    return this.status;
  }

  getContext(): AudioContext | undefined {
    return this.ctx;
  }

  /**
   * Wires the player's element into the graph. Calling it twice with the same
   * element is safe; a different element rebuilds nothing, because the player
   * only ever creates one.
   */
  attach(element: HTMLAudioElement): EngineStatus {
    if (this.attached && this.element === element) return this.status;
    if (typeof window === "undefined") {
      this.status = { supported: false, hrtf: false, reason: "sin ventana" };
      return this.status;
    }
    const Ctor = window.AudioContext ?? (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
    if (!Ctor) {
      this.status = { supported: false, hrtf: false, reason: "este equipo no expone Web Audio" };
      return this.status;
    }

    const ctx = new Ctor({ latencyHint: "playback" });
    let hrtf = false;
    let source: MediaElementAudioSourceNode;
    try {
      source = ctx.createMediaElementSource(element);
    } catch (error) {
      void ctx.close().catch(() => undefined);
      this.status = {
        supported: false,
        hrtf: false,
        reason: error instanceof Error ? error.message : "no se pudo leer el audio del reproductor",
      };
      return this.status;
    }

    this.ctx = ctx;
    this.element = element;
    this.source = source;

    // Input, preamplifier and the ten bands, built by the shared helper.
    const eqChain = buildEqChain(ctx, this.settings ?? {
      version: 1,
      eq: { enabled: true, preset: "flat", bands: new Array(EQ_BANDS).fill(0), preamp: 0 },
      spatial: {
        enabled: false,
        mode: "off",
        intensity: 0,
        speed: 0.06,
        direction: 1,
        distance: 2.5,
        reverb: 0,
        mix: 0,
        width: 100,
      },
    });
    source.connect(eqChain.input);
    const inputGain = eqChain.input;
    this.filters = eqChain.filters;
    const eqOut = eqChain.output;

    // Spatial branch: dry keeps the stereo image untouched, wet is panned.
    const dryGain = ctx.createGain();
    dryGain.gain.value = 1;
    const wetGain = ctx.createGain();
    wetGain.gain.value = 0;
    const mixIn = ctx.createGain();

    eqOut.connect(dryGain);
    dryGain.connect(mixIn);

    const panner = ctx.createPanner() as AnyPanner;
    try {
      panner.panningModel = "HRTF";
      hrtf = panner.panningModel === "HRTF";
    } catch {
      hrtf = false;
    }
    if (!hrtf) {
      try {
        panner.panningModel = "equalpower";
      } catch {
        // The node still works with its default model.
      }
    }
    panner.distanceModel = "inverse";
    panner.refDistance = 1;
    panner.rolloffFactor = 0.6;
    panner.maxDistance = 30;
    panner.coneInnerAngle = 360;

    // Air absorption: distance dulls the highs the way a real room does.
    const airFilter = ctx.createBiquadFilter();
    airFilter.type = "lowpass";
    airFilter.frequency.value = 20000;
    airFilter.Q.value = 0.0001;

    // Reverb: a short synthetic impulse, rebuilt only when its length changes.
    const convolver = ctx.createConvolver();
    convolver.normalize = true;
    const reverbIn = ctx.createGain();
    reverbIn.gain.value = 0.2;
    const reverbOut = ctx.createGain();
    reverbOut.gain.value = 1;

    panner.connect(airFilter);
    airFilter.connect(wetGain);
    airFilter.connect(reverbIn);
    reverbIn.connect(convolver);
    convolver.connect(reverbOut);
    reverbOut.connect(wetGain);
    wetGain.connect(mixIn);

    // Stereo width: only the side signal is boosted, so the mid stays put and
    // a wide setting never collapses the image to mono.
    const sideSplit = ctx.createChannelSplitter(2);
    const sideLeft = ctx.createGain();
    const sideRight = ctx.createGain();
    sideLeft.gain.value = 0.5;
    sideRight.gain.value = -0.5;
    const sideSum = ctx.createGain();
    const sideExtra = ctx.createGain();
    sideExtra.gain.value = 0;
    eqOut.connect(sideSplit);
    sideSplit.connect(sideLeft, 0);
    sideSplit.connect(sideRight, 1);
    sideLeft.connect(sideSum);
    sideRight.connect(sideSum);
    sideSum.connect(sideExtra);
    sideExtra.connect(mixIn);

    // Output protection. With ratio 1 and threshold at full scale it does
    // nothing, which is exactly what a full bypass should sound like.
    const limiter = ctx.createDynamicsCompressor();
    limiter.threshold.value = -1.5;
    limiter.knee.value = 0;
    limiter.ratio.value = 1;
    limiter.attack.value = 0.003;
    limiter.release.value = 0.25;

    const outputGain = ctx.createGain();
    outputGain.gain.value = 1;

    const analyser = ctx.createAnalyser();
    analyser.fftSize = 2048;
    analyser.smoothingTimeConstant = 0.82;
    analyser.minDecibels = -90;
    analyser.maxDecibels = -10;

    mixIn.connect(limiter);
    limiter.connect(outputGain);
    outputGain.connect(analyser);
    analyser.connect(ctx.destination);

    this.inputGain = inputGain;
    this.eqOut = eqOut;
    this.dryGain = dryGain;
    this.wetGain = wetGain;
    this.limiter = limiter;
    this.analyser = analyser;
    this.panner = panner;
    this.convolver = convolver;
    this.reverbIn = reverbIn;
    this.airFilter = airFilter;
    this.sideExtra = sideExtra;

    this.status = { supported: true, hrtf };
    if (this.settings) this.apply(this.settings);
    return this.status;
  }

  /** The context is born suspended until a user gesture; play is that gesture. */
  resume(): void {
    const ctx = this.ctx;
    if (!ctx) return;
    if (ctx.state === "suspended") void ctx.resume().catch(() => undefined);
  }

  apply(settings: AudioSettings): void {
    this.settings = settings;
    const ctx = this.ctx;
    if (!ctx || !this.inputGain || !this.filters.length) return;

    const eqOn = settings.eq.enabled;
    const preampDb = eqOn ? settings.eq.preamp : 0;
    ramp(this.inputGain.gain, Math.pow(10, preampDb / 20), ctx);

    this.filters.forEach((filter, index) => {
      const gain = eqOn ? (settings.eq.bands[index] ?? 0) : 0;
      ramp(filter.gain, gain, ctx);
    });

    this.applySpatial(settings);
  }

  private applySpatial(settings: AudioSettings): void {
    const ctx = this.ctx;
    if (!ctx || !this.dryGain || !this.wetGain || !this.panner) return;
    const spatial = settings.spatial;
    const active = spatial.enabled && spatial.mode !== "off";

    if (!active) {
      // Fade the wet path out, then stop feeding it: an idle HRTF panner is
      // pure CPU burn.
      ramp(this.wetGain.gain, 0, ctx, 0.05);
      ramp(this.dryGain.gain, 1, ctx, 0.05);
      if (this.sideExtra) ramp(this.sideExtra.gain, 0, ctx, 0.05);
      if (this.spatialRouted) {
        const panner = this.panner;
        const eqOut = this.eqOut;
        window.setTimeout(() => {
          if (!this.spatialRouted) {
            try {
              eqOut?.disconnect(panner);
            } catch {
              // Already disconnected.
            }
          }
        }, 140);
        this.spatialRouted = false;
        this.stopOrbit();
      }
      return;
    }

    if (!this.spatialRouted) {
      try {
        this.eqOut?.connect(this.panner);
      } catch {
        // Already connected.
      }
      this.spatialRouted = true;
    }

    const { dry, wet } = equalPowerGains(spatial.mix);
    ramp(this.dryGain.gain, dry, ctx, 0.06);
    ramp(this.wetGain.gain, wet, ctx, 0.06);

    if (spatial.mode === "wide") {
      this.stopOrbit();
      if (this.sideExtra) {
        const extra = Math.max(0, spatial.width / 100 - 1);
        ramp(this.sideExtra.gain, extra, ctx, 0.08);
      }
      ramp(this.wetGain.gain, 0, ctx, 0.06);
      return;
    }

    if (this.sideExtra) ramp(this.sideExtra.gain, 0, ctx, 0.08);
    this.updateReverb(spatial.reverb);
    this.updateDistance(spatial.distance);

    if (spatial.mode === "orbit" || spatial.mode === "immersive") {
      this.startOrbit();
    } else {
      this.stopOrbit();
      // Subtle and custom park the source slightly off-centre and let the
      // intensity decide how far.
      const spread = (spatial.intensity / 100) * 0.8;
      this.setPosition(spatial.direction > 0 ? spread : -spread, 0, -0.4);
    }
  }

  private updateReverb(amount: number): void {
    const ctx = this.ctx;
    if (!ctx || !this.convolver || !this.reverbIn) return;
    const seconds = 0.6 + (amount / 100) * 2.2;
    const rounded = Math.round(seconds * 4) / 4;
    if (!this.convolver.buffer || Math.abs((this.convolver.buffer.duration || 0) - rounded) > 0.2) {
      this.convolver.buffer = this.buildImpulse(rounded);
    }
    ramp(this.reverbIn.gain, (amount / 100) * 0.55, ctx, 0.1);
  }

  private updateDistance(distance: number): void {
    const ctx = this.ctx;
    if (!ctx || !this.panner || !this.airFilter) return;
    this.panner.refDistance = 1;
    this.panner.rolloffFactor = Math.min(2, 0.35 + distance / 6);
    // Far sources lose their top end before they lose their level.
    const cutoff = Math.min(20000, Math.max(1800, 20000 / (1 + distance / 2.2)));
    ramp(this.airFilter.frequency, cutoff, ctx, 0.12);
  }

  /**
   * A synthetic room: exponentially decaying noise. Cheap, deterministic and
   * good enough for a 20% wet send; a real impulse would be megabytes of
   * samples shipped with the app for no audible gain at this mix level.
   */
  private buildImpulse(seconds: number): AudioBuffer {
    const ctx = this.ctx!;
    const rate = ctx.sampleRate;
    const length = Math.max(1, Math.floor(rate * seconds));
    const buffer = ctx.createBuffer(2, length, rate);
    for (let channel = 0; channel < 2; channel += 1) {
      const data = buffer.getChannelData(channel);
      for (let i = 0; i < length; i += 1) {
        const decay = Math.pow(1 - i / length, 2.6);
        data[i] = (Math.random() * 2 - 1) * decay;
      }
    }
    return buffer;
  }

  private startOrbit(): void {
    if (this.orbitTimer) return;
    const ctx = this.ctx;
    if (!ctx) return;
    this.scheduledUntil = ctx.currentTime;
    this.orbitTimer = setInterval(() => this.pushOrbit(), 50);
    this.pushOrbit();
  }

  private stopOrbit(): void {
    if (this.orbitTimer) {
      clearInterval(this.orbitTimer);
      this.orbitTimer = undefined;
    }
  }

  /**
   * Positions are written into the audio param timeline ahead of time, not on
   * every animation frame: a frontend that drops to 5 fps still gets a smooth
   * orbit because the audio thread interpolates what was already scheduled.
   */
  private pushOrbit(): void {
    const ctx = this.ctx;
    const panner = this.panner;
    if (!ctx || !panner || !this.settings) return;
    const spatial = this.settings.spatial;
    const radius = 0.6 + (spatial.intensity / 100) * 2.2;
    const horizon = ctx.currentTime + ORBIT_LOOKAHEAD;
    let t = Math.max(this.scheduledUntil, ctx.currentTime);
    if (t > horizon) return;
    while (t < horizon) {
      const position = orbitPosition(t, radius, spatial.speed, spatial.direction);
      this.setPositionAt(position, t);
      t += ORBIT_STEP;
    }
    this.scheduledUntil = t;
  }

  private setPositionAt(position: { x: number; y: number; z: number }, when: number): void {
    const panner = this.panner;
    if (!panner) return;
    if (panner.positionX) {
      panner.positionX.setValueAtTime(position.x, when);
      panner.positionY.setValueAtTime(position.y, when);
      panner.positionZ.setValueAtTime(position.z, when);
      return;
    }
    panner.setPosition?.(position.x, position.y, position.z);
  }

  private setPosition(x: number, y: number, z: number): void {
    const ctx = this.ctx;
    const panner = this.panner;
    if (!ctx || !panner) return;
    const when = ctx.currentTime;
    if (panner.positionX) {
      ramp(panner.positionX, x, ctx, 0.12);
      ramp(panner.positionY, y, ctx, 0.12);
      ramp(panner.positionZ, z, ctx, 0.12);
      return;
    }
    panner.setPosition?.(x, y, z);
    void when;
  }

  /** Real spectrum, in 0-255 magnitudes. Returns false when nothing is playing. */
  readSpectrum(target: Uint8Array): boolean {
    if (!this.analyser) return false;
    this.analyser.getByteFrequencyData(target as Uint8Array<ArrayBuffer>);
    return true;
  }

  /** Where the source actually is right now, for the panel's radar. */
  readPosition(): { x: number; y: number; z: number } {
    const panner = this.panner;
    if (!panner) return { x: 0, y: 0, z: 0 };
    return { x: panner.positionX?.value ?? 0, y: panner.positionY?.value ?? 0, z: panner.positionZ?.value ?? 0 };
  }

  /**
   * Frequency response of the ten filters combined, in dB, sampled for drawing.
   * This is the real curve the audio goes through, not a picture of the gains.
   */
  frequencyResponse(frequencies: Float32Array): Float32Array | null {
    const settings = this.settings;
    if (!settings) return null;
    const filters = this.probeFilters(settings);
    if (!filters.length) return null;

    const magnitude = new Float32Array(frequencies.length);
    const phase = new Float32Array(frequencies.length);
    const responses = filters.map((filter) => {
      filter.getFrequencyResponse(frequencies as Float32Array<ArrayBuffer>, magnitude, phase);
      const db = new Float32Array(frequencies.length);
      for (let i = 0; i < frequencies.length; i += 1) {
        db[i] = 20 * Math.log10(Math.max(1e-6, magnitude[i]));
      }
      return db;
    });
    return sumResponseDb(responses, frequencies.length);
  }

  /** A silent copy of the band filters, kept in step with the target values. */
  private probeFilters(settings: AudioSettings): BiquadFilterNode[] {
    const scope = globalThis as unknown as { OfflineAudioContext?: typeof OfflineAudioContext };
    if (!scope.OfflineAudioContext) return this.filters;
    if (!this.probe) {
      const ctx = new scope.OfflineAudioContext(1, 128, 44100);
      this.probe = buildEqChain(ctx, settings).filters;
    }
    const eqOn = settings.eq.enabled;
    this.probe.forEach((filter, index) => {
      filter.gain.value = eqOn ? (settings.eq.bands[index] ?? 0) : 0;
    });
    return this.probe;
  }

  dispose(): void {
    this.stopOrbit();
    try {
      this.source?.disconnect();
      this.analyser?.disconnect();
      this.limiter?.disconnect();
    } catch {
      // Nodes may already be gone during teardown.
    }
    void this.ctx?.close().catch(() => undefined);
    this.ctx = undefined;
    this.source = undefined;
    this.element = undefined;
    this.filters = [];
  }
}

export const audioEngine = new AudioEngine();
export { EQ_BANDS, EQ_FREQUENCIES };
