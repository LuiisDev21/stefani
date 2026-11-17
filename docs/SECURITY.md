# Seguridad

## Modelo de amenaza

Aplicacion de escritorio monousuario. Los activos a proteger son la clave de la
API de IA, la biblioteca local (rutas de archivos) y el contenido de la base de
datos. Las superficies de entrada son: respuestas de proveedores externos,
archivos de audio y letra elegidos por el usuario, y la URL base del proveedor de
IA.

## Credenciales

- La clave de IA se guarda con `github.com/zalando/go-keyring`: Credential
  Manager en Windows, Keychain en macOS, Secret Service en Linux.
- **Nunca** se escribe en SQLite, en `config.json`, en los registros ni en los
  DTOs que van al frontend. `Settings` solo expone `hasKey`.
- Se puede eliminar desde Ajustes (`ClearLLMKey`).
- Si el almacen no responde, la app lo dice (`keyStoreError`) y no cae a un
  almacenamiento en texto plano.
- Los perfiles de prueba usan otro espacio de nombres
  (`STEFANI_KEYRING_SERVICE`) para no leer ni escribir las claves reales.

## Entradas y validacion

- **URL del proveedor de IA**: se parsea con `net/url`; solo `https`, o `http`
  cuando el host es `localhost`, `127.0.0.1` o `::1`. Se aplica en la UI (zod) y
  en Go (autoridad final).
- **Enlaces externos**: `OpenExternal` solo acepta `http`, `https` y `mailto`.
- **Rutas de audio**: se resuelven por ID de pista en la base de datos; la URL
  nunca contiene una ruta del sistema.
- **Letras**: tamano maximo 4 MB; un archivo que no produce versos se rechaza.
- **Respuestas de IA**: JSON estricto, IDs existentes, texto original intacto,
  longitudes acotadas. Nada se persiste sin pasar la validacion.
- **SQL**: todas las consultas usan parametros; no hay interpolacion de valores.

## Servidor de medios local

- Escucha en `127.0.0.1` en un puerto efimero.
- Exige un token aleatorio de 128 bits generado por sesion.
- Solo sirve IDs de pista que existen en la base y tienen audio local.
- Rechaza rutas con `/` o `..`.
- Responde con `Range` para permitir busqueda dentro de la pista.
- Se cierra al terminar la aplicacion.

## WebView

- No se carga contenido remoto con privilegios: la ventana sirve solo los assets
  empaquetados y el audio por el listener loopback.
- No se ejecuta codigo que venga de respuestas de red.
- La navegacion externa se delega al navegador del sistema.

## Registros

`log/slog` a archivo y a stderr. No se registran claves, cabeceras de
autorizacion ni cuerpos completos de letras. Los errores de red se registran con
su causa, sin credenciales.

## Datos de terceros

- Las letras de LRCLIB se cachean localmente para estudio privado.
- El catalogo de YouTube Music se consulta sin cuenta, sin cookies y solo para
  metadatos; no se descarga audio.
- No hay telemetria, analitica ni envio de datos a servidores propios.

## Riesgos conocidos

| Riesgo | Estado |
|---|---|
| Proveedor de IA recibe texto de la letra | Aviso en Ajustes; el usuario elige el proveedor. No se envian letras sin accion explicita |
| El audio local es legible por otros procesos del usuario | Inherente a un reproductor de escritorio sin sandbox |
| Endpoint HTTP local malicioso | Solo se permite `localhost`; el usuario decide a que se conecta |
| Cambio de contrato de InnerTube | Se degrada a busqueda local, sin bloquear |

## Reporte

No hay programa de bug bounty. Los defectos se documentan como issues del
repositorio con pasos de reproduccion, impacto y version afectada.

## Musica en linea y proxy de audio

- La app no extrae audio por si misma: delega en un `yt-dlp` instalado por el
  usuario. El binario no se empaqueta ni se descarga automaticamente.
- La URL resuelta (que contiene la IP del usuario y un `expire`) **no** se
  envia al frontend, no se guarda en SQLite y no se registra en logs.
- `internal/media` solo reenvia peticiones a `https://*.googlevideo.com`; un
  resolver que devuelva otro host, `http` o una cadena vacia termina en 404.
- El proxy copia unicamente cabeceras de contenido (`Content-Type`,
  `Content-Range`, `Accept-Ranges`, etc.) y no reenvia `Cookie` ni cabeceras del
  cliente mas alla de `Range`/`If-Range`.
- Sin sesion de Google: no se leen ni se exportan cookies.
- `STEFANI_DISABLE_ONLINE=1` apaga el proveedor para CI y entornos sin red.
