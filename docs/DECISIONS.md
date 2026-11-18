# Decisiones de arquitectura

Cada entrada registra el contexto, la decision y la consecuencia. Formato ADR
abreviado.

## ADR-001: Wails 3 en lugar de Wails 2 o Electron

**Contexto**: se necesita una app de escritorio con Go. Wails 3.0.0 salio estable
el 26 de octubre de 2026 y la 3.1.0 estaba disponible al iniciar el proyecto.
**Decision**: usar Wails 3.1.0 con React y TypeScript.
**Consecuencia**: bindings generados y ventana nativa con WebView2/WebKit; el
build `server` permite probar el backend real con Playwright sin emular nada.

## ADR-002: SQLite embebido con `modernc.org/sqlite`

**Contexto**: no se quiere un servidor externo ni cgo extra para la base.
**Decision**: `modernc.org/sqlite` (puro Go), WAL y una sola conexion de escritura.
**Consecuencia**: un unico archivo en el perfil del usuario; las migraciones se
versionan con `PRAGMA user_version`.

## ADR-003: Renderizador de letras propio

**Contexto**: Braccato (MIT) ofrece renderizado y sincronizacion.
**Decision**: se implemento un renderizador propio y se mantuvo la forma del
modelo (`Line`/`Word` con `timingSource`).
**Consecuencia**: control total de tipografia, estados vacios y accesibilidad, sin
atar el diseno a la API de un tercero. La integracion queda como opcion futura.

## ADR-004: Demo sintetizada en lugar de audio de terceros

**Contexto**: se necesita musica reproducible y con tiempos por palabra sin
problemas de derechos.
**Decision**: generar WAV con un sintetizador propio a partir de la misma
partitura que define los tiempos de cada palabra.
**Consecuencia**: audio y letra alineados por construccion, cero licencias de
terceros, ~20 MB en el perfil del usuario la primera vez.

## ADR-005: Sin eSpeak NG

**Contexto**: eSpeak NG puede generar IPA localmente, pero es GPLv3.
**Decision**: no incluirlo. El IPA lo produce el proveedor de IA configurado; la
demo trae valores precalculados.
**Consecuencia**: se evita la obligacion de liberar el binario completo. El costo
es que sin IA no hay IPA para canciones nuevas, y la UI lo dice explicitamente.

## ADR-006: Clave de API en el almacen del sistema

**Contexto**: el requisito prohibe claves en SQLite en texto plano.
**Decision**: `go-keyring`; la base solo guarda el resto de la configuracion.
**Consecuencia**: si el almacen no esta disponible, la app lo informa y no degrada
a texto plano. Los E2E usan un espacio de nombres distinto.

## ADR-007: Servidor loopback con token para el audio

**Contexto**: el elemento `<audio>` necesita `Range` para buscar y la misma URL
debe servir en dev, en produccion y en el build `server`.
**Decision**: listener en `127.0.0.1:0` con token por sesion y resolucion de IDs
en la base.
**Consecuencia**: busqueda precisa y paridad entre entornos; la superficie de red
se limita a loopback y se cierra al salir.

## ADR-008: Tema oscuro unico

**Contexto**: el producto es un reproductor con letras; la portada es el elemento
mas luminoso.
**Decision**: no ofrecer tema claro en el MVP.
**Consecuencia**: se evita un toggle a medias; queda registrado como mejora
futura si aparece demanda real.

## ADR-009: IDs de verso direccionables por contenido

**Contexto**: reconstruir la letra (por ejemplo al reimportar un `.lrc`) borraba
las anotaciones generadas.
**Decision**: el ID del verso combina documento, posicion y hash del texto, y
`SaveLyrics` hace upsert en vez de borrar todo.
**Consecuencia**: las anotaciones sobreviven si el texto no cambio; si cambia, la
anotacion vieja desaparece con el verso viejo, lo cual es correcto.

## ADR-010: La semilla de la demo no sobrescribe datos generados

**Contexto**: `demo.Ensure` corre en cada arranque y guardaba sus anotaciones
precalculadas con upsert, borrando las que el usuario habia generado con IA.
**Decision**: `SeedAnnotation` inserta solo si no existe anotacion para ese verso.
**Consecuencia**: hay una prueba de regresion que verifica que una anotacion
generada sobrevive a un segundo arranque.

## ADR-011: Estados previsibles como datos, no como errores

**Contexto**: "no hay letra" o "falta la clave" son estados normales, no fallos.
**Decision**: los DTOs llevan `available`, `notice`, `needsSetup`; los `error` de
Go se reservan para fallos reales.
**Consecuencia**: la UI no depende de mensajes de excepcion y los tests E2E pueden
afirmar estados, no cadenas de error.

