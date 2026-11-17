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
 * Home: one focal band (the demo session you can start right now), then the two
 * lists that matter, listening history and study progress. No metric grid.
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

      <div className="px-7 pb-10">
        <section className="grid gap-5 lg:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]">
          {featured ? (
            <article className="relative overflow-hidden rounded-[16px] border border-line bg-elevated p-6">
              <div
                aria-hidden="true"
                className="pointer-events-none absolute -right-16 -top-24 h-[280px] w-[280px] rounded-full opacity-40 blur-3xl"
                style={{ background: "radial-gradient(circle, #2b5aa6 0%, #171d2c 70%)" }}
              />
              <div className="relative flex items-start gap-5">
                <Cover title={featured.title} src={featured.artworkUrl} size={104} rounded="rounded-[10px]" />
                <div className="min-w-0">
                  <p className="text-[11px] font-medium text-accent">Sesion de estudio lista</p>
                  <h2 className="mt-2 text-[25px] font-medium leading-tight tracking-[-0.02em] text-ink">
                    {featured.title}
                  </h2>
                  <p className="mt-1 text-[12.5px] text-muted">
                    {featured.artist} · {formatTime(featured.durationMs)} · letra original con tiempos por palabra
                  </p>
                  <p className="mt-3 max-w-[46ch] text-[12.5px] leading-relaxed text-muted">
                    Audio sintetizado para la demo, IPA, aproximacion al espanol y traduccion ya calculadas. Sirve
                    para probar el flujo completo sin configurar nada.
                  </p>
                  <div className="mt-4 flex flex-wrap items-center gap-2">
                    <Button
                      variant="primary"
                      icon={<Play className="h-3.5 w-3.5 fill-current" />}
                      onClick={() => void openTrack(featured, demo)}
                    >
                      Empezar a escuchar
                    </Button>
                    <Button variant="ghost" onClick={() => go("learning")}>
                      Ver mi progreso
                    </Button>
                  </div>
                </div>
              </div>
            </article>
          ) : (
            <EmptyState
              icon={<Headphones className="h-5 w-5" />}
              title="La biblioteca de demostracion no esta disponible"
              description="Revisa los registros de la app o importa tus propios archivos de audio para empezar."
              actions={
                <Button size="sm" onClick={() => void importAudio()} icon={<Upload className="h-3.5 w-3.5" />}>
                  Importar audio
                </Button>
              }
            />
          )}

          <div>
            <SectionTitle
              title="Continuar"
              hint={totalStudy > 0 ? `${formatStudyTime(totalStudy)} de practica acumulada` : undefined}
              action={
                progress.length > 0 ? (
                  <button
                    type="button"
                    onClick={() => go("learning")}
                    className="flex items-center gap-1 text-[11px] text-muted transition-colors hover:text-ink"
                  >
                    Detalle <ArrowRight className="h-3 w-3" />
                  </button>
                ) : undefined
              }
            />
            {progress.length === 0 ? (
              <p className="rounded-[10px] border border-dashed border-line/80 px-4 py-5 text-[12.5px] leading-relaxed text-muted">
                Todavia no hay sesiones. Cuando escuches una cancion, aqui aparecera el tiempo de practica y los
                versos que trabajaste.
              </p>
            ) : (
              <ul className="space-y-1">
                {progress.slice(0, 4).map((item) => (
                  <li
                    key={item.trackId}
                    className="flex items-center justify-between gap-3 rounded-[8px] px-2 py-2 transition-colors hover:bg-surface/50"
                  >
                    <div className="min-w-0">
                      <p className="truncate text-[12.5px] text-ink">{item.title || item.trackId}</p>
                      <p className="truncate text-[11px] text-faint">
                        {item.artist || "Artista desconocido"} · {formatStudyTime(item.studySeconds)}
                      </p>
                    </div>
                    <span className="shrink-0 text-[11px] text-faint">
                      {item.status === "known" ? "conocida" : "en curso"}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>

        <div className="signal-line my-7" />

        <section>
          <SectionTitle
            title="Canciones de demostracion"
            hint="Audio generado por la app; puedes estudiar sin cuenta ni servicios externos"
            action={
              <Button size="sm" variant="ghost" onClick={() => go("search")}>
                Buscar mas
              </Button>
            }
          />
          {demo.length === 0 ? (
            <p className="text-[12.5px] text-muted">Sin canciones de demostracion.</p>
          ) : (
            <TrackList tracks={demo} onAdd={openPicker} />
          )}
        </section>

        {recently.length > 0 ? (
          <section className="mt-7">
            <SectionTitle title="Escuchado hace poco" />
            <TrackList tracks={recently} onAdd={openPicker} />
          </section>
        ) : null}

        {playlists.length > 0 ? (
          <section className="mt-7">
            <SectionTitle
              title="Tus listas"
              action={
                <Button size="sm" variant="ghost" onClick={() => go("playlists")}>
                  Administrar
                </Button>
              }
            />
            <ul className="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
              {playlists.slice(0, 6).map((playlist) => (
                <li key={playlist.id}>
                  <button
                    type="button"
                    onClick={() => go("playlists", playlist)}
                    className="flex w-full items-center gap-3 rounded-[9px] border border-line/70 px-3 py-2.5 text-left transition-colors hover:border-line hover:bg-surface/40"
                  >
                    <Cover title={playlist.name} size={34} />
                    <span className="min-w-0">
                      <span className="block truncate text-[12.5px] text-ink">{playlist.name}</span>
                      <span className="block text-[11px] text-faint tnum">{playlist.trackCount} canciones</span>
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          </section>
        ) : null}

        {current && !current.playable ? (
          <p className="mt-6 text-[11.5px] text-faint">{current.note}</p>
        ) : null}
      </div>
    </div>
  );
}
