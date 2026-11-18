# Sistema de diseno

## Lectura de diseno

> Se lee como: una aplicacion de escritorio para escuchar y estudiar canciones,
> construida con el lenguaje de `apple.design.md`: lienzo claro, tipografia con
> tracking negativo, un unico azul de accion y nada de chrome decorativo. Diales
> **ENERGY 2 / RHYTHM 2 / MOTION 1**.

- **ENERGY 2**: la pantalla es blanca. El unico color saturado es Action Blue y
  aparece solo donde hay una accion real (pill, enlace, foco, switch). La
  portada y la letra activa son el foco; el resto recede.
- **RHYTHM 2**: Inicio abre con un tile en Parchment; Biblioteca y Listas son
  listas; Aprendizaje es una rejilla; Ajustes es una columna de lectura. Las
  secciones se separan por superficie y hairline, nunca por cajas repetidas.
- **MOTION 1**: transiciones de color, opacidad y `scale(0.97)` al pulsar. El
  autoscroll de la letra es funcional. No hay pulsos ni bucles.

La referencia completa esta en `apple.design.md` en la raiz del repositorio. Las
decisiones de adaptacion (por que el cuerpo de la app no va a 17 px en todas
partes, por que el modo inmersivo invierte la superficie) estan en
[DECISIONS.md](DECISIONS.md).

## Paleta

| Token | Valor | Uso |
|---|---|---|
| `bg` | `#FFFFFF` | lienzo dominante |
| `elevated` | `#F5F5F7` | Parchment: tiles, barras frosted, estados vacios |
| `surface` | `#EDEDF0` | hover y pista del control |
| `ink` | `#1D1D1F` | titulares y cuerpo |
| `muted` | `#5F5F64` | texto secundario |
| `faint` | `#737378` | metadatos y microcopy |
| `accent` | `#0066CC` | Action Blue: toda accion interactiva |
| `accent-focus` | `#0071E3` | anillo de foco |
| `line` | `#E0E0E0` | hairline de 1 px |
| `line-soft` | `#F0F0F0` | separadores internos |
| `danger` | `#D70015` | error y accion destructiva |
| `dark` | `#1D1D1F` | texto sobre blanco en el modo inmersivo invertido |
| `dark-tile` | `#272729` | tiles oscuros del modo inmersivo |
| `on-accent` | `#FFFFFF` | texto sobre azul |

Contraste WCAG 2.1 calculado sobre `#FFFFFF`: `ink` 16.8:1, `muted` 6.35:1,
`faint` 4.72:1, `accent` 5.57:1 y blanco sobre azul 5.57:1. Todo el texto cumple
AA; el principal llega a AAA. Sobre Parchment: `ink` 15.5:1 y `muted` 5.8:1.

**Sin segundo acento.** El violeta anterior desaparecio: el IPA se lee en
`muted` y la lectura en espanol en `ink/75`. El estado de reproduccion (palabra
que se ilumina) usa tinta, no color.

**Sin gradientes decorativos.** El unico gradiente del producto es el relleno
duro de dos paradas que pinta la palabra activa, porque es la unica forma de
mostrar progreso dentro de un solo nodo de texto. La portada sin imagen es un
sleeve tipografico, no un degradado generado.

## Tipografia

SF Pro Display/Text cuando el sistema la tiene; si no, Inter (incluida en el
binario, OFL). Tracking negativo en todos los tamanos de titular.

| Nivel | Tamano / peso | Uso |
|---|---|---|
| Vista | 34 px / 600 / -0.374 px | `TopBar`, unico H1 por pantalla |
| Seccion | 21 px / 600 / -0.231 px | `SectionTitle`, titulo de cancion |
| Hero | 40 px / 600 / -0.28 px | solo el tile de Inicio |
| Cuerpo | 14-15 px / 400 / 1.47 | UI y listas |
| Lectura | 16-21 px / 400-600 | letra y panel de estudio |
| Secundario | 14 px / 400 | descripciones |
| Metadatos | 12-13 px / 400 | duraciones, contadores, atajos |

La escalera es 300 / 400 / 600 / 700: **no se usa 500**. El cuerpo de la app va
a 14-15 px porque es una superficie densa de listas, igual que Apple Music; el
panel de letra y el hero si usan los pasos editoriales de 16-22 px.

## Espaciado y radios

- Base 8 px; gutters de 8/12/16/24; padding lateral de vista 32 px.
- Radios: `pill` (9999) para toda accion, `18 px` para tiles y tarjetas,
  `8 px` para miniaturas y utilidades, `0` para el tile full-bleed del hero.
- **Una sola sombra** (`3px 5px 30px rgba(0,0,0,0.18)`) y solo bajo artwork que
  descansa sobre una superficie. Las barras usan blur funcional
  (`saturate(180%) blur(20px)` al 82 % de Parchment), no sombra.

## Motivo de identidad

La letra como objeto. El verso activo crece a 21 px/600, la palabra en curso se
rellena de tinta y debajo aparecen IPA y lectura en espanol. La barra negra de
navegacion y el lienzo blanco son el marco; la letra es lo que se expone.

## Componentes

- **Nav global**: 44 px, negro puro, 12 px, sin iconos, marca a la izquierda y
  accion de importar a la derecha. Es el unico negro de la pantalla.
- **TopBar**: H1 de 34 px + linea de 14 px + acciones pill a la derecha.
- **Pills**: primaria azul, secundaria con hairline, fantasma azul. `scale(0.97)`
  al pulsar.
- **Tiles**: Parchment, radio 18, sin borde ni sombra.
- **Listas**: filas de 44-56 px, hairline entre filas, hover `surface`, sin
  tarjetas por fila.
- **Panel de estudio**: 400 px, blanco, borde izquierdo hairline, letra a 16 px.
- **Reproductor**: barra frosted de 72 px, sin sombra, con la portada de 48 px
  como unico objeto elevado.
- **Modo inmersivo**: dos columnas sobre `#0b0b0c`. A la izquierda, la portada
  (26vw, tope 460 px) y debajo titulo 23 px / artista 15 px; la propia portada
  ilumina el fondo con una copia desenfocada al 40 % y velo negro al 72 %. A la
  derecha, la letra en 20-29 px semibold con la linea activa en ambar `#ffc000`
  y el resto en blanco al 65 %. Controles y progreso en una fila inferior
  discreta. Invierte la superficie, no la identidad.
- **Barra superior**: cuando hay pista, muestra portada de 28 px, `Artista —
  Titulo` truncado y play/pausa y siguiente, separados del resto por una linea
  vertical. El contador de la biblioteca se oculta mientras suena algo.

## Movimiento

Solo transiciones de 150-200 ms en color, opacidad y `transform`. El pulso
desaparecio del arranque: la carga usa un skeleton con la geometria del
contenido. No hay animaciones en bucle.

## Accesibilidad

- Foco visible: anillo de 2 px en `accent-focus` con offset de 2 px.
- Todo control de icono tiene nombre accesible y tooltip; los chips son
  `aria-pressed`; los switches, `role="switch"` con `aria-checked`.
- Objetivos tactiles de 32-44 px; el transporte principal, 40 px.
- En el modo inmersivo el texto blanco sobre negro alcanza 21:1 y el secundario
  6.2:1. `prefers-reduced-motion` desactiva transiciones y autoscroll.
