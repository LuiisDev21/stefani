import { useEffect, useRef, useState } from "react";
import { RotateCcw } from "lucide-react";
import { audioEngine } from "../../audio/engine";
import { logBands, spectrumBars } from "../../audio/dsp";
import { EQ_FREQUENCIES, PRESETS, SPATIAL_MODES, SpatialMode } from "../../audio/settings";
import { SectionTitle } from "../../components/ui/SectionTitle";
import { Slider } from "../../components/ui/Slider";
import { Switch } from "../../components/ui/Switch";
import { useAudio } from "../../stores/audioStore";
import { ensureAudioElement } from "../../stores/playerStore";

const MODE_LABELS: Record<SpatialMode, string> = {
  off: "Apagado",
  subtle: "Suave",
  orbit: "Órbita",
  wide: "Ancho",
  immersive: "Inmersivo",
  custom: "Personalizado",
};

const BAND_LABELS = ["31", "62", "125", "250", "500", "1k", "2k", "4k", "8k", "16k"];

function dbLabel(value: number): string {
  const rounded = Math.round(value * 10) / 10;
  if (rounded === 0) return "0 dB";
  return `${rounded > 0 ? "+" : ""}${rounded} dB`;
}

/**
 * The equaliser panel: a real response curve, ten vertical bands and the
 * controls that actually move them. Every value here goes straight to the
 * audio graph, so what the curve shows is what the filters do.
 */
