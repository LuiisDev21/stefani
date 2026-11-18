# Sistema de diseno

Derivado de `DESIGN.md` (el sistema de Raycast) y vigente desde el rediseno de
la version 0.1. Sustituye al sistema claro basado en `apple.design.md`.

## Diales

- **ENERGIA 2**: superficie oscura unica, cero decoracion, el unico momento
  cromatico es la linea de letra que suena en el modo inmersivo.
- **RITMO 2**: el inicio es un dashboard de reproductor (hero, accesos rapidos,
  estanteria de portadas, listas); las demas vistas son listas densas.
- **MOVIMIENTO 1**: transiciones de 150 ms en color y opacidad; sin bucles, sin
  pulsos, sin animaciones de entrada.

## Paleta

| Token | Valor | Uso |
|---|---|---|
| `bg` | `#07080a` | lienzo |
| `elevated` | `#0d0d0d` | tarjetas y hero |
| `surface` | `#101111` | filas activas, inputs |
| `card` | `#121212` | modales, tooltips, portada tipografica |
| `ink` | `#f4f4f6` | texto principal (18.2:1) |
| `body` | `#cdcdcd` | texto de lectura (12.6:1) |
| `muted` | `#a9a9ab` | metadatos (8.5:1) |
| `faint` | `#909193` | metadatos menores (6.3:1) |
| `ash` | `#7c7d80` | iconos inactivos (4.9:1) |
| `line` | `#242728` | hairline de toda tarjeta |
| `accent` | `#ffffff` | unica accion primaria y enlaces |
| `on-accent` | `#000000` | texto sobre la accion primaria |
| `danger` | `#ff6161` | errores |

La escalera de superficies es la que da profundidad: **no hay sombras**. La
unica excepcion es la portada del modo inmersivo, que conserva
`--shadow-product` porque esa vista se diseno aparte y su referencia lo pide.

El ambar `#ffc000` no pertenece a este sistema: vive solo en el modo inmersivo,
donde marca el verso que suena. No se usa en el resto de la interfaz.

## Tipografia

Inter variable local, con `font-feature-settings: "calt", "kern", "liga",
"ss03"`. La escala:

- Titulo de vista: 40 px / 600 / -0.035em.
- Hero: 36-44 px / 600 / -0.04em.
- Seccion: 20 px / 500.
- Cuerpo: 14 px / 400, `body`.
- Metadatos: 12-13 px / 400, `muted` o `faint`.
- Etiquetas de accion: 13-14 px / 500.

Nada de mayusculas con tracking amplio: las etiquetas van en caja normal.

## Forma

- Radios: 6 px (filas y chips), 8 px (botones e inputs), 10 px (tarjetas,
  portadas), 16 px (paneles y modales), 9999 px solo para la pildora del
  reproductor y los avatares.
- Bordes: 1 px `line` en toda tarjeta, input y separador. No hay bordes de 2 px
  ni degradados decorativos.
- Foco: anillo blanco de 2 px con 2 px de separacion.

## Componentes

- **Boton primario**: fondo blanco, texto negro, radio 8, alto 36, peso 500. Es
  la unica accion blanca de la pantalla.
- **Boton secundario**: transparente con hairline; al pasar, borde mas claro y
  superficie `surface`.
- **Nav superior**: 52 px, `bg`, hairline inferior, destino activo sobre `card`.
- **Hero de inicio**: tarjeta de 16 px con portada de 220, titulo 44, una accion
  primaria y tres contadores reales (practica, versos, pistas).
- **Estanteria**: tarjetas de portada cuadradas con radio 10; el boton de play
  aparece al pasar o al enfocar.
- **Reproductor**: barra solida `elevated` con hairline superior; sin blur.
- **Panel de letra**: 400 px, `bg`, hairline izquierdo.
- **Inmersivo**: sin cambios. Negro `#0b0b0c`, letra blanca, verso activo
  ambar, portada con la unica sombra del sistema.

## Desviaciones conscientes

- `ash` se subio de `#6a6b6c` a `#7c7d80` para que los metadatos de 11-12 px
  cumplan AA (4.9:1). Raycast lo usa mas apagado.
- El modo inmersivo conserva su paleta y su sombra: es una superficie aparte,
  pedida por el usuario, y no se toca en este rediseno.
- La portada sin arte es una funda tipografica, no un album inventado.