## ADR-012: Sin `build/ios` en el pipeline

**Contexto**: la plantilla de Wails incluye un paquete iOS que no compila en
Windows/Linux con `go build ./...`.
**Decision**: la CI compila el paquete raiz (`go build .`) y ejecuta
`go vet ./...`; el target iOS queda fuera del alcance del MVP.
**Consecuencia**: pipeline determinista; si se retoma movil, se agrega un job con
`GOOS=ios` y Xcode.

## ADR-013: reproduccion no oficial de YouTube Music con yt-dlp

**Contexto**: el pliego inicial prohibia extraer audio. El producto necesita
musica en linea real: sin ella, la app solo reproduce archivos locales y el
estudio depende de que el usuario los tenga. Las apps de referencia resuelven
esto en un backend propio (Liner) o con un wrapper de la web oficial.

**Decision** (aprobada por el dueno del producto, uso personal y no comercial):
resolver la URL de audio con `yt-dlp` local y reproducirla en el `<audio>` de la
app mediante un proxy con `Range`. El flag `features.youtubeMusic` es la fuente
principal de musica en linea y viene activo; se degrada solo si falta la
herramienta.

**Alternativas descartadas**: iframe oficial (limitado a videos embebibles y sin
control de velocidad ni A-B), wrapper de `music.youtube.com` (exige sesion de
Google dentro de la app y rompe el reloj unico), backend propio tipo Liner
(infraestructura y responsabilidad legal que el MVP no necesita).

**Consecuencia**: la app depende de una herramienta externa y de un servicio que
puede cambiar sin aviso. Se mitiga con cache, validacion, mensajes claros y
fallback a local/demo. yt-dlp no se empaqueta. El riesgo de terminos de servicio
queda documentado y asumido.

## ADR-014: pronunciacion local con CMUdict en lugar de depender de la IA

**Contexto**: la demo traia IPA precalculado, pero cualquier otra cancion
mostraba el panel vacio hasta que el usuario configurara un proveedor de IA y
pulsara "Anotar". Eso convertia la funcion principal del producto en rehén de
una credencial externa.

**Decision**: embeber `cmudict.dict` (135k palabras, ARPAbet con acento) y
convertirlo a IPA y a lectura espanola en el proceso. La IA queda para
traduccion, notas de contexto y vocabulario.

**Alternativas descartadas**: g2p por reglas (precision ~70 %, inaceptable para
ensenar), eSpeak NG (GPL y proceso externo), exigir clave para ver fonetica.

**Consecuencia**: el binario crece ~3,5 MB y el diccionario se parsea en el
primer uso (~150 ms). La pronunciacion funciona offline desde el primer arranque
y la IA pasa a ser una mejora, no un requisito.

## ADR-015: rediseno sobre apple.design.md (tema claro, un solo azul)

**Contexto**: el MVP nacio con un tema oscuro propio. El dueno del producto
aporto `apple.design.md` como referencia y pidio redisenar toda la interfaz sobre
ese sistema: lienzo claro, Parchment, tipografia SF Pro, Action Blue unico,
tiles full-bleed, una sola sombra reservada al artwork y navegacion superior en
lugar de barra lateral.

**Decision**: adoptar la spec como fuente de tokens y estructura. La navegacion
pasa a una barra negra de 44 px; el panel de estudio queda como columna de
400 px; la barra de reproduccion es la sticky bar frosted; el modo inmersivo
invierte a negro puro. El violeta y los degradados decorativos desaparecen.

**Desviaciones conscientes** (documentadas porque se apartan de la letra de la
spec):
1. El cuerpo de la app usa 14-15 px, no 17: es una superficie de listas densas.
   La lectura (letra, hero) si mantiene los pasos editoriales.
2. La spec es de marketing; una app necesita estados vacios, loading y error.
   Se resolvieron con superficie y hairlines, sin tarjetas de relleno.
3. El placeholder de portada es un sleeve tipografico, no una imagen fingida.

**Consecuencia**: el tema oscuro unico (ADR-008) queda superado para la
interfaz; el modo inmersivo sigue siendo oscuro por funcion, no por defecto.

## ADR-016: cliente LRCLIB sobre v2 y tolerancia al 400

**Contexto**: con la busqueda de YouTube devolviendo filas sin artista, el panel
mostraba «LRCLIB answered 400». El endpoint `/api/get` exige `track_name` y
`artist_name`; sin artista el framework responde 400 en texto plano y el error
abortaba el fallback.

