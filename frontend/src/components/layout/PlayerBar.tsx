import {
  Heart,
  Maximize2,
  Pause,
  Play,
  Repeat,
  Repeat1,
  SkipBack,
  SkipForward,
  Volume1,
  Volume2,
  VolumeX,
} from "lucide-react";
import { formatTime } from "../../lib/format";
import { cn } from "../../lib/cn";
import { useApp } from "../../stores/appStore";
import { PLAYBACK_RATES, usePlayer } from "../../stores/playerStore";
import { capabilitiesFor } from "../../player/capabilities";
import { useDuration, usePlaybackTime } from "../../player/clock";
import { Cover } from "../ui/Cover";
import { IconButton } from "../ui/IconButton";
import { Slider } from "../ui/Slider";

/**
 * Persistent transport: the design system's floating sticky bar. Parchment at
 * 82% with a 20px blur, one hairline on top, no shadow — the only elevation on
 * screen belongs to the artwork.
 */
export function PlayerBar() {
  const current = usePlayer((state) => state.current);
  const playing = usePlayer((state) => state.playing);
  const volume = usePlayer((state) => state.volume);
  const muted = usePlayer((state) => state.muted);
  const rate = usePlayer((state) => state.rate);
  const repeat = usePlayer((state) => state.repeat);
  const loop = usePlayer((state) => state.loop);
  const error = usePlayer((state) => state.error);
  const toggle = usePlayer((state) => state.toggle);
  const next = usePlayer((state) => state.next);
  const previous = usePlayer((state) => state.previous);
  const seekMs = usePlayer((state) => state.seekMs);
  const setVolume = usePlayer((state) => state.setVolume);
  const toggleMute = usePlayer((state) => state.toggleMute);
  const setRate = usePlayer((state) => state.setRate);
  const setRepeat = usePlayer((state) => state.setRepeat);

  const toggleFavorite = useApp((state) => state.toggleFavorite);
  const setImmersive = useApp((state) => state.setImmersive);

  const timeMs = usePlaybackTime();
  const durationMs = useDuration();
  const totalMs = durationMs || current?.durationMs || 0;
  const caps = capabilitiesFor(current);

  const cycleRate = () => {
    const index = PLAYBACK_RATES.indexOf(rate);
    setRate(PLAYBACK_RATES[(index + 1) % PLAYBACK_RATES.length] ?? 1);
  };

  const cycleRepeat = () => {
    setRepeat(repeat === "off" ? "all" : repeat === "all" ? "one" : "off");
  };

  return (
    <footer className="frosted shrink-0 border-t border-line px-6 py-3">
      <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)_minmax(0,1fr)] items-center gap-6">
        {/* Now playing */}
        <div className="flex min-w-0 items-center gap-3">
          {current ? (
            <>
              <Cover title={current.title} src={current.artworkUrl} size={48} />
              <div className="min-w-0">
                <p className="truncate text-[13.5px] font-semibold tracking-[-0.2px] text-ink">{current.title}</p>
                <p className="truncate text-[12px] text-muted">
                  {current.artist || "Artista desconocido"}
                  {!current.playable ? " · sin reproducción" : ""}
                </p>
              </div>
              <IconButton
                label={current.favorite ? "Quitar de favoritos" : "Guardar en favoritos"}
                size="sm"
                onClick={() => void toggleFavorite(current)}
                className={cn("ml-1", current.favorite && "text-accent hover:text-accent")}
              >
                <Heart className={cn("h-4 w-4", current.favorite && "fill-current")} />
              </IconButton>
            </>
          ) : (
            <p className="text-[13px] text-faint">Elige una canción para empezar</p>
          )}
        </div>

        {/* Transport */}
        <div className="flex flex-col items-center gap-1.5">
          <div className="flex items-center gap-1">
            <IconButton label="Anterior" onClick={previous} disabled={!current}>
              <SkipBack className="h-4 w-4" />
            </IconButton>
            <button
              type="button"
              onClick={toggle}
              disabled={!current}
              aria-label={playing ? "Pausar" : "Reproducir"}
              className={cn(
                "mx-1 flex h-10 w-10 items-center justify-center rounded-full transition-[background-color,transform]",
                "bg-ink text-white hover:bg-black active:scale-[0.95] disabled:opacity-40",
              )}
            >
              {playing ? <Pause className="h-4 w-4 fill-current" /> : <Play className="ml-0.5 h-4 w-4 fill-current" />}
            </button>
            <IconButton label="Siguiente" onClick={next} disabled={!current}>
              <SkipForward className="h-4 w-4" />
            </IconButton>
            <IconButton
              label={
                repeat === "one"
                  ? "Repetir una: activo"
                  : repeat === "all"
                    ? "Repetir todo: activo"
                    : "Repetir: desactivado"
              }
              active={repeat !== "off"}
              onClick={cycleRepeat}
              size="sm"
            >
              {repeat === "one" ? <Repeat1 className="h-3.5 w-3.5" /> : <Repeat className="h-3.5 w-3.5" />}
            </IconButton>
          </div>

          <div className="flex w-full items-center gap-2">
            <span className="w-9 text-right text-[11px] text-faint tnum">{formatTime(timeMs)}</span>
            <Slider
              label="Progreso de la canción"
              value={Math.min(timeMs, totalMs)}
              max={totalMs}
              onChange={seekMs}
              format={(value) => `${formatTime(value)} de ${formatTime(totalMs)}`}
              className="flex-1"
              disabled={!current?.playable}
            />
            <span className="w-9 text-[11px] text-faint tnum">{formatTime(totalMs)}</span>
          </div>
          {error ? (
            <p className="max-w-[52ch] truncate text-[11px] text-danger" role="status">
              {error}
            </p>
          ) : loop ? (
            <p className="max-w-[52ch] truncate text-[11px] text-muted">Repitiendo: {loop.label}</p>
          ) : null}
        </div>

        {/* Output controls */}
        <div className="flex items-center justify-end gap-2">
          <button
            type="button"
            onClick={cycleRate}
            disabled={!caps.rate}
            aria-label={`Velocidad de reproducción: ${rate}x. Cambiar`}
            title={caps.rate ? "Cambiar velocidad" : "Esta fuente no controla la velocidad"}
            className="rounded-full px-2.5 py-1 text-[12px] font-medium text-muted tnum transition-colors hover:bg-surface hover:text-ink disabled:opacity-40"
          >
            {rate}x
          </button>
          <IconButton label={muted ? "Activar sonido" : "Silenciar"} size="sm" onClick={toggleMute}>
            {muted || volume === 0 ? (
              <VolumeX className="h-4 w-4" />
            ) : volume < 0.5 ? (
              <Volume1 className="h-4 w-4" />
            ) : (
              <Volume2 className="h-4 w-4" />
            )}
          </IconButton>
          <Slider
            label="Volumen"
            value={muted ? 0 : Math.round(volume * 100)}
            max={100}
            onChange={(value) => setVolume(value / 100)}
            format={(value) => `${value}%`}
            className="w-[86px]"
          />
          <IconButton label="Abrir modo inmersivo" onClick={() => setImmersive(true)} disabled={!current}>
            <Maximize2 className="h-4 w-4" />
          </IconButton>
        </div>
      </div>
    </footer>
  );
}
