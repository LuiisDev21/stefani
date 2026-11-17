import { CheckCircle2, KeyRound, PlugZap, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { z } from "zod";
import { useApp } from "../../stores/appStore";
import type { Settings } from "../../lib/types";
import { Button } from "../../components/ui/Button";
import { SectionTitle } from "../../components/ui/SectionTitle";
import { TopBar } from "../../components/layout/TopBar";

const llmSchema = z.object({
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

  const submit = async () => {
    const parsed = llmSchema.safeParse(form.llm);
    if (!parsed.success) {
      setErrors(parsed.error.issues.map((issue) => issue.message));
      return;
    }
    setErrors([]);
    const ok = await saveSettings({
      llm: { ...form.llm, ...parsed.data },
      features: form.features,
      accent: form.accent,
      apiKey,
    });
    if (ok) setApiKey("");
  };

  return (
    <div className="min-h-0">
      <TopBar
        title="Ajustes"
        subtitle="Conexion de IA, catalogo experimental y datos locales"
        actions={
          <Button size="sm" variant="primary" onClick={() => void submit()}>
            Guardar cambios
          </Button>
        }
      />

      <div className="max-w-[720px] px-7 pb-12">
        <section>
          <SectionTitle
            title="Proveedor de IA"
            hint="Cualquier API compatible con OpenAI Chat Completions. La clave se guarda en el almacen de credenciales del sistema, nunca en la base de datos."
          />
          <div className="grid gap-3 sm:grid-cols-2">
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
              <input
                value={form.llm.model}
                onChange={(event) => update({ model: event.target.value })}
                placeholder="gpt-4o-mini"
                className={inputClass}
                spellCheck={false}
              />
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
            <Field label="Modo JSON">
              <label className="flex h-9 items-center gap-2 text-[12.5px] text-muted">
                <input
                  type="checkbox"
                  checked={form.llm.jsonMode}
                  onChange={(event) => update({ jsonMode: event.target.checked })}
                  className="h-4 w-4 accent-[#85afff]"
                />
                Pedir respuesta JSON al proveedor
              </label>
            </Field>
          </div>

          <div className="mt-3 grid gap-3 sm:grid-cols-2">
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
            <Field label="Variedad de ingles">
              <select
                value={form.accent}
                onChange={(event) => setForm({ ...form, accent: event.target.value })}
                className={inputClass}
              >
                <option value="en-US">Ingles estadounidense (en-US)</option>
                <option value="en-GB">Ingles britanico (en-GB)</option>
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
            <Button size="sm" variant="subtle" icon={<PlugZap className="h-3.5 w-3.5" />} onClick={() => void testLLM()}>
              Probar conexion
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
        </section>

        <div className="signal-line my-6" />

        <section>
          <SectionTitle
            title="Catalogo experimental"
            hint="Busqueda de metadatos mediante el cliente web publico de YouTube Music. No descarga audio ni evade restricciones."
          />
          <label className="flex items-start gap-3 rounded-[9px] border border-line bg-elevated/40 px-3 py-3">
            <input
              type="checkbox"
              checked={form.features.youtubeMusic}
              onChange={(event) => setForm({ ...form, features: { ...form.features, youtubeMusic: event.target.checked } })}
              className="mt-0.5 h-4 w-4 accent-[#85afff]"
            />
            <span>
              <span className="block text-[12.5px] text-ink">Activar busqueda en YouTube Music</span>
              <span className="mt-0.5 block text-[11.5px] leading-relaxed text-muted">
                Los resultados se pueden guardar y ordenar, pero no se reproducen dentro de la app: se abren en el
                reproductor oficial.
              </span>
            </span>
          </label>
        </section>

        <div className="signal-line my-6" />

        <section>
          <SectionTitle title="Proveedores" />
          <ul className="space-y-1.5">
            {providers.map((provider) => (
              <li key={provider.name} className="flex items-start gap-2 text-[12px]">
                <CheckCircle2
                  className="mt-0.5 h-3.5 w-3.5 shrink-0"
                  style={{ color: provider.state === "disabled" ? "#6b7789" : "#85afff" }}
                />
                <span>
                  <span className="text-ink">{provider.name}</span>
                  <span className="text-muted"> · {provider.detail}</span>
                </span>
              </li>
            ))}
          </ul>
        </section>

        <div className="signal-line my-6" />

        <section>
          <SectionTitle title="Acerca de" />
          <dl className="space-y-1 text-[11.5px] text-muted">
            <Row label="Version" value={info ? `${info.name} ${info.version}` : "—"} />
            <Row label="Identificador" value={info?.identifier ?? "—"} />
            <Row label="Base de datos" value={info?.dbPath ?? "—"} mono />
            <Row label="Esquema" value={info ? `v${info.schema}` : "—"} />
            <Row label="Configuracion" value={form.configPath} mono />
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
