# Arquitectura

## Vista general

```
React 19 (frontend/src)
   │  bindings generados (frontend/bindings)
   ▼
internal/app  ── servicio Wails (DTOs, validacion de entrada, eventos)
   ├── internal/music      proveedores de catalogo e importacion
   ├── internal/lyrics     parseo LRC, LRCLIB, cache
   ├── internal/phonetics  prompts, validacion estricta, cache
   ├── internal/ai         cliente OpenAI-compatible (timeout, retry, JSON)
   ├── internal/demo       fixture de tests: catalogo generado
   ├── internal/learning   progreso (via storage)
   ├── internal/storage    SQLite, migraciones, repositorios
   ├── internal/config     ajustes tipados en JSON
   ├── internal/secret     almacen de credenciales del SO
   ├── internal/media      sintesis WAV y servidor loopback con Range
   └── internal/domain     tipos compartidos, sin dependencias
```

La direccion de dependencias es siempre hacia `domain` y `storage`; ningun
paquete de dominio importa Wails. `internal/app` es el unico que conoce el
contrato con el frontend (DTOs) y el unico que habla con todos los servicios.

## Por que un servidor loopback para el audio

El reproductor necesita `Range` para poder buscar dentro de una pista, y la misma
URL tiene que funcionar en `wails3 dev`, en la app empaquetada y en el build
`server` que usa Playwright. El asset server de Wails no cubre los tres casos con
la misma semantica, asi que `internal/media.Server` abre un listener en
`127.0.0.1:0`, exige un token aleatorio por sesion y sirve solo IDs de pista
resueltos en la base de datos. La URL no contiene rutas: un request manipulado no
puede leer archivos arbitrarios.

## Reloj de reproduccion

Hay un solo `HTMLAudioElement` (creado de forma perezosa en
`frontend/src/stores/playerStore.ts`). Un bucle `requestAnimationFrame` publica su
posicion en un store externo (`frontend/src/player/clock.ts`) y el resto de la UI
se suscribe por campo (`usePlaybackTime`, `useDuration`, `useIsPlaying`). Ningun
componente mantiene su propio temporizador: la letra, la barra de progreso y el
bucle A-B derivan todo del mismo valor, lo que elimina la deriva entre paneles.

## Modelo de datos

SQLite con migraciones embebidas y versionadas (`PRAGMA user_version`). Cada
migracion corre en su propia transaccion; una migracion fallida deja la base en el
ultimo paso completo.

- `tracks` guarda la ruta del audio local (`audio_path`) para los proveedores que
  lo tienen; el resto son entradas de catalogo.
- `lyrics_documents` / `lyrics_lines` / `lyrics_words` normalizan la letra. Los
  IDs de verso son direccionables por contenido (`lrc_<hash>-l000-<hash>`), de
  modo que una reconstruccion conserva las anotaciones de los versos que no
  cambiaron.
- `phonetic_annotations` tiene un indice unico por `(line_id, accent)`; la clave
  de cache incluye version de motor, modelo, prompt y formato de salida.
- `saved_vocabulary`, `learning_progress`, `playback_history` y `settings`
  completan el estado local.

## Contratos del servicio

`internal/app` expone DTOs planos (nunca tipos de dominio) y sigue una regla:

- Estados previsibles (sin letra, sin clave de IA, proveedor deshabilitado) se
  devuelven como campos del DTO (`available`, `notice`, `needsSetup`).
- Errores de programacion o de infraestructura se devuelven como `error` de Go y
  la UI los muestra como aviso.

Asi la interfaz no depende de mensajes de excepcion para pintar estados.

## Concurrencia

- HTTP saliente con `context.WithTimeout` y reintentos con backoff exponencial y
  jitter, acotados por `maxRetries`.
- La base abre con `SetMaxOpenConns(1)` (SQLite tiene un solo escritor) y WAL.
- Las anotaciones se emiten por lotes y publican progreso por eventos
  (`annotate:progress`), sin bloquear el hilo de UI.

## Errores del proveedor de IA

`internal/ai` clasifica los fallos (`ErrAuth`, `ErrRateLimited`,
`ErrNoEndpoint`, `ErrEmptyAnswer`) para que `internal/phonetics` y `internal/app`
puedan reaccionar distinto: pedir configuracion, reintentar, o informar un error
real. Nunca se guarda una respuesta que no pase la validacion.
