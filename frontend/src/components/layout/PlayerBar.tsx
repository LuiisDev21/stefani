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
 * Persistent transport bar. It is the only element with an elevation shadow:
 * everything else on screen is flat, so the bar reads as a separate plane that
 * stays reachable from any view.
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
    <footer className="border-t border-line/80 bg-elevated/85 px-4 py-2.5 backdrop-blur-md shadow-[0_-14px_30px_rgba(3,5,10,0.45)]">
      <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)_minmax(0,1fr)] items-center gap-4">
        {/* Now playing */}
        <div className="flex min-w-0 items-center gap-3">
          {current ? (
            <>
              <Cover title={current.title} src={current.artworkUrl} size={44} />
              <div className="min-w-0">
                <p className="truncate text-[13px] font-medium text-ink">{current.title}</p>
                <p className="truncate text-[11.5px] text-muted">
                  {current.artist || "Artista desconocido"}
                  {!current.playable ? " · sin reproduccion" : ""}
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
            <p className="text-[12px] text-faint">Elige una cancion para empezar</p>
          )}
        </div>

        {/* Transport */}
        <div className="flex flex-col items-center gap-1">
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
                "mx-1 flex h-9 w-9 items-center justify-center rounded-full transition-colors",
                "bg-ink text-[#0b0e16] hover:bg-white disabled:opacity-40",
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
            <span className="w-9 text-right text-[10.5px] text-faint tnum">{formatTime(timeMs)}</span>
            <Slider
              label="Progreso de la cancion"
              value={Math.min(timeMs, totalMs)}
              max={totalMs}
              onChange={seekMs}
              format={(value) => `${formatTime(value)} de ${formatTime(totalMs)}`}
              className="flex-1"
              disabled={!current?.playable}
            />
            <span className="w-9 text-[10.5px] text-faint tnum">{formatTime(totalMs)}</span>
          </div>
          {error ? (
            <p className="max-w-[52ch] truncate text-[10.5px] text-danger" role="status">
              {error}
            </p>
          ) : loop ? (
            <p className="max-w-[52ch] truncate text-[10.5px] text-accent/85">
              Repitiendo: {loop.label}
            </p>
          ) : null}
        </div>

        {/* Output controls */}
        <div className="flex items-center justify-end gap-2">
          <button
            type="button"
            onClick={cycleRate}
            disabled={!caps.rate}
            aria-label={`Velocidad de reproduccion: ${rate}x. Cambiar`}
            title={caps.rate ? "Cambiar velocidad" : "Esta fuente no controla la velocidad"}
            className="rounded-[6px] px-2 py-1 text-[11px] font-medium text-muted tnum transition-colors hover:bg-surface hover:text-ink disabled:opacity-40"
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
          <IconButton
            label="Abrir modo inmersivo"
            onClick={() => setImmersive(true)}
            disabled={!current}
          >
            <Maximize2 className="h-4 w-4" />
          </IconButton>
        </div>
      </div>
    </footer>
  );
}
