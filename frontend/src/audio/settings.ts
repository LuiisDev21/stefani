/**
 * Audio settings: the shape that is persisted, the presets that ship with the
 * app, and the small pieces of maths the engine and the UI share.
 *
 * Everything here is pure: no Web Audio, no DOM. That keeps the rules that
 * decide what a preset means and what a valid value is testable without a
 * sound card.
 */

export const EQ_FREQUENCIES = [31, 62, 125, 250, 500, 1000, 2000, 4000, 8000, 16000] as const;
export const EQ_BANDS = EQ_FREQUENCIES.length;
export const GAIN_MIN = -12;
export const GAIN_MAX = 12;

export const SPATIAL_MODES = ["off", "subtle", "orbit", "wide", "immersive", "custom"] as const;
export type SpatialMode = (typeof SPATIAL_MODES)[number];

export interface EqSettings {
  enabled: boolean;
  preset: string;
  /** Gains in dB, one per band, same order as EQ_FREQUENCIES. */
  bands: number[];
  /** Input gain in dB, compensated by the preset headroom. */
  preamp: number;
}

export interface SpatialSettings {
  enabled: boolean;
  mode: SpatialMode;
  /** 0-100: how far the source travels from the centre. */
  intensity: number;
  /** Revolutions per second for the orbit; slow by default. */
  speed: number;
  /** 1 clockwise, -1 counter-clockwise. */
  direction: 1 | -1;
  /** Virtual distance in metres, 1-10. */
  distance: number;
  /** 0-100: how much of the wet path goes through the room. */
  reverb: number;
  /** 0-100: dry/wet mix. */
  mix: number;
  /** 0-100: stereo width for the wide mode, 100 leaves the image alone. */
  width: number;
}

export interface AudioSettings {
  version: number;
  eq: EqSettings;
  spatial: SpatialSettings;
}

export const SETTINGS_VERSION = 1;

export interface Preset {
  id: string;
  label: string;
  bands: number[];
  preamp: number;
}

/**
 * Presets are written with the headroom already taken out of the preamp: a
 * curve that adds +8 dB in the bass is scaled back so the loudest boost cannot
 * push a normal master past full scale.
 */
export const PRESETS: Preset[] = [
  { id: "flat", label: "Plano", bands: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0], preamp: 0 },
  { id: "bass-boost", label: "Graves", bands: [6, 5, 3, 1, 0, 0, 0, 0, 0, 0], preamp: -3 },
  { id: "deep-bass", label: "Graves profundos", bands: [8, 7, 5, 2, 0, 0, 0, 0, -0.5, -1], preamp: -5 },
  { id: "treble-boost", label: "Agudos", bands: [0, 0, 0, 0, 0, 1, 3, 5, 6, 6], preamp: -3 },
  { id: "vocal", label: "Voz clara", bands: [-2, -1, 0, 2, 4, 4, 3, 1, 0, -1], preamp: -2 },
  { id: "pop", label: "Pop", bands: [-1, 0, 1, 2, 2, 1, 0, 1, 2, 2], preamp: -1 },
  { id: "rock", label: "Rock", bands: [4, 3, 1, -2, -1, 1, 3, 4, 4, 3], preamp: -3 },
  { id: "electronic", label: "Electrónica", bands: [5, 4, 1, 0, -2, 1, 2, 3, 5, 5], preamp: -3 },
  { id: "hip-hop", label: "Hip-hop", bands: [6, 5, 3, 1, -1, 0, 1, 2, 3, 3], preamp: -4 },
  { id: "classical", label: "Clásica", bands: [2, 1, 0, 0, 0, 0, -1, -2, -2, -2], preamp: -1 },
  { id: "jazz", label: "Jazz", bands: [3, 2, 1, 1, -1, -1, 0, 1, 2, 3], preamp: -2 },
  { id: "acoustic", label: "Acústica", bands: [2, 1, 1, 1, 1, 0, 1, 2, 3, 3], preamp: -2 },
];

export const CUSTOM_PRESET = "custom";

