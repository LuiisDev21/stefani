import { CheckCircle2, KeyRound, PlugZap, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { z } from "zod";
import { useApp } from "../../stores/appStore";
import type { Settings } from "../../lib/types";
import { Button } from "../../components/ui/Button";
import { SectionTitle } from "../../components/ui/SectionTitle";
import { Switch } from "../../components/ui/Switch";
import { TopBar } from "../../components/layout/TopBar";
import { LLM_PRESETS, presetByID, type LLMPreset } from "../../lib/llmPresets";

const llmSchema = z.object({
  provider: z.string().optional().default(""),
  baseUrl: z
    .string()
    .trim()
    .min(1, "La URL base es obligatoria")
    .refine((value) => {
      try {
        const url = new URL(value);
        if (url.protocol === "https:") return true;
        const local = ["localhost", "127.0.0.1", "::1"].includes(url.hostname);
        return url.protocol === "http:" && local;
      } catch {
        return false;
      }
    }, "Usa una URL https, o http solo para localhost"),
  model: z.string().trim().min(1, "Indica el modelo"),
  temperature: z.number().min(0, "Minimo 0").max(2, "Maximo 2"),
  timeoutSeconds: z.number().int().min(5, "Minimo 5 s").max(300, "Maximo 300 s"),
  maxRetries: z.number().int().min(0, "Minimo 0").max(5, "Maximo 5"),
});

/** Settings: AI connection, experimental providers and app information. */
export function SettingsView() {
  const settings = useApp((state) => state.settings);
  const info = useApp((state) => state.info);
  const providers = useApp((state) => state.providers);
  const saveSettings = useApp((state) => state.saveSettings);
  const testLLM = useApp((state) => state.testLLM);
  const clearKey = useApp((state) => state.clearKey);

  const [form, setForm] = useState<Settings | null>(settings);
  const [apiKey, setApiKey] = useState("");
  const [errors, setErrors] = useState<string[]>([]);
  // Models discovered by the provider on /models. Empty until the connection is
  // tested, so the selector always reflects the API instead of a hardcoded list.
  const [models, setModels] = useState<string[]>([]);
  const [modelsNotice, setModelsNotice] = useState("");
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ ok: boolean; message: string; latencyMs: number } | null>(null);

  useEffect(() => {
    setForm(settings);
  }, [settings]);

  if (!form) {
    return (
      <div className="px-7 py-6">
        <p className="text-[12.5px] text-muted">Cargando ajustes…</p>
      </div>
    );
  }

  const update = (patch: Partial<Settings["llm"]>) => {
    setForm({ ...form, llm: { ...form.llm, ...patch } });
  };

  const applyPreset = (preset: LLMPreset | undefined) => {
    // The models belong to the previous endpoint: drop them with the change.
    setModels([]);
    setModelsNotice("");
    setTestResult(null);
    if (!preset) {
      update({ provider: "" });
      return;
    }
    // A preset fills a working endpoint and model; both stay editable.
    update({ provider: preset.id, baseUrl: preset.baseUrl, model: preset.model });
  };

  const activePreset =
    presetByID(form.llm.provider) ??
    LLM_PRESETS.find((preset) => preset.baseUrl.replace(/\/$/, "") === form.llm.baseUrl.replace(/\/$/, ""));

  const persist = async () => {
    const parsed = llmSchema.safeParse(form.llm);
    if (!parsed.success) {
      setErrors(parsed.error.issues.map((issue) => issue.message));
      return false;
    }
    setErrors([]);
    const ok = await saveSettings({
      llm: { ...form.llm, ...parsed.data },
      features: form.features,
      accent: form.accent,
      apiKey,
    });
    if (ok) setApiKey("");
    return ok;
  };

  const submit = async () => {
    await persist();
  };

  const runTest = async () => {
    setTesting(true);
    try {
      // The backend tests what is saved, so an edited endpoint is saved first:
      // "probar" must test what the screen shows, not the previous value.
      const saved = await persist();
      if (!saved) return;
      const outcome = await testLLM();
      if (!outcome) return;
      setTestResult({ ok: outcome.ok, message: outcome.message, latencyMs: outcome.latencyMs });
      if (outcome.ok) {
        setModels(outcome.models ?? []);
        setModelsNotice(outcome.modelsNotice ?? "");
      }
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="min-h-0">
      <TopBar
        title="Ajustes"
        subtitle="Conexión de IA, música en línea y datos locales"
        actions={
          <Button size="sm" variant="primary" onClick={() => void submit()}>
            Guardar cambios
          </Button>
        }
      />

      <div className="mx-auto max-w-[820px] px-8 pb-24">
        <section>
          <SectionTitle
            title="Proveedor de IA"
            hint="Elige un proveedor para rellenar el endpoint, prueba la conexión para descubrir sus modelos y guarda. La clave vive en el almacén de credenciales del sistema."
          />
          <div className="mt-4">
            <p className="text-[12px] font-semibold text-ink">Proveedor</p>
            <div className="mt-2 flex flex-wrap gap-2" role="group" aria-label="Presets de proveedor de IA">
              {LLM_PRESETS.map((preset) => {
                const active = activePreset?.id === preset.id;
                return (
                  <button
                    key={preset.id}
                    type="button"
                    aria-pressed={active}
                    onClick={() => applyPreset(preset)}
                    className={
                      active
                        ? "rounded-full border border-accent bg-accent/6 px-3.5 py-1.5 text-[13px] text-accent"
                        : "rounded-full border border-line px-3.5 py-1.5 text-[13px] text-muted transition-colors hover:bg-elevated hover:text-ink"
                    }
                  >
                    {preset.name}
                  </button>
                );
              })}
              <button
                type="button"
                aria-pressed={!activePreset}
                onClick={() => applyPreset(undefined)}
                className={
                  !activePreset
                    ? "rounded-full border border-accent bg-accent/6 px-3.5 py-1.5 text-[13px] text-accent"
                    : "rounded-full border border-line px-3.5 py-1.5 text-[13px] text-muted transition-colors hover:bg-elevated hover:text-ink"
                }
              >
                Personalizado
              </button>
            </div>
            {activePreset ? (
              <div className="mt-3 rounded-[14px] bg-elevated px-4 py-3">
                <p className="text-[13px] leading-[1.43] text-muted">{activePreset.note}</p>
                <p className="mt-1.5 text-[12px] text-faint">{activePreset.docsUrl}</p>
                {activePreset.session ? (
                  <p className="mt-1.5 text-[12px] text-accent">
                    Sesión automática: cada petición lleva x-opencode-session con un id estable.
                  </p>
                ) : null}
              </div>
            ) : null}
          </div>

          <div className="mt-4 grid gap-3 sm:grid-cols-2">
            <Field label="URL base">
              <input
                value={form.llm.baseUrl}
                onChange={(event) => update({ baseUrl: event.target.value })}
                placeholder="https://api.openai.com/v1"
                className={inputClass}
                spellCheck={false}
              />
            </Field>
            <Field label="Modelo">
              {models.length > 0 ? (
                <select
                  value={form.llm.model}
                  onChange={(event) => update({ model: event.target.value })}
                  className={inputClass}
                  aria-label="Modelo disponible en el proveedor"
                >
                  <option value="" disabled>
                    Elegir modelo…
                  </option>
                  {Array.from(new Set([...(form.llm.model ? [form.llm.model] : []), ...models])).map((model) => (
                    <option key={model} value={model}>
                      {model === form.llm.model ? `${model} (actual)` : model}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  list="llm-model-options"
                  value={form.llm.model}
                  onChange={(event) => update({ model: event.target.value })}
                  placeholder={activePreset?.model ?? "nombre-del-modelo"}
                  className={inputClass}
                  spellCheck={false}
                />
              )}
              <datalist id="llm-model-options">
                {(activePreset?.models ?? []).map((model) => (
                  <option key={model} value={model} />
                ))}
              </datalist>
              {models.length > 0 ? (
                <p className="mt-1 text-[11px] text-faint">{models.length} modelos descubiertos. Elige uno y pulsa Guardar cambios.</p>
              ) : modelsNotice ? (
                <p className="mt-1 text-[11px] text-faint">{modelsNotice}</p>
              ) : (
                <p className="mt-1 text-[11px] text-faint">
                  Prueba la conexión para cargar los modelos reales del proveedor.
                </p>
              )}
            </Field>
            <Field label="Temperatura">
              <input
                type="number"
                step="0.1"
                min={0}
                max={2}
                value={form.llm.temperature}
                onChange={(event) => update({ temperature: Number(event.target.value) })}
                className={inputClass}
              />
            </Field>
            <Field label="Timeout (segundos)">
              <input
                type="number"
                min={5}
                max={300}
                value={form.llm.timeoutSeconds}
                onChange={(event) => update({ timeoutSeconds: Number(event.target.value) })}
                className={inputClass}
              />
            </Field>
            <Field label="Reintentos">
              <input
                type="number"
                min={0}
                max={5}
                value={form.llm.maxRetries}
                onChange={(event) => update({ maxRetries: Number(event.target.value) })}
                className={inputClass}
              />
            </Field>
          </div>

          <div className="mt-4 border-t border-line-soft pt-4">
            <Switch
              checked={form.llm.jsonMode}
              onChange={(value) => update({ jsonMode: value })}
              label="Pedir respuesta JSON al proveedor"
              description="Recomendado: la app valida la respuesta contra el esquema de anotación."
            />
          </div>

          <div className="mt-4 grid gap-3 sm:grid-cols-2">
            <Field label={form.hasKey ? "Clave guardada (escribe otra para reemplazarla)" : "Clave de API"}>
              <div className="relative">
                <KeyRound className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint" />
                <input
                  type="password"
                  value={apiKey}
                  autoComplete="off"
                  onChange={(event) => setApiKey(event.target.value)}
                  placeholder={form.hasKey ? "••••••••••••" : "sk-..."}
                  className={`${inputClass} pl-9`}
                />
              </div>
            </Field>
            <Field label="Variedad de inglés">
              <select
                value={form.accent}
                onChange={(event) => setForm({ ...form, accent: event.target.value })}
                className={inputClass}
              >
                <option value="en-US">Inglés estadounidense (en-US)</option>
                <option value="en-GB">Inglés britanico (en-GB)</option>
              </select>
            </Field>
          </div>

          {errors.length > 0 ? (
            <ul className="mt-3 space-y-1" role="alert">
              {errors.map((message) => (
                <li key={message} className="text-[12px] text-danger">
                  {message}
                </li>
              ))}
            </ul>
          ) : null}

          {form.keyStoreError ? (
            <p className="mt-3 rounded-[7px] border border-danger/30 bg-danger/5 px-3 py-2 text-[11.5px] text-danger">
              {form.keyStoreError}
            </p>
          ) : null}

          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Button
              size="sm"
              variant="subtle"
              icon={<PlugZap className="h-3.5 w-3.5" />}
              disabled={testing}
              onClick={() => void runTest()}
            >
              {testing ? "Probando…" : "Probar y buscar modelos"}
            </Button>
            <Button
              size="sm"
              variant="danger"
              icon={<Trash2 className="h-3.5 w-3.5" />}
              disabled={!form.hasKey}
              onClick={() => void clearKey()}
            >
              Eliminar clave
            </Button>
            <span className="text-[11px] text-faint">
              {form.hasKey ? "Hay una clave guardada en el sistema" : "Sin clave guardada"}
            </span>
          </div>
          {testResult ? (
            <p
              role="status"
              className={
                testResult.ok
                  ? "mt-2 flex items-center gap-2 text-[13px] text-ink"
                  : "mt-2 flex items-center gap-2 text-[13px] text-danger"
              }
            >
              <CheckCircle2 className={testResult.ok ? "h-4 w-4 text-accent" : "h-4 w-4"} aria-hidden="true" />
              {testResult.message}
              <span className="text-[12px] text-faint tnum">{testResult.latencyMs} ms</span>
            </p>
          ) : null}
        </section>

        <div className="hairline my-8" />

        <section>
          <SectionTitle
            title="Música en linea"
            hint="Catalogo y reproducción de YouTube Music mediante yt-dlp, una herramienta local y no oficial. Uso personal: puede dejar de funcionar si YouTube cambia, y va contra sus terminos de servicio."
          />
          <Switch
            checked={form.features.youtubeMusic}
            onChange={(value) => setForm({ ...form, features: { ...form.features, youtubeMusic: value } })}
            label="Buscar y reproducir desde YouTube Music"
            description="La búsqueda usa el catálogo público. El audio se resuelve en esta máquina y se reproduce dentro de la app, con letra, velocidad y repetición."
          />

          <div className="mt-3 flex items-start gap-2 rounded-[9px] border border-line/70 px-3 py-2.5">
            <span
              className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full"
              style={{ background: info?.online.available ? "#0066cc" : "#ff9b9b" }}
              aria-hidden
            />
            <p className="text-[11.5px] leading-relaxed text-muted">
              {info?.online.available ? (
                <>
                  Reproductor en linea listo. <span className="text-ink">{info.online.detail}</span>
                </>
              ) : (
                <>
                  <span className="text-ink">Falta yt-dlp.</span> Instalalo con{" "}
                  <span className="font-mono text-[11px] text-ink">winget install yt-dlp</span> y reinicia STEFANI. Sin
                  el, las pistas de YouTube Music se pueden guardar y abrir en la app oficial, pero no suenan aquí.
                </>
              )}
            </p>
          </div>
        </section>

        <div className="hairline my-8" />

        <section>
          <SectionTitle title="Proveedores" />
          <ul className="space-y-1.5">
            {providers.map((provider) => (
              <li key={provider.name} className="flex items-start gap-2 text-[12px]">
                <CheckCircle2
                  className="mt-0.5 h-3.5 w-3.5 shrink-0"
                  style={{ color: provider.state === "disabled" ? "#8e8e93" : "#0066cc" }}
                />
                <span>
                  <span className="text-ink">{provider.name}</span>
                  <span className="text-muted"> · {provider.detail}</span>
                </span>
              </li>
            ))}
          </ul>
        </section>

        <div className="hairline my-8" />

        <section>
          <SectionTitle title="Acerca de" />
          <dl className="space-y-1 text-[11.5px] text-muted">
            <Row label="Versión" value={info ? `${info.name} ${info.version}` : "—"} />
            <Row label="Identificador" value={info?.identifier ?? "—"} />
            <Row label="Base de datos" value={info?.dbPath ?? "—"} mono />
            <Row label="Esquema" value={info ? `v${info.schema}` : "—"} />
            <Row label="Configuración" value={form.configPath} mono />
          </dl>
        </section>
      </div>
    </div>
  );
}

const inputClass =
  "h-9 w-full rounded-[7px] border border-line bg-bg px-3 text-[12.5px] text-ink placeholder:text-faint focus:border-accent/60";

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-[11px] text-faint">{label}</span>
      {children}
    </label>
  );
}

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex gap-3">
      <dt className="w-28 shrink-0 text-faint">{label}</dt>
      <dd className={`min-w-0 break-all text-muted ${mono ? "font-mono text-[11px]" : ""}`}>{value}</dd>
    </div>
  );
}
