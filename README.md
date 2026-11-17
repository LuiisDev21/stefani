# STEFANI

Reproductor musical de escritorio con letras sincronizadas palabra por palabra,
transcripcion IPA, aproximacion fonetica para hispanohablantes y traduccion al
espanol. Construido con Go, Wails 3, React 19 y SQLite.

El objetivo del MVP es un flujo educativo completo y estable: escuchar una
cancion, leer su letra siguiendo el audio, entender como se pronuncia y guardar
lo aprendido. La aplicacion funciona sin cuenta, sin servidor externo y sin
configuracion inicial: incluye tres canciones originales de demostracion con
audio sintetizado y tiempos por palabra.

## Estado

MVP funcional en Windows (probado en esta maquina), con el codigo preparado para
Linux y macOS. Lo que esta verificado esta en [docs/TESTING.md](docs/TESTING.md);
las limitaciones conocidas, en la seccion final de este documento.

## Funcionalidades

- **Reproduccion local**: importa tus archivos (mp3, m4a, aac, wav, flac, ogg,
  opus). Los archivos se referencian en su ubicacion original, no se copian.
- **Demo integrada**: tres canciones originales, audio sintetizado por la propia
  aplicacion, letra con tiempos por palabra e IPA precalculado.
- **Letras sincronizadas**: LRC, LRC mejorado (tiempos por palabra) y texto sin
  sincronizar, con autoscroll que se rinde ante el scroll manual.
- **Lectura fonetica sin configurar nada**: IPA estadounidense y aproximacion
  en espanol para cualquier letra, generadas en local con el diccionario CMU.
  No hace falta clave de IA para ver la pronunciacion.
- **Notas y traduccion con IA**: cuando configuras un proveedor, la anotacion
  anade traduccion contextual y explicaciones de los sonidos dificiles.
- **Traduccion y vocabulario**: traduccion contextual al espanol y consulta de
  palabras con el proveedor de IA configurado.
- **Repeticion de versos (A-B)** y control de velocidad 0.75x / 1x / 1.25x.
- **Musica en linea**: busqueda en el catalogo de YouTube Music y reproduccion
  dentro de la app mediante `yt-dlp`, con las mismas letras, velocidad y A-B que
  el audio local. Es una integracion no oficial, activa por defecto y se apaga
  sola si la herramienta no esta instalada.
- **Biblioteca, favoritos, playlists locales e importacion de playlists de
  YouTube Music**.
- **Progreso persistente**: tiempo de practica, versos trabajados, vocabulario.
- **Modo inmersivo** de pantalla completa para estudiar la letra.
- **Ajustes de IA**: cualquier endpoint compatible con OpenAI Chat Completions;
  la clave se guarda en el almacen de credenciales del sistema.
- **Ajustes de musica en linea**: interruptor propio y estado real del
  reproductor (`yt-dlp` detectado o ausente) con la instruccion de instalacion.

## Requisitos

- Go 1.26 o superior (probado con 1.27.1)
- Node 22 o superior y pnpm 11
- Wails CLI v3: `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0`
- Windows: WebView2 (viene con Windows 11) y un compilador C para cgo
- Linux: `libgtk-3-dev` y `libwebkit2gtk-4.1-dev`
- macOS: Xcode Command Line Tools
- Nada extra para la pronunciacion: el diccionario CMU va incluido en el
  binario.
- Opcional, para musica en linea: `yt-dlp` en el PATH
  (`winget install yt-dlp`). No se empaqueta con la aplicacion: sin el, la app
  funciona igual con audio local y demo.

## Desarrollo

```bash
# 1. dependencias del frontend
cd frontend && pnpm install && cd ..

# 2. bindings de Go a TypeScript (tras cambiar la API del servicio)
wails3 generate bindings

# 3. modo desarrollo con recarga en caliente
wails3 dev

# 4. compilacion de escritorio
go build -o bin/stefani.exe .        # Windows
go build -o bin/stefani .            # Linux / macOS
```

`wails3 build` empaqueta instaladores; en Windows requiere NSIS, que no forma
parte del alcance actual (ver [docs/DECISIONS.md](docs/DECISIONS.md)).

## Pruebas

```bash
go test ./internal/... -count=1      # dominio, storage, letras, IA, demo
go vet ./...
cd frontend && pnpm exec tsc --noEmit && pnpm test
cd frontend && pnpm exec playwright test   # E2E contra el servidor Go real
```

El arnes E2E compila la aplicacion con `-tags server` y Playwright la maneja
contra `http://127.0.0.1:8080` con un `STEFANI_HOME` temporal, de modo que nunca
toca el perfil real. Detalles y cobertura en [docs/TESTING.md](docs/TESTING.md).

## Datos y privacidad

- Todo se guarda en el directorio de configuracion del usuario:
  `%APPDATA%\stefani` (Windows), `~/.config/stefani` (Linux y macOS).
- La base de datos es SQLite en `data/stefani.db`.
- La clave de la API de IA vive en el almacen de credenciales del sistema, nunca
  en la base de datos ni en los registros.
- Las letras obtenidas de LRCLIB se guardan localmente para estudio privado.
- No hay telemetria ni cuentas.

## Limitaciones actuales

- **YouTube Music**: la busqueda y la importacion de playlists usan el cliente
  web publico (InnerTube). No se descarga audio ni se evaden restricciones, por
  lo que esas pistas no se reproducen dentro de la app; se abren en la pagina
  oficial. Requiere activar el catalogo experimental en Ajustes.
- **eSpeak NG** no se incluye: su licencia GPL complica la distribucion del
  binario. La transcripcion IPA la genera el proveedor de IA configurado; la demo
  trae valores precalculados.
- **Alineacion de audio**: cuando una letra solo tiene tiempos por verso, la app
  lo indica y no inventa tiempos por palabra.
- **Velocidad y A-B loop** dependen del motor HTML5 Audio; para fuentes que no lo
  soportan (catalogo de YouTube) los controles quedan deshabilitados.
- **Sin firmar**: los instaladores no estan firmados ni notarizados.
- **Validacion nativa**: el ejecutable de Windows se compila y arranca, pero las
  pruebas automatizadas se ejecutan sobre el build `server`; la revision visual de
  la ventana nativa queda como tarea manual (ver docs/TESTING.md).

## Documentacion

- [Arquitectura](docs/ARCHITECTURE.md)
- [Sistema de diseno](docs/DESIGN_SYSTEM.md)
- [Proveedores](docs/PROVIDERS.md)
- [Pruebas](docs/TESTING.md)
- [Seguridad](docs/SECURITY.md)
- [Licencias](docs/LICENSES.md)
- [Decisiones](docs/DECISIONS.md)
- [Hoja de ruta](docs/ROADMAP.md)

El nombre, identificador y branding viven en `internal/brand/brand.go` y en
`build/config.yml`: renombrar el producto no requiere tocar la arquitectura.
