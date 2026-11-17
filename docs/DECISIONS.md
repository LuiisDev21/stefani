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