export function defaultAudioSettings(): AudioSettings {
  return {
    version: SETTINGS_VERSION,
    eq: { enabled: false, preset: "flat", bands: [...PRESETS[0].bands], preamp: 0 },
    spatial: {
      enabled: false,
      mode: "off",
      intensity: 55,
      speed: 0.06,
      direction: 1,
      distance: 2.5,
      reverb: 22,
      mix: 45,
      width: 100,
    },
  };
}

function clamp(value: unknown, min: number, max: number, fallback: number): number {
  const n = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(n)) return fallback;
  return Math.min(max, Math.max(min, n));
}

function readBands(input: unknown, fallback: number[]): number[] {
  if (!Array.isArray(input) || input.length !== EQ_BANDS) return [...fallback];
  return input.map((value, index) => clamp(value, GAIN_MIN, GAIN_MAX, fallback[index] ?? 0));
}

/**
 * Takes whatever came out of storage and returns something the engine can use.
 * A corrupt file, an old version or a value edited by hand must never leave
 * the player with a silent graph, so every field falls back on its default.
 */
export function clampAudioSettings(input: unknown): AudioSettings {
  const base = defaultAudioSettings();
  if (!input || typeof input !== "object") return base;
  const raw = input as Partial<AudioSettings>;
  const rawEq = (raw.eq ?? {}) as Partial<EqSettings>;
  const rawSpatial = (raw.spatial ?? {}) as Partial<SpatialSettings>;

  const presetId = typeof rawEq.preset === "string" ? rawEq.preset : base.eq.preset;
  const knownPreset = PRESETS.find((preset) => preset.id === presetId);
  const preset = knownPreset || presetId === CUSTOM_PRESET ? presetId : base.eq.preset;
  const presetBands = knownPreset ? knownPreset.bands : base.eq.bands;
  const presetPreamp = knownPreset ? knownPreset.preamp : base.eq.preamp;

  const mode = SPATIAL_MODES.includes(rawSpatial.mode as SpatialMode)
    ? (rawSpatial.mode as SpatialMode)
    : base.spatial.mode;

  return {
    version: SETTINGS_VERSION,
    eq: {
      enabled: typeof rawEq.enabled === "boolean" ? rawEq.enabled : base.eq.enabled,
      preset,
      bands: readBands(rawEq.bands, presetBands),
      preamp: clamp(rawEq.preamp, GAIN_MIN, GAIN_MAX, presetPreamp),
    },
    spatial: {
      enabled: typeof rawSpatial.enabled === "boolean" ? rawSpatial.enabled : base.spatial.enabled,
      mode,
      intensity: clamp(rawSpatial.intensity, 0, 100, base.spatial.intensity),
      speed: clamp(rawSpatial.speed, 0.01, 0.6, base.spatial.speed),
      direction: rawSpatial.direction === -1 ? -1 : 1,
      distance: clamp(rawSpatial.distance, 1, 10, base.spatial.distance),
      reverb: clamp(rawSpatial.reverb, 0, 100, base.spatial.reverb),
      mix: clamp(rawSpatial.mix, 0, 100, base.spatial.mix),
      width: clamp(rawSpatial.width, 0, 200, base.spatial.width),
    },
  };
}

/** A curve that no longer matches its preset is a custom curve, not that preset. */
export function presetMatchesBands(preset: Preset, bands: number[]): boolean {
  return preset.bands.every((gain, index) => Math.abs(gain - (bands[index] ?? 0)) < 0.05);
}

/**
 * Equal-power mix: the two gains keep their amplitudes summed in power, so
 * sweeping dry to wet does not climb 6 dB in the middle or dip on the way.
 */
export function equalPowerGains(mixPercent: number): { dry: number; wet: number } {
  const theta = (clamp(mixPercent, 0, 100, 0) / 100) * (Math.PI / 2);
  return { dry: Math.cos(theta), wet: Math.sin(theta) };
}

/**
 * Position on the horizontal circle around the listener at time t. The engine
 * feeds these values to the panner from the audio clock, so a stalled UI does
 * not stall the orbit.
 */
export function orbitPosition(
  t: number,
  radius: number,
  speed: number,
  direction: 1 | -1,
): { x: number; y: number; z: number } {
  const angle = 2 * Math.PI * speed * t * direction;
  return { x: radius * Math.cos(angle), y: 0, z: radius * Math.sin(angle) };
}
