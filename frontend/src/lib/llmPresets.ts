/**
 * Provider presets for the AI settings.
 *
 * Every entry is an OpenAI-compatible chat completions endpoint, which is what
 * the backend speaks. The model ids come from each provider's documentation and
 * are only a starting point: the field stays editable and the list is offered
 * as suggestions.
 *
 * OpenCode Go is special: its docs ask clients to send a stable
 * x-opencode-session header per conversation. The backend adds it
 * automatically, so the preset only has to set the id `opencode-go`.
 */

export interface LLMPreset {
  id: string;
  name: string;
  baseUrl: string;
  model: string;
  models: string[];
  docsUrl: string;
  note: string;
  /** No real API key: the field only needs a placeholder for local servers. */
  local?: boolean;
  /** The backend sends x-opencode-session for this provider. */
  session?: boolean;
}

export const LLM_PRESETS: LLMPreset[] = [
  {
    id: "opencode-go",
    name: "OpenCode Go",
    baseUrl: "https://opencode.ai/inference/go/openai/v1",
    model: "deepseek-v4.1-flash",
    models: [
      "deepseek-v4.1-flash",
      "deepseek-v4-pro",
      "glm-5.3-flash",
      "glm-5.3",
      "kimi-k3.1-pro",
      "kimi-k3",
      "kimi-k2.7-code",
      "qwen3.8-flash",
      "qwen4-max",
      "minimax-m3",
      "mimo-v2.6-flash",
      "longcat-2.0",
      "grok-4.7",
      "gpt-6-luna",
      "gemini-4-argon",
      "claude-haiku-5-5",
    ],
    docsUrl: "https://opencode.ai/docs/console/go/",
    note: "Plan Go de OpenCode ($10/mes). La app envía el id de sesión (x-opencode-session) automáticamente para enrutado y caché de prompt. DeepSeek V4 requiere región Global en Settings › Privacy.",
    session: true,
  },
  {
    id: "deepseek",
    name: "DeepSeek",
    baseUrl: "https://api.deepseek.com/v1",
    model: "deepseek-flash",
    models: ["deepseek-flash", "deepseek-v4-pro", "deepseek-chat", "deepseek-reasoner"],
    docsUrl: "https://api-docs.deepseek.com/",
    note: "deepseek-flash apunta a DeepSeek V4.1 Flash. Compatible con OpenAI; el nombre v1 de la URL no es la versión del modelo.",
  },
  {
    id: "openai",
    name: "OpenAI",
    baseUrl: "https://api.openai.com/v1",
    model: "gpt-6-luna",
    models: ["gpt-6-luna", "gpt-6.1-terra", "gpt-6.1-sol", "gpt-6-astra", "gpt-5.6-luna"],
    docsUrl: "https://developers.openai.com/api/docs/models",
    note: "gpt-6-luna es el modelo eficiente para tareas de alto volumen como anotar letras. Chat Completions sigue disponible para estos modelos.",
  },
  {
    id: "gemini",
    name: "Google Gemini",
    baseUrl: "https://generativelanguage.googleapis.com/v1beta/openai/",
    model: "gemini-3.8-flash",
    models: ["gemini-3.8-flash", "gemini-4-hydrogen", "gemini-3-8-pro"],
    docsUrl: "https://ai.google.dev/gemini-api/docs/openai",
    note: "Endpoint compatible con OpenAI de Gemini. La clave se crea en Google AI Studio.",
  },
  {
    id: "xai",
    name: "xAI (Grok)",
    baseUrl: "https://api.x.ai/v1",
    model: "grok-4.7",
    models: ["grok-4.7", "grok-4.6", "grok-4.5", "grok-4.20-reasoning"],
    docsUrl: "https://docs.x.ai/",
    note: "Misma API compatible con OpenAI. grok-4.7 admite 500K de contexto.",
  },
  {
    id: "groq",
    name: "Groq",
    baseUrl: "https://api.groq.com/openai/v1",
    model: "llama-3.3-70b-versatile",
    models: ["llama-3.3-70b-versatile", "openai/gpt-oss-120b", "openai/gpt-oss-20b", "qwen/qwen3-32b"],
    docsUrl: "https://console.groq.com/docs/openai",
    note: "Inferencia muy rápida. Ideal para anotar muchas canciones seguidas.",
  },
  {
    id: "mistral",
    name: "Mistral",
    baseUrl: "https://api.mistral.ai/v1",
    model: "mistral-large-latest",
    models: ["mistral-large-latest", "mistral-medium-latest", "mistral-small-latest", "ministral-8b-latest"],
    docsUrl: "https://docs.mistral.ai/",
    note: "API compatible con OpenAI. mistral-small-latest es la opción económica.",
  },
  {
    id: "openrouter",
    name: "OpenRouter",
    baseUrl: "https://openrouter.ai/api/v1",
    model: "deepseek/deepseek-chat",
    models: ["deepseek/deepseek-chat", "openai/gpt-6-luna", "google/gemini-3.8-flash", "anthropic/claude-haiku-5.5"],
    docsUrl: "https://openrouter.ai/docs",
    note: "Un solo endpoint para muchos modelos. El id del modelo lleva el proveedor delante (proveedor/modelo).",
  },
  {
    id: "ollama",
    name: "Ollama (local)",
    baseUrl: "http://127.0.0.1:11434/v1",
    model: "llama3.2",
    models: ["llama3.2", "qwen2.5", "mistral", "deepseek-r1"],
    docsUrl: "https://docs.ollama.com/api/openai-compatibility",
    note: "Sin coste ni clave real: escribe cualquier texto en el campo de clave (por ejemplo local). El modelo debe estar descargado (ollama pull).",
    local: true,
  },
  {
    id: "lmstudio",
    name: "LM Studio (local)",
    baseUrl: "http://127.0.0.1:1234/v1",
    model: "local-model",
    models: ["local-model"],
    docsUrl: "https://lmstudio.ai/docs/api/openai-api",
    note: "Servidor local compatible con OpenAI. La clave es un marcador; el modelo es el que tengas cargado.",
    local: true,
  },
];

export function presetByID(id: string | undefined): LLMPreset | undefined {
  return LLM_PRESETS.find((preset) => preset.id === id);
}
