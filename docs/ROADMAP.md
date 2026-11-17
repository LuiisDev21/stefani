# Hoja de ruta

## Completado en el MVP

- [x] Proyecto Wails 3 con React 19, TypeScript estricto y Tailwind 4
- [x] SQLite con migraciones versionadas y repositorios probados
- [x] Biblioteca de demostracion sintetizada con tiempos por palabra
- [x] Importacion de audio local con metadatos y streaming con `Range`
- [x] Letras desde LRCLIB, archivo `.lrc` y sidecar junto al audio
- [x] Renderizador con resaltado progresivo por palabra y autoscroll respetuoso
- [x] Anotaciones de IA (IPA, aproximacion, traduccion) con validacion estricta
- [x] Cache de anotaciones y de letras
- [x] Diccionario de palabras y vocabulario guardado
- [x] Repeticion de versos (A-B) y velocidad 0.75x/1x/1.25x
- [x] Progreso de estudio persistente e historial de reproduccion
- [x] Modo inmersivo
- [x] Ajustes de IA con clave en el almacen del sistema
- [x] Catalogo experimental de YouTube Music (busqueda e importacion)
- [x] Pruebas Go, Vitest y Playwright contra el backend real
- [x] CI con matriz Go, frontend y E2E

## Siguiente

- [ ] Revisar la ventana nativa en Linux y macOS con runners reales
- [ ] Firmar y notarizar instaladores cuando exista certificado
- [ ] Empaquetar con NSIS en Windows (`wails3 package`)
- [ ] Alineacion forzada de audio para letras sin tiempos por palabra (Whisper
      local u otro metodo autorizado), marcada como estimada en la UI
- [ ] Exportar vocabulario a CSV/Anki
- [ ] Edicion manual del desfase por cancion (ahora es por sesion)
- [ ] Soporte TTML para letras que ya lo incluyen
- [ ] Perfil ligero/movil con el mismo `main.go` (Wails 3 lo permite)

## Ideas registradas, fuera de alcance

- Cuentas propias, suscripciones o pagos
- Reconocimiento de voz para puntuar pronunciacion
- Sincronizacion entre dispositivos
- Gamificacion compleja
- Descargas de audio de plataformas externas

Cualquier propuesta que implique descargar audio de terceros, evadir DRM o
reutilizar contenido sin licencia queda descartada por diseno.