export function AudioPanel() {
  const settings = useAudio((state) => state.settings);
  const unsupported = useAudio((state) => state.unsupported);
  const loaded = useAudio((state) => state.loaded);
  const setEqEnabled = useAudio((state) => state.setEqEnabled);
  const setPreset = useAudio((state) => state.setPreset);
  const setBand = useAudio((state) => state.setBand);
  const setPreamp = useAudio((state) => state.setPreamp);
  const resetEq = useAudio((state) => state.resetEq);
  const setSpatialEnabled = useAudio((state) => state.setSpatialEnabled);
  const setSpatialMode = useAudio((state) => state.setSpatialMode);
  const setSpatial = useAudio((state) => state.setSpatial);
  const resetSpatial = useAudio((state) => state.resetSpatial);

  // Opening the panel is enough to build the graph, so the curve and the
  // spectrum work before anything has been played.
  useEffect(() => {
    const element = ensureAudioElement();
    audioEngine.attach(element);
  }, []);

  if (!loaded) return null;

  return (
    <section>
      <SectionTitle title="Audio" />

      {unsupported ? (
        <p className="rounded-[10px] border border-line bg-elevated px-4 py-3 text-[13px] leading-[1.45] text-muted">
          Este equipo no permite procesar el audio del reproductor: {unsupported}. La reproducción sigue
          funcionando sin efectos.
        </p>
      ) : (
        <>
          <div className="rounded-[10px] border border-line bg-elevated">
            <div className="flex flex-wrap items-center justify-between gap-3 px-4 pt-4">
              <div className="min-w-0">
                <Switch
                  checked={settings.eq.enabled}
                  onChange={setEqEnabled}
                  label="Ecualizador"
                  id="eq-enabled"
                />
              </div>
              <div className="flex items-center gap-2">
                <select
                  aria-label="Preset del ecualizador"
                  value={settings.eq.preset}
                  onChange={(event) => setPreset(event.target.value)}
                  disabled={!settings.eq.enabled}
                  className="h-8 rounded-[8px] border border-line bg-transparent px-2 text-[12.5px] text-ink transition-colors hover:border-line-strong disabled:opacity-40"
                >
                  {PRESETS.map((preset) => (
                    <option key={preset.id} value={preset.id} className="bg-elevated text-ink">
                      {preset.label}
                    </option>
                  ))}
                  <option value="custom" className="bg-elevated text-ink">
                    Personalizado
                  </option>
                </select>
                <button
                  type="button"
                  onClick={resetEq}
                  aria-label="Restablecer el ecualizador"
                  title="Restablecer el ecualizador"
                  className="flex h-8 w-8 items-center justify-center rounded-[8px] text-muted transition-colors hover:bg-surface hover:text-ink"
                >
                  <RotateCcw className="h-4 w-4" />
                </button>
              </div>
            </div>

            <EqCurve active={settings.eq.enabled} version={settings.eq.bands.join(",") + settings.eq.preamp} />

            <div className="px-4 pb-2">
              <div className="flex items-end justify-between gap-1">
                {EQ_FREQUENCIES.map((frequency, index) => (
                  <BandControl
                    key={frequency}
                    index={index}
                    label={BAND_LABELS[index]}
                    value={settings.eq.bands[index] ?? 0}
                    disabled={!settings.eq.enabled}
                    onChange={setBand}
                  />
                ))}
              </div>
            </div>

            <div className="flex items-center gap-3 border-t border-line px-4 py-3">
              <span className="w-[100px] shrink-0 text-[11.5px] text-muted">Preamplificador</span>
              <div className="min-w-0 flex-1">
                <Slider
                  value={settings.eq.preamp + 12}
                  max={24}
                  label="Preamplificador"
                  format={(value) => dbLabel(value - 12)}
                  disabled={!settings.eq.enabled}
                  onChange={(value) => setPreamp(Math.round((value - 12) * 2) / 2)}
                />
              </div>
              <span className="w-14 shrink-0 text-right text-[11.5px] text-muted tnum">
                {dbLabel(settings.eq.preamp)}
              </span>
            </div>
          </div>

          <div className="mt-4 rounded-[10px] border border-line bg-elevated">
            <div className="px-4 pt-4">
              <Switch
                checked={settings.spatial.enabled}
                onChange={setSpatialEnabled}
                label="Audio espacial"
                description="Mueve la música alrededor de la cabeza. Se disfruta con auriculares."
                id="spatial-enabled"
              />
            </div>

            <div className="mt-3 flex flex-wrap gap-1.5 px-4">
              {SPATIAL_MODES.map((mode) => {
                const active = settings.spatial.mode === mode;
                return (
                  <button
                    key={mode}
                    type="button"
                    aria-pressed={active}
                    disabled={!settings.spatial.enabled && mode !== "off"}
                    onClick={() => setSpatialMode(mode)}
                    className={
                      "rounded-full border px-2.5 py-[3px] text-[11.5px] transition-colors disabled:opacity-40 " +
                      (active
                        ? "border-accent bg-accent/6 text-accent"
                        : "border-line text-muted hover:border-line-strong hover:text-ink")
                    }
                  >
                    {MODE_LABELS[mode]}
                  </button>
                );
              })}
            </div>

            <div className="px-4 pb-4 pt-4">
              <div className="grid gap-x-6 gap-y-3 sm:grid-cols-2">
                <ControlRow
                  label="Intensidad"
                  value={`${Math.round(settings.spatial.intensity)}%`}
                  slider={
                    <Slider
                      value={settings.spatial.intensity}
                      max={100}
                      label="Intensidad espacial"
                      format={(value) => `${Math.round(value)}%`}
                      disabled={!settings.spatial.enabled}
                      onChange={(value) => setSpatial("intensity", value)}
                    />
                  }
                />
                <ControlRow
                  label="Mezcla"
                  value={`${Math.round(settings.spatial.mix)}%`}
                  slider={
                    <Slider
                      value={settings.spatial.mix}
                      max={100}
                      label="Mezcla de efecto"
                      format={(value) => `${Math.round(value)}%`}
                      disabled={!settings.spatial.enabled}
                      onChange={(value) => setSpatial("mix", value)}
                    />
                  }
                />
                <ControlRow
                  label="Velocidad"
                  value={`${settings.spatial.speed.toFixed(2)} Hz`}
                  slider={
                    <Slider
                      value={Math.round(settings.spatial.speed * 100)}
                      max={60}
                      label="Velocidad de giro"
                      format={(value) => `${(value / 100).toFixed(2)} Hz`}
                      disabled={!settings.spatial.enabled}
                      onChange={(value) => setSpatial("speed", Math.max(1, value) / 100)}
                    />
                  }
                />
                <ControlRow
                  label="Distancia"
                  value={`${settings.spatial.distance.toFixed(1)} m`}
                  slider={
                    <Slider
                      value={Math.round(settings.spatial.distance * 10)}
                      max={100}
                      label="Distancia virtual"
                      format={(value) => `${(value / 10).toFixed(1)} m`}
                      disabled={!settings.spatial.enabled}
                      onChange={(value) => setSpatial("distance", Math.max(10, value) / 10)}
                    />
                  }
                />
                <ControlRow
                  label="Reverberación"
                  value={`${Math.round(settings.spatial.reverb)}%`}
                  slider={
                    <Slider
                      value={settings.spatial.reverb}
                      max={100}
                      label="Reverberación"
                      format={(value) => `${Math.round(value)}%`}
                      disabled={!settings.spatial.enabled}
                      onChange={(value) => setSpatial("reverb", value)}
                    />
                  }
                />
                <ControlRow
                  label="Amplitud estéreo"
                  value={`${Math.round(settings.spatial.width)}%`}
                  slider={
                    <Slider
                      value={settings.spatial.width}
                      max={200}
                      label="Amplitud estéreo"
                      format={(value) => `${Math.round(value)}%`}
                      disabled={!settings.spatial.enabled}
                      onChange={(value) => setSpatial("width", value)}
                    />
                  }
                />
              </div>

              <div className="mt-4 flex items-center justify-between gap-4">
                <button
                  type="button"
                  onClick={() => setSpatial("direction", settings.spatial.direction === 1 ? -1 : 1)}
                  disabled={!settings.spatial.enabled}
                  className="text-[11.5px] text-muted transition-colors hover:text-ink disabled:opacity-40"
                >
                  {settings.spatial.direction === 1 ? "Sentido horario" : "Sentido antihorario"}
                </button>
                <button
                  type="button"
                  onClick={resetSpatial}
                  className="text-[11.5px] text-muted transition-colors hover:text-ink"
                >
                  Restablecer
                </button>
              </div>
            </div>

            <SpatialRadar active={settings.spatial.enabled && settings.spatial.mode !== "off"} />
          </div>

          <Spectrum />
        </>
      )}
    </section>
  );
}

