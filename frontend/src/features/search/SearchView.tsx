import { useApp } from "../../stores/appStore";
import { EmptyState } from "../../components/ui/EmptyState";
import { TopBar } from "../../components/layout/TopBar";
import { TrackList } from "../../components/track/TrackList";

/**
 * Search results. The field itself lives in the top bar, so this view only
 * reports what the query found and which providers answered.
 */
export function SearchView() {
  const searching = useApp((state) => state.searching);
  const result = useApp((state) => state.searchResult);
  const providers = useApp((state) => state.providers);
  const openPicker = useApp((state) => state.openPicker);

  const items = result?.items ?? [];

  return (
    <div className="min-h-0">
      <TopBar title="Buscar" subtitle="Tu biblioteca local y el catalogo experimental" />

      <div className="px-7 pb-10">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-faint">
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
              description="Escribe el titulo o el artista en la barra de arriba. La busqueda consulta tu biblioteca y el catalogo de YouTube Music si la música en linea esta activa."
            />
          )}
        </div>
      </div>
    </div>
  );
}
