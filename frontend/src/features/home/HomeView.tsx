import { useEffect } from "react";
import { ArrowRight, Headphones, Heart, Library, ListMusic, Pause, Play, Search, Upload } from "lucide-react";
import { formatTime, plural } from "../../lib/format";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { Button } from "../../components/ui/Button";
import { Cover } from "../../components/ui/Cover";
import { EmptyState } from "../../components/ui/EmptyState";
import { SectionTitle } from "../../components/ui/SectionTitle";
import { TrackList } from "../../components/track/TrackList";
import { TopBar } from "../../components/layout/TopBar";
import { MediaCard } from "../../components/home/MediaCard";
import { QuickCard } from "../../components/home/QuickCard";
import { ShelfRow } from "../../components/home/ShelfRow";

/**
 * Home: the record that is playing — or the first one in the library — owns the
 * top of the page with its cover and the one white action. Below it: shortcuts,
 * the library, your lists, and what you played last. Every number comes from
 * the database.
 */
export function HomeView() {
  const library = useApp((state) => state.library);
  const recently = useApp((state) => state.recently);
  const playlists = useApp((state) => state.playlists);
  const counts = useApp((state) => state.counts);
  const openTrack = useApp((state) => state.openTrack);
  const go = useApp((state) => state.go);
  const importAudio = useApp((state) => state.importAudio);
  const openPicker = useApp((state) => state.openPicker);
  const focusSearch = useApp((state) => state.focusSearch);
  const homeFeed = useApp((state) => state.homeFeed);
  const loadHomeFeed = useApp((state) => state.loadHomeFeed);

  const current = usePlayer((state) => state.current);
  const playing = usePlayer((state) => state.playing);
  const toggle = usePlayer((state) => state.toggle);

  // The personalised rows come from the signed-in account, so they load after
  // the first paint and never block the library.
  useEffect(() => {
    void loadHomeFeed();
  }, [loadHomeFeed]);

  const hero = current ?? library[0] ?? recently[0];
  const heroIsCurrent = Boolean(current && hero && current.id === hero.id);
  const heroAction = !hero ? undefined : heroIsCurrent ? "toggle" : "open";
  const shelf = library.slice(0, 8);

  return (
    <div className="min-h-0">
      <TopBar title="Inicio" />

      <div className="px-8 pb-28">
        {hero ? (
          <section
            aria-label={heroIsCurrent ? "Sonando ahora" : "Tu biblioteca"}
            className="grid gap-7 rounded-[16px] border border-line bg-elevated p-6 lg:grid-cols-[220px_minmax(0,1fr)_200px] lg:items-center lg:p-8"
          >
            <Cover
              title={hero.title}
              artist={hero.artist}
              src={hero.artworkUrl}
              size={220}
              rounded="rounded-[10px]"
              className="hidden lg:block"
            />

            <div className="min-w-0">
              <p className="text-[12px] text-muted">
                {heroIsCurrent ? (playing ? "Sonando ahora" : "En pausa") : "De tu biblioteca"}
              </p>
              <h2 className="mt-2 text-[36px] font-semibold leading-[1.04] tracking-[-0.04em] text-ink sm:text-[44px]">
                {hero.title}
              </h2>
              <p className="mt-2 truncate text-[14px] text-body">
                {hero.artist || "Artista desconocido"}
                {hero.album ? ` · ${hero.album}` : ""}
                {hero.durationMs ? ` · ${formatTime(hero.durationMs)}` : ""}
              </p>
              <p className="mt-4 max-w-[54ch] text-[14px] leading-[1.5] text-muted">
                La letra se ilumina palabra por palabra mientras suena, con IPA y lectura en español.
              </p>
              <div className="mt-6 flex flex-wrap items-center gap-2">
                <Button
                  variant="primary"
                  icon={
                    heroAction === "toggle" && playing ? (
                      <Pause className="h-4 w-4 fill-current" />
                    ) : (
                      <Play className="h-4 w-4 fill-current" />
                    )
                  }
                  onClick={() => {
                    if (!hero) return;
                    if (heroAction === "toggle") toggle();
                    else void openTrack(hero, library);
                  }}
                >
                  {heroAction === "toggle" ? (playing ? "Pausar" : "Reanudar") : "Reproducir"}
                </Button>
              </div>
            </div>

            <dl className="hidden shrink-0 flex-col gap-5 border-l border-line pl-8 lg:flex">
              <div>
                <dt className="text-[12px] text-muted">Pistas</dt>
                <dd className="mt-1 text-[24px] font-semibold tracking-[-0.03em] text-ink tnum">
                  {counts?.tracks ?? 0}
                </dd>
              </div>
              <div>
                <dt className="text-[12px] text-muted">Favoritas</dt>
                <dd className="mt-1 text-[24px] font-semibold tracking-[-0.03em] text-ink tnum">
                  {counts?.favorites ?? 0}
                </dd>
              </div>
              <div>
                <dt className="text-[12px] text-muted">Listas</dt>
                <dd className="mt-1 text-[24px] font-semibold tracking-[-0.03em] text-ink tnum">
                  {counts?.playlists ?? 0}
                </dd>
              </div>
            </dl>
          </section>
        ) : (
          <EmptyState
            icon={<Headphones className="h-5 w-5" />}
            title="Todavía no guardaste nada"
            description="Importa tus archivos de audio o busca en el catálogo y guarda lo que quieras escuchar."
            actions={
              <Button size="sm" onClick={() => void importAudio()} icon={<Upload className="h-3.5 w-3.5" />}>
                Importar audio
              </Button>
            }
          />
        )}

        {homeFeed.map((shelf) => (
          <ShelfRow key={shelf.title} shelf={shelf} />
        ))}

        <section className="mt-10" aria-label="Accesos rápidos">
          <SectionTitle title="Accesos rápidos" />
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <QuickCard
              icon={<Library className="h-4 w-4" />}
              label="Biblioteca"
              detail={plural(counts?.libraryTracks ?? 0, "pista", "pistas")}
              onClick={() => go("library")}
            />
            <QuickCard
              icon={<Heart className="h-4 w-4" />}
              label="Favoritos"
              detail={plural(counts?.favorites ?? 0, "canción", "canciones")}
              onClick={() => go("favorites")}
            />
            <QuickCard
              icon={<ListMusic className="h-4 w-4" />}
              label="Listas"
              detail={plural(counts?.playlists ?? 0, "lista", "listas")}
              onClick={() => go("playlists")}
            />
            <QuickCard
              icon={<Search className="h-4 w-4" />}
              label="Buscar"
              detail="Biblioteca y catálogo"
              onClick={() => focusSearch()}
            />
          </div>
        </section>

        <section className="mt-12 grid gap-10 lg:grid-cols-[minmax(0,1.25fr)_minmax(0,0.75fr)]">
          <div>
            <SectionTitle
              title="Tu biblioteca"
              action={
                <Button size="sm" variant="ghost" onClick={() => go("library")}>
                  Ver todo
                </Button>
              }
            />
            {shelf.length === 0 ? (
              <p className="text-[14px] text-muted">Todavía no hay canciones guardadas.</p>
            ) : (
              <TrackList tracks={shelf} onAdd={openPicker} />
            )}
          </div>

          <div>
            <SectionTitle
              title="Tus listas"
              action={
                playlists.length > 0 ? (
                  <button
                    type="button"
                    onClick={() => go("playlists")}
                    className="flex items-center gap-1 text-[13px] text-body transition-colors hover:text-ink"
                  >
                    Administrar <ArrowRight className="h-3 w-3" />
                  </button>
                ) : undefined
              }
            />
            {playlists.length === 0 ? (
              <div className="rounded-[10px] border border-line bg-elevated px-4 py-4">
                <p className="text-[13px] leading-[1.5] text-muted">
                  Aún no tienes listas. Crea una desde Listas o importa una playlist pública de YouTube Music.
                </p>
                <Button size="sm" variant="ghost" className="mt-2 px-0" onClick={() => go("playlists")}>
                  Ir a Listas
                </Button>
              </div>
            ) : (
              <ul className="space-y-1">
                {playlists.slice(0, 5).map((playlist) => (
                  <li key={playlist.id}>
                    <button
                      type="button"
                      onClick={() => go("playlists", playlist)}
                      className="flex w-full items-center gap-3 rounded-[8px] px-2 py-2 text-left transition-colors hover:bg-surface"
                    >
                      <Cover title={playlist.name} src={playlist.artwork} size={36} rounded="rounded-[6px]" />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-[13px] text-ink">{playlist.name}</span>
                        <span className="block truncate text-[12px] text-ash tnum">
                          {plural(playlist.trackCount, "canción", "canciones")}
                        </span>
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>

        {recently.length > 0 ? (
          <section className="mt-12" aria-label="Escuchado hace poco">
            <SectionTitle
              title="Escuchado hace poco"
              action={
                <Button size="sm" variant="ghost" onClick={() => go("library")}>
                  Ver biblioteca
                </Button>
              }
            />
            <div className="grid grid-cols-2 gap-5 sm:grid-cols-3 xl:grid-cols-5">
              {recently.slice(0, 5).map((track) => (
                <MediaCard key={track.id} track={track} onPlay={() => void openTrack(track, recently)} />
              ))}
            </div>
          </section>
        ) : null}
      </div>
    </div>
  );
}