function ControlRow({ label, value, slider }: { label: string; value: string; slider: React.ReactNode }) {
  return (
    <div className="min-w-0">
      <div className="mb-1.5 flex items-baseline justify-between gap-3">
        <span className="truncate text-[11.5px] text-muted">{label}</span>
        <span className="shrink-0 text-[11.5px] text-muted tnum">{value}</span>
      </div>
      {slider}
    </div>
  );
}

function BandControl({
  index,
  label,
  value,
  disabled,
  onChange,
}: {
  index: number;
  label: string;
  value: number;
  disabled: boolean;
  onChange: (index: number, value: number) => void;
}) {
  const id = `eq-band-${index}`;
  return (
    <div className="flex min-w-0 flex-1 flex-col items-center gap-1.5">
      <span className="text-[10.5px] text-faint tnum">{dbLabel(value).replace(" dB", "")}</span>
      <input
        id={id}
        type="range"
        min={-12}
        max={12}
        step={0.5}
        value={value}
        disabled={disabled}
        aria-label={`Banda ${label} hercios`}
        aria-valuetext={dbLabel(value)}
        onDoubleClick={() => onChange(index, 0)}
        onChange={(event) => onChange(index, Number(event.target.value))}
        className="stefani-band"
      />
      <span className="text-[10.5px] text-faint tnum">{label}</span>
    </div>
  );
}

/**
 * The curve is computed from the filters themselves with getFrequencyResponse,
 * not drawn from the slider positions: a shelf and a peaking filter do not
 * have the same shape, and pretending they do is how EQ graphs start lying.
 */
function EqCurve({ active, version }: { active: boolean; version: string }) {
  const [points, setPoints] = useState<string>("");

  useEffect(() => {
    const width = 100;
    const height = 100;
    const samples = 160;
    const frequencies = new Float32Array(samples);
    for (let i = 0; i < samples; i += 1) {
      const t = i / (samples - 1);
      frequencies[i] = 20 * Math.pow(1000, t); // 20 Hz .. 20 kHz
    }
    const response = audioEngine.frequencyResponse(frequencies);
    const real = response ?? new Float32Array(samples);
    const path = Array.from(real)
      .map((db, index) => {
        const x = (index / (samples - 1)) * width;
        const y = height / 2 - (Math.max(-15, Math.min(15, db)) / 15) * (height / 2);
        return `${index === 0 ? "M" : "L"}${x.toFixed(2)},${y.toFixed(2)}`;
      })
      .join(" ");
    setPoints(path);
  }, [version, active]);

  return (
    <div className="relative mx-4 mt-3 overflow-hidden rounded-[8px] border border-line/60 bg-[#0b0c0d]">
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="block h-[96px] w-full" aria-hidden="true">
        <line x1="0" y1="50" x2="100" y2="50" stroke="#242728" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
        {[12.5, 25, 37.5, 62.5, 75, 87.5].map((y) => (
          <line
            key={y}
            x1="0"
            y1={y}
            x2="100"
            y2={y}
            stroke="#1b1d1e"
            strokeWidth="0.5"
            vectorEffect="non-scaling-stroke"
          />
        ))}
        {points ? (
          <path
            d={points}
            fill="none"
            stroke={active ? "var(--color-ink)" : "#5a5c5e"}
            strokeWidth="1.4"
            vectorEffect="non-scaling-stroke"
            strokeLinejoin="round"
          />
        ) : null}
      </svg>
      <div className="pointer-events-none absolute inset-x-0 bottom-0 flex justify-between px-2 pb-1 text-[9.5px] text-faint">
        <span>20</span>
        <span>100</span>
        <span>1k</span>
        <span>10k</span>
        <span>20k</span>
      </div>
    </div>
  );
}

