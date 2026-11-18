import { LoaderCircle, Search } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useApp } from "../../stores/appStore";
import { EmptyState } from "../../components/ui/EmptyState";
import { TopBar } from "../../components/layout/TopBar";
import { TrackList } from "../../components/track/TrackList";

/** Search across the local library and, when enabled, the YouTube Music catalog. */
export function SearchView() {
  const search = useApp((state) => state.search);
  const searching = useApp((state) => state.searching);
  const result = useApp((state) => state.searchResult);
  const query = useApp((state) => state.searchQuery);
  const providers = useApp((state) => state.providers);
  const openPicker = useApp((state) => state.openPicker);

  const [text, setText] = useState(query);
  const inputRef = useRef<HTMLInputElement>(null);
  const timer = useRef<number | null>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  useEffect(() => {
    return () => {
      if (timer.current) window.clearTimeout(timer.current);
    };
  }, []);

  const schedule = (value: string) => {
    setText(value);
    if (timer.current) window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => void search(value), 350);
  };

  const items = result?.items ?? [];

  return (
    <div className="min-h-0">
      <TopBar title="Buscar" subtitle="Tu biblioteca local y el catalogo experimental" />

      <div className="px-7 pb-10">
        <form
          role="search"
          onSubmit={(event) => {
            event.preventDefault();
            void search(text);
          }}
          className="relative max-w-[560px]"
        >
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-faint" aria-hidden="true" />
          <input
            ref={inputRef}
            type="search"
            value={text}
            onChange={(event) => schedule(event.target.value)}
            placeholder="Canción, artista o álbum"
            aria-label="Buscar canciones"
            className="h-10 w-full rounded-[8px] border border-line bg-elevated pl-9 pr-10 text-[13px] text-ink placeholder:text-faint focus:border-line-strong"
          />
          {searching ? (
            <LoaderCircle className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin text-faint" aria-hidden="true" />
          ) : null}
        </form>

        <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-faint">
          {providers.map((provider) => (
            <span key={provider.name} className="flex items-center gap-1.5" title={provider.detail}>
              <span
                className="h-1.5 w-1.5 rounded-full"
                style={{
                  background:
                    provider.state === "ready" || provider.state === "enabled"
                      ? "var(--color-ink)"
                      : "var(--color-ash)",
                }}
              />
              {provider.name}
            </span>
          ))}
        </div>

        {result?.notice ? (
          <p className="mt-4 max-w-[70ch] rounded-[8px] border border-line/70 bg-elevated/50 px-3 py-2 text-[11.5px] leading-relaxed text-muted">
            {result.notice}
          </p>
        ) : null}

        <div className="mt-5">
          {searching && items.length === 0 ? (
            <div className="space-y-2" aria-busy="true">
              {[0, 1, 2, 3, 4].map((index) => (
                <div key={index} className="h-12 animate-pulse rounded-[8px] bg-surface/50" />
              ))}
            </div>
          ) : items.length > 0 ? (
            <TrackList tracks={items} onAdd={openPicker} />
          ) : result && result.query ? (
            <EmptyState
              title="Sin resultados"
              description={`No encontre nada para "${result.query}". Prueba con otra grafia o revisa la música en linea en Ajustes.`}
            />
          ) : (
            <EmptyState
              title="Busca en tu música"
              description="Escribe el titulo o el artista. La busqueda consulta tu biblioteca y el catalogo de YouTube Music si la música en linea esta activa."
            />
          )}
        </div>
      </div>
    </div>
  );
}