**Decision**: el flujo pasa a `/api/v2/search` (solo metadatos y flags) y
`/api/get/{id}` para el registro elegido. El 400 se trata como "sin letra", el
cliente limpia titulos y artistas de YouTube antes de consultar, y respeta
`Retry-After` en 429/503 con un maximo de 5 s para no bloquear la UI.

**Consecuencia**: una consulta ya no descarga 20 letras completas, el panel cae
a "sin letra" en vez de error, y un catalogo con metadatos incompletos se
corrige al primer resultado (la fila local se refresca).

## ADR-017: presets de IA con descubrimiento real y OpenCode Go

**Contexto**: ajustar la IA exigia conocer de memoria la URL y el modelo de cada
proveedor, y OpenCode Go no funcionaba porque su API pide un header de sesion
(`x-opencode-session`) y un User-Agent identificable.

**Decision**: los presets son datos del frontend que rellenan endpoint y modelo;
la lista definitiva de modelos se pide al proveedor con `GET /models` dentro de
la accion "Probar y buscar modelos" (guardar + ping + descubrimiento). El
backend gana `ai.Client.Models`, `LLM.Provider` y un `client-id` de instalacion
persistido aparte; el cliente envia `x-opencode-session` solo para
`opencode-go` y siempre manda su User-Agent.

**Consecuencia**: no hay modelos hardcodeados en la UI (los presets son solo un
punto de partida editable), un proveedor sin `/models` sigue siendo usable, y el
session id sobrevive a cambios de configuracion sin exponerse en config.json.

## ADR-018: inmersivo de dos columnas y sesion de medios del sistema

**Contexto**: el modo inmersivo centraba la letra y no mostraba la portada, y la
app no aparecia en el widget de medios de Windows.

**Decision**: el inmersivo adopta la composicion de referencia (portada y
creditos a la izquierda, letra a la derecha, activa en ambar). La integracion con
el sistema va por la **Media Session API** del renderer: metadatos, caratula de
544 px, estado de reproduccion, posicion y handlers de play/pausa/siguiente/
anterior/seek. No se implementa SMTC por COM porque WebView2 ya publica la
sesion.

**Evidencia**: con el WebView2 de la app reproduciendo, PowerShell enumera
`msedgewebview2.exe | Friend, Please | Twenty One Pilots | Playing`, y
`TryTogglePlayPauseAsync` sobre esa sesion pausa y reanuda el `<audio>` real
(`paused` cambia en el renderer).

**Consecuencia**: el widget de Windows y las teclas multimedia controlan la app
sin codigo nativo por plataforma. En macOS/Linux el equivalente depende del
soporte de Media Session del WebView del sistema y no se ha verificado.

## ADR-020: rediseno sobre DESIGN.md (Raycast) y inicio como reproductor web

**Contexto**: el sistema claro de apple.design.md daba una app correcta pero de
documento; el usuario pidio el sistema de DESIGN.md y un inicio con la
consistencia de un reproductor web.

**Decision**: se sustituyen los tokens por la paleta Raycast (lienzo #07080a,
escalera de cuatro superficies, hairlines #242728, CTA blanca) con Inter y ss03.
El inicio pasa a ser un dashboard: hero con la portada de lo que suena (o de la
sesion lista), tres contadores reales, accesos rapidos, canciones, listas y una
estanteria de lo escuchado. El modo inmersivo queda intacto, incluido su ambar.

**Consecuencia**: cero sombras fuera del inmersivo, un solo color de accion y
un unico momento cromatico. Los metadatos suben a `#a9a9ab`/`#8b8c8d` para
cumplir AA sobre negro. Cualquier vista nueva hereda el tema usando los tokens.

## ADR-021: el producto es un reproductor, no una herramienta de estudio

**Contexto**: la app acumulo vocabulario, progreso de practica, repeticion A-B y
un catalogo generado de tres canciones. El uso real es escuchar con la letra
sincronizada, la pronunciacion y la traduccion.

**Decision**: se elimina la capa de aprendizaje (vista Aprendizaje, vocabulario,
progreso, repeticion de versos y su dominio en Go) y el catalogo generado deja de
sembrarse en produccion. Se conservan letra, IPA + aproximacion local, traduccion
por IA, desfase, busqueda, biblioteca, listas, favoritos y modo inmersivo. La
migracion `0002_player_only.sql` borra las pistas demo y las tablas de
aprendizaje; `internal/demo` queda solo como fixture (`STEFANI_SEED_DEMO=1`).

**Consecuencia**: la app arranca vacia y se llena con el usuario. El boton del
panel pasa a "Generar" para no confundirse con el chip "Traduccion", y el chip de
pronunciacion se llama "Pronunciacion". Los tests usan `STEFANI_DISABLE_AI=1`
para no llamar a un proveedor real.