/** The live spectrum, read from the analyser at the end of the chain. */
function Spectrum() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const wrapRef = useRef<HTMLDivElement | null>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const wrap = wrapRef.current;
    if (!wrap) return;
    const observer = new IntersectionObserver((entries) => {
      setVisible(entries.some((entry) => entry.isIntersecting));
    });
    observer.observe(wrap);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (!visible) return;
    let frame = 0;
    let last = 0;
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx2d = canvas.getContext("2d");
    if (!ctx2d) return;

    const draw = (time: number) => {
      frame = requestAnimationFrame(draw);
      if (document.hidden) return;
      if (time - last < 33) return; // ~30 fps is plenty for a spectrum
      last = time;

      const context = audioEngine.getContext();
      const bars = logBands(48, 1024, context?.sampleRate ?? 44100, 2048);
      const data = new Uint8Array(1024);
      const ok = audioEngine.readSpectrum(data);
      const values = ok ? spectrumBars(data, bars, { minDb: -90, maxDb: -10 }) : new Float32Array(bars.length);

      const width = canvas.clientWidth;
      const height = canvas.clientHeight;
      const dpr = window.devicePixelRatio || 1;
      if (canvas.width !== Math.floor(width * dpr) || canvas.height !== Math.floor(height * dpr)) {
        canvas.width = Math.floor(width * dpr);
        canvas.height = Math.floor(height * dpr);
      }
      ctx2d.setTransform(dpr, 0, 0, dpr, 0, 0);
      ctx2d.clearRect(0, 0, width, height);
      const gap = 2;
      const barWidth = Math.max(1, (width - gap * (values.length - 1)) / values.length);
      ctx2d.fillStyle = "rgba(244, 244, 246, 0.62)";
      values.forEach((value, index) => {
        const barHeight = Math.max(1, value * height);
        ctx2d.fillRect(index * (barWidth + gap), height - barHeight, barWidth, barHeight);
      });
    };

    frame = requestAnimationFrame(draw);
    return () => cancelAnimationFrame(frame);
  }, [visible]);

  return (
    <div ref={wrapRef} className="mt-4 rounded-[10px] border border-line bg-elevated px-4 py-3">
      <div className="mb-2 flex items-baseline justify-between">
        <span className="text-[11.5px] text-muted">Espectro</span>
        <span className="text-[11.5px] text-faint">30 Hz a 18 kHz</span>
      </div>
      <canvas ref={canvasRef} className="block h-[64px] w-full" aria-label="Espectro de la reproducción" />
    </div>
  );
}

/** Where the source really is, straight from the panner's audio params. */
function SpatialRadar({ active }: { active: boolean }) {
  const [dot, setDot] = useState({ x: 50, y: 50 });

  useEffect(() => {
    if (!active) {
      setDot({ x: 50, y: 50 });
      return;
    }
    let frame = 0;
    let last = 0;
    const tick = (time: number) => {
      frame = requestAnimationFrame(tick);
      if (time - last < 50) return;
      last = time;
      const position = audioEngine.readPosition();
      const radius = Math.max(1, Math.hypot(position.x, position.z));
      // Normalise against the widest the engine will ever set.
      const scale = Math.min(1, radius / 3);
      const angle = Math.atan2(position.z, position.x);
      setDot({ x: 50 + Math.cos(angle) * scale * 42, y: 50 + Math.sin(angle) * scale * 42 });
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [active]);

  return (
    <div className="flex items-center justify-center border-t border-line px-4 py-4">
      <svg viewBox="0 0 100 100" className="h-[92px] w-[92px]" aria-hidden="true">
        <circle cx="50" cy="50" r="42" fill="none" stroke="#242728" strokeWidth="0.8" />
        <circle cx="50" cy="50" r="28" fill="none" stroke="#1c1e1f" strokeWidth="0.6" />
        <circle cx="50" cy="50" r="14" fill="none" stroke="#1c1e1f" strokeWidth="0.6" />
        <circle cx="50" cy="50" r="1.6" fill="#5a5c5e" />
        <circle
          cx={dot.x}
          cy={dot.y}
          r="3"
          fill={active ? "var(--color-ink)" : "#4a4c4e"}
          style={{ transition: "cx 90ms linear, cy 90ms linear" }}
        />
      </svg>
    </div>
  );
}
