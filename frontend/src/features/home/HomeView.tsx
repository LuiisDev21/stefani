import { ArrowRight, Headphones, Play, Upload } from "lucide-react";
import { formatStudyTime, formatTime } from "../../lib/format";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { Button } from "../../components/ui/Button";
import { Cover } from "../../components/ui/Cover";
import { EmptyState } from "../../components/ui/EmptyState";
import { SectionTitle } from "../../components/ui/SectionTitle";
import { TrackList } from "../../components/track/TrackList";
import { TopBar } from "../../components/layout/TopBar";

/**
 * Home: a full-bleed parchment tile for the session you can start right now,
 * then the two lists that matter — study progress and your listening history.
 * No metric grid, no decorative glow: the artwork is the only elevation.
 */
export function HomeView() {
  const demo = useApp((state) => state.demo);
  const recently = useApp((state) => state.recently);
  const progress = useApp((state) => state.progress);
  const playlists = useApp((state) => state.playlists);
  const counts = useApp((state) => state.counts);
  const openTrack = useApp((state) => state.openTrack);
  const go = useApp((state) => state.go);
  const importAudio = useApp((state) => state.importAudio);
  const openPicker = useApp((state) => state.openPicker);
  const current = usePlayer((state) => state.current);

  const featured = demo[0];
  const totalStudy = progress.reduce((sum, item) => sum + item.studySeconds, 0);

  return (
    <div className="min-h-0">
      <TopBar
        title="Inicio"
        subtitle={
          counts
            ? `${counts.tracks} pistas en la app · ${counts.annotated} anotaciones de estudio`
            : "Preparando tu biblioteca"
        }
      />

      <div className="px-8 pb-24">
        {featured ? (
          <section className="grid items-center gap-10 rounded-[18px] bg-elevated px-10 py-9 lg:grid-cols-[minmax(0,1fr)_auto]">
            <div className="min-w-0">
              <p className="text-[12px] font-semibold tracking-[-0.12px] text-accent">Sesión de estudio lista</p>
              <h2 className="mt-3 max-w-[18ch] text-[40px] font-semibold leading-[1.1] tracking-[-0.28px] text-ink">
                {featured.title}
              </h2>
              <p className="mt-2 text-[21px] font-normal leading-[1.19] text-muted">
                {featured.artist} · {formatTime(featured.durationMs)}
              </p>
              <p className="mt-4 max-w-[46ch] text-[22px] font-light leading-[1.35] text-ink/80">
                La letra se ilumina palabra por palabra mientras suena. IPA y lectura en español, sin configurar nada.
              </p>
              <p className="mt-3 max-w-[52ch] text-[14px] leading-[1.43] text-muted">
                Traducción y notas de pronunciación cuando conectes tu proveedor de IA.
              </p>
              <div className="mt-7 flex flex-wrap items-center gap-3">
                <Button
                  variant="primary"
                  size="lg"
                  icon={<Play className="h-4 w-4 fill-current" />}
                  onClick={() => void openTrack(featured, demo)}
                >
                  Empezar a escuchar
                </Button>
                <button
                  type="button"
                  onClick={() => go("learning")}
                  className="text-[15px] text-accent transition-opacity hover:opacity-75"
                >
                  Ver mi progreso
                </button>
              </div>
            </div>
            <Cover
              title={featured.title}
              src={featured.artworkUrl}
              size={204}
              rounded="rounded-[18px]"
              shadow
              className="hidden justify-self-end lg:block"
            />
          </section>
        ) : (
          <EmptyState
            icon={<Headphones className="h-5 w-5" />}
            title="La biblioteca de demostración no está disponible"
            description="Revisa los registros de la app o importa tus propios archivos de audio para empezar."
            actions={
              <Button size="sm" onClick={() => void importAudio()} icon={<Upload className="h-3.5 w-3.5" />}>
                Importar audio
              </Button>
            }
          />
        )}

        <section className="mt-12 grid gap-10 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
          <div>
            <SectionTitle
              title="Continuar"
              hint={totalStudy > 0 ? `${formatStudyTime(totalStudy)} de práctica acumulada` : undefined}
              action={
                progress.length > 0 ? (
                  <button
                    type="button"
                    onClick={() => go("learning")}
                    className="flex items-center gap-1 text-[13px] text-accent transition-opacity hover:opacity-75"
                  >
                    Detalle <ArrowRight className="h-3 w-3" />
                  </button>
                ) : undefined
              }
            />
            {progress.length === 0 ? (
              <div className="rounded-[18px] bg-elevated px-5 py-5">
                <p className="max-w-[48ch] text-[14px] leading-[1.43] text-muted">
                  Todavía no hay sesiones. Cuando escuches una canción, aquí aparecerá el tiempo de práctica y los
                  versos que trabajaste.
                </p>
                <Button size="sm" variant="ghost" className="mt-2 px-0" onClick={() => void openTrack(featured!, demo)}>
                  Empezar con la demo
                </Button>
              </div>
            ) : (
              <ul className="divide-y divide-line-soft">
                {progress.slice(0, 4).map((item) => (
                  <li key={item.trackId} className="flex items-center justify-between gap-3 py-3">
                    <div className="min-w-0">
                      <p className="truncate text-[14px] text-ink">{item.title || item.trackId}</p>
                      <p className="truncate text-[12px] text-faint">
                        {item.artist || "Artista desconocido"} · {formatStudyTime(item.studySeconds)}
                      </p>
                    </div>
                    <span className="shrink-0 text-[12px] text-muted">
                      {item.status === "known" ? "conocida" : "en curso"}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div>
            <SectionTitle
              title="Tus listas"
              action={
                playlists.length > 0 ? (
                  <Button size="sm" variant="ghost" onClick={() => go("playlists")}>
                    Administrar
                  </Button>
                ) : undefined
              }
            />
            {playlists.length === 0 ? (
              <div className="rounded-[18px] bg-elevated px-5 py-5">
                <p className="max-w-[48ch] text-[14px] leading-[1.43] text-muted">
                  Aún no tienes listas. Crea una desde Listas o importa una playlist pública de YouTube Music.
                </p>
                <Button size="sm" variant="ghost" className="mt-2 px-0" onClick={() => go("playlists")}>
                  Ir a Listas
                </Button>
              </div>
            ) : (
              <ul className="grid gap-2">
                {playlists.slice(0, 4).map((playlist) => (
                  <li key={playlist.id}>
                    <button
                      type="button"
                      onClick={() => go("playlists", playlist)}
                      className="flex w-full items-center gap-3 rounded-[18px] px-3 py-2.5 text-left transition-colors hover:bg-elevated"
                    >
                      <Cover title={playlist.name} size={38} />
                      <span className="min-w-0">
                        <span className="block truncate text-[14px] text-ink">{playlist.name}</span>
                        <span className="block text-[12px] text-faint tnum">{playlist.trackCount} canciones</span>
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>

        <div className="hairline my-10" />

        <section>
          <SectionTitle
            title="Canciones de demostración"
            hint="Audio generado por la app; puedes estudiar sin cuenta ni servicios externos"
            action={
              <Button size="sm" variant="ghost" onClick={() => go("search")}>
                Buscar mas
              </Button>
            }
          />
          {demo.length === 0 ? (
            <p className="text-[14px] text-muted">Sin canciones de demostración.</p>
          ) : (
            <TrackList tracks={demo} onAdd={openPicker} />
          )}
        </section>

        {recently.length > 0 ? (
          <section className="mt-10">
            <SectionTitle title="Escuchado hace poco" />
            <TrackList tracks={recently} onAdd={openPicker} />
          </section>
        ) : null}

        {current && !current.playable ? <p className="mt-8 text-[12px] text-faint">{current.note}</p> : null}
      </div>
    </div>
  );
}
