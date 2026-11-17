# Proveedores

Toda integracion externa vive detras de un contrato y se puede desactivar. El
motor educativo no depende de ninguno en particular.

## Resumen de capacidades

| Proveedor | Buscar | Metadatos | Portada | Reproducir | Letras | Estado real |
|---|---|---|---|---|---|---|
| Demo (sintetizado) | si | si | degradado generado | **si** | si, con tiempos por palabra | verificado |
| Archivos locales | si | segun etiquetas ID3/Vorbis | no | **si** | archivo `.lrc` al lado del audio | verificado |
| LRCLIB | — | — | — | — | si, por verso o por palabra | verificado contra la API real |
| YouTube Music (experimental) | si | si | si | **no** | no | busqueda verificada en vivo |
| IA (OpenAI-compatible) | — | — | — | — | IPA, aproximacion, traduccion, diccionario | verificado con un stub local |

## Demo

`internal/demo` contiene tres canciones originales escritas para el proyecto.
El audio no se descarga: se sintetiza con `internal/media` (pads, melodia, pulso)
a partir de la misma partitura que produce los tiempos por palabra, asi que audio
y letra estan alineados por construccion. Los WAV se generan en
`<datos>/demo/*.wav` la primera vez y se reutilizan.

Las anotaciones vienen precalculadas con `model = "precomputed"`. Son una
**semilla**: si el usuario genera anotaciones con IA, la semilla no las
sobrescribe en el siguiente arranque (hay una prueba de regresion para eso).

## Archivos locales

- Extensiones: `.mp3 .m4a .aac .wav .flac .ogg .opus`.
- Metadatos leidos con `github.com/dhowden/tag`; si no hay etiquetas, el titulo es
  el nombre del archivo.
- La ruta se guarda tal cual: la app **no copia** el archivo. Si el usuario lo
  mueve, la pista deja de sonar y la UI lo informa.
- La duracion se completa cuando el motor de audio la reporta
  (`SetTrackDuration`).
- Letra: si existe `<mismo nombre>.lrc` junto al audio, se importa sola.

## LRCLIB

- `GET /api/get` con titulo, artista y duracion; si no hay coincidencia exacta,
  `GET /api/search` y se elige el candidato con mejor puntuacion (titulo, artista,
  tolerancia de duracion y preferencia por tiempos por palabra).
- Formatos: LRC (`[mm:ss.xx]`) y LRC mejorado (`<mm:ss.xx>` por palabra). El
  texto plano se guarda como `UNSYNCED`.
- Una consulta por pista y el resultado se cachea en SQLite; el `User-Agent`
  identifica la app, segun la politica del servicio.
- Si la letra solo trae tiempos por verso, la UI lo dice y no inventa tiempos por
  palabra.

## YouTube Music (experimental)

- Cliente InnerTube publico (`WEB_REMIX`), la misma superficie que usa el sitio
  web. La clave de API se descubre en el HTML de `music.youtube.com` y se
  refresca una vez si el servicio la rechaza.
- Operaciones: busqueda (`/youtubei/v1/search`) y lectura de playlists publicas
  (`/youtubei/v1/browse` con `browseId=VL...`).
- **No** extrae audio, **no** usa cookies ni cuenta, **no** evade restricciones y
  **no** promete reproduccion: los resultados se marcan `Playable=false` y se
  abren en la pagina oficial.
- Detras de un flag (`features.youtubeMusic`). Si falla, la busqueda local sigue
  funcionando y se registra un aviso.

## IA (fonetica, traduccion, diccionario)

- Cualquier endpoint compatible con `POST /chat/completions`. Configurable: URL
  base, modelo, temperatura, timeout, reintentos y modo JSON.
- `https` obligatorio; `http` solo para `localhost`, `127.0.0.1` o `::1`.
- La clave se guarda en el almacen del sistema operativo (`go-keyring`), nunca en
  SQLite ni en los registros.
- Respuestas validadas: JSON bien formado, `lineId` exactos (no inventados),
  `original` identico al verso pedido, IPA con caracteres IPA, longitudes maximas
  y notas acotadas. Un lote rechazado se reintenta una vez con el error; lo que
  sigue fallando no se guarda.
- Cache por `(texto normalizado, acento, version de motor, modelo, version de
  prompt, version de formato)`.

## Politica de fallos

| Situacion | Comportamiento |
|---|---|
| Sin red | busqueda local y demo siguen funcionando; el error se informa como aviso |
| 429 / 5xx del LLM | reintentos acotados con backoff y jitter; luego error claro |
| 401 / 403 del LLM | "el proveedor rechazo la clave", sin reintentos |
| Sin clave | `needsSetup`: la UI lleva a Ajustes en lugar de fallar |
| Playlist privada o regional | cero resultados + aviso |
| Cambio de contrato de YouTube | el proveedor queda no disponible y se registra; no bloquea la app |

## Reproduccion en linea (YouTube Music)

`music.YtDlp` resuelve una URL de audio con la herramienta local `yt-dlp` y
`media` la reenvia al reproductor con soporte de `Range`.

| Aspecto | Comportamiento |
|---|---|
| Deteccion | `exec.LookPath("yt-dlp")` al arrancar; la version se muestra en Ajustes |
| Resolucion | `yt-dlp -f bestaudio -g --no-playlist <music.youtube.com/watch?v=ID>`, timeout 25 s |
| Validacion | solo `https`, host `*.googlevideo.com`, `videoID` con forma de 11 caracteres |
| Cache | por `videoID` hasta `expire - 90 s` (o 4 h si la URL no lo declara) |
| Entrega | proxy en Go con `Range`/`If-Range`; el frontend solo ve `/media/<id>` |
| Caducidad | la URL caduca y esta atada a la IP; si falla, se vuelve a resolver |
| Sin yt-dlp | la pista queda `playable: false` con explicacion; no hay error fatal |
| Flag | `features.youtubeMusic`, activo por defecto; `STEFANI_DISABLE_ONLINE=1` lo fuerza apagado (CI) |

Capacidades reales con esta fuente: `play`, `seek`, `rate`, `abLoop` = true,
porque el reloj es el propio `<audio>` y no un iframe.

Riesgos asumidos: va contra los terminos de YouTube, puede romperse cuando
cambie el servicio y el binario no se distribuye con la aplicacion.
