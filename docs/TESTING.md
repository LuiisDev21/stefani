# Pruebas

## Resumen de lo ejecutado

| Capa | Comando | Resultado |
|---|---|---|
| Go (dominio, storage, proveedores, letras, IA, demo, medios, app) | `go test ./internal/... -count=1` | **79 pruebas, 8 paquetes, todas pasan** |
| Go estatico | `go vet ./...` | sin hallazgos |
| Frontend unitario | `pnpm test` (Vitest + Testing Library) | **29 pruebas, 3 archivos, todas pasan** |
| Tipos | `pnpm exec tsc --noEmit` | sin errores, `strict: true` |
| E2E | `pnpm exec playwright test` | **7 pruebas, todas pasan** contra el build `server` real |
| Build escritorio | `go build -o bin/stefani.exe .` | binario de ~26 MB |
| Humo nativo (Windows) | script `stefani-native-smoke.ps1` | ventana "STEFANI" abierta, DB y WAV creados, cierre limpio (exit 0) y geometria persistida en `config.json` |

## Que cubre cada capa

### Go

- **storage**: migraciones aplicadas una sola vez, claves foraneas activas, upsert
  de pistas, orden y reubicacion de playlists, borrado en cascada, acumulacion de
  progreso, vocabulario unico por palabra, `SaveLyrics` conserva anotaciones de
  versos sin cambios y descarta las de versos modificados.
- **demo**: tiempos monotonos por palabra, WAV valido (RIFF, mono, 44100),
  `Ensure` idempotente y **regresion**: una anotacion generada por el usuario
  sobrevive a un segundo arranque. Ademas, un invariante: unir los tiempos de
  palabra reconstruye el verso exacto, lo que impide que el karaoke parta
  palabras por la mitad.
- **media**: servidor loopback con token (403 sin token), 200 con cuerpo completo,
  206 con `Range` y `Content-Range`, 404 para pista desconocida, rechazo de
  traversal, tipos MIME correctos. Proxy remoto: reenvia `Range` al origen,
  conserva `Content-Range`/`Content-Type`, responde 404 ante hosts ajenos,
  `http` plano o fuente vacia.
- **music (yt-dlp)**: deteccion ausente, URL valida y cache (una sola invocacion),
  identificadores invalidos sin ejecutar el binario, host no permitido, rechazo
  de `http`, error con detalle saneado (sin URLs) y expiracion de la cache.
- **music**: parseo de metadatos de InnerTube, refresco de credenciales una sola
  vez, `parseClock`, importacion de archivos no soportados, catalogo sin
  `Playable`.
- **lyrics**: LRC plano (UNSYNCED), LRC con tiempos de linea, LRC mejorado con
  tiempos por palabra, offsets, timestamps duplicados, orden fuera de secuencia,
  seleccion por duracion en LRCLIB, cache en SQLite (la segunda consulta no toca
  la red), lectura de `.lrc` al lado del audio.
- **ai**: reintento ante 429, limite de reintentos, 401 sin reintentar, respuesta
  vacia, cancelacion por contexto, limpieza de bloques markdown, backoff acotado.
- **phonetics**: validacion estricta (lineId inventado, `original` alterado),
  cache por clave, reparacion en un solo reintento, error de credenciales,
  normalizacion de IPA, busqueda de palabras.
- **app**: bootstrap con la demo, letra inexistente como estado (no error),
  validacion de URL de IA, clave solo en el almacen seguro, `needsSetup` sin
  clave, biblioteca, busqueda, importacion, parseo de URL de playlist, geometria
  de ventana y rechazo de esquemas peligrosos en `OpenExternal`.

### Frontend

- **timeline**: linea activa (incluido el caso "antes del primer verso"), progreso
  por palabra, palabra activa, limites A-B, offset, soporte de karaoke.
- **format**: tiempos, duracion de estudio, fechas relativas, clamp.
- **LyricLine**: cada palabra se renderiza una sola vez (sin duplicar el texto del
  karaoke), IPA/aproximacion/traduccion solo en el verso activo, toggles, click y
  teclado sobre palabras, seek al pulsar el verso.

### E2E (Playwright)

Contra `go run -tags server,production .`, con `STEFANI_HOME` y
`STEFANI_KEYRING_SERVICE` propios:

1. carga la biblioteca de demostracion desde el servicio Go y no registra errores
   de consola (la suite corre con `STEFANI_DISABLE_ONLINE=1`: determinista y sin
   red);
2. navega las siete secciones y mantiene `aria-current`;
3. muestra letra con tiempos por palabra, IPA, aproximacion y traduccion;
4. crea una playlist que sobrevive a un `reload`;
5. sin clave de IA pide configuracion y lleva a Ajustes en vez de fallar;
6. rechaza un endpoint `http` remoto antes de guardar;
7. filtra la biblioteca desde la busqueda.

## Verificaciones manuales con servicios reales

| Verificacion | Resultado |
|---|---|
| `https://lrclib.net/api/get` y `/search` | respuesta real con `hasWordSync` |
| YouTube Music InnerTube (busqueda) | `STEFANI_LIVE=1 go test ./internal/music -run Live -v` devuelve metadatos reales |
| Streaming de audio | `curl` con y sin `Range`: 200 y 206, WAV RIFF valido, 3.6 MB |
| Ruta completa de IA | stub OpenAI-compatible local: "Probar conexion" OK (4 ms) y 6 versos anotados, luego cacheados |
| Persistencia | tras reiniciar: playlists, ajustes, clave y progreso siguen ahi |

## Limitaciones del arnes

- **Audio en CI**: Chromium headless del entorno de automatizacion no decodifica
  audio (se comprobo tambien con un WAV en `data:`), asi que los E2E afirman los
  estados derivados del catalogo, la base y los eventos, no la forma de onda. La
  reproduccion real se verifico en la ventana nativa y con el servidor de medios
  via `curl`.
- **UI nativa**: el humo comprueba proceso, ventana, datos y cierre limpio; no
  compara pixeles. La revision visual se hizo sobre el build `server` en Chromium
  (capturas durante el desarrollo).
- **Linux y macOS**: el codigo no usa APIs exclusivas de Windows y la CI compila
  en Ubuntu, pero no se ejecuto la app nativa en esos sistemas.

## Como reproducir

```bash
# backend
go test ./internal/... -count=1
go vet ./...

# frontend
cd frontend
pnpm install --frozen-lockfile
pnpm exec tsc --noEmit
pnpm test

# E2E (compila el frontend y arranca el servidor Go real)
CI=1 pnpm exec playwright test

# humo nativo (Windows)
go build -o bin/stefani.exe .
powershell -NoProfile -ExecutionPolicy Bypass -File "$LOCALAPPDATA/Temp/stefani-native-smoke.ps1"
```
