# Sistema de diseno

## Lectura de diseno

> Se lee como: aplicacion de escritorio de escucha y estudio para adultos que
> aprenden ingles, en un lenguaje visual nocturno y editorial, con diales
> **ENERGY 2 / RHYTHM 2 / MOTION 2**.

- **ENERGY 2**: la interfaz no grita. Un solo acento saturado y un solo borde
  luminoso (el foco) por pantalla.
- **RHYTHM 2**: las vistas cambian de composicion (banda destacada + listas en
  Inicio, lista unica en Biblioteca, maestro-detalle en Listas), pero comparten
  la misma retícula de 7 px de padding lateral y la misma linea de senal.
- **MOTION 2**: transiciones de color y opacidad, autoscroll suave de la letra y
  relleno progresivo de la palabra activa. Nada se mueve de forma continua.

## Paleta

| Token | Valor | Uso | Razon |
|---|---|---|---|
| `bg` | `#0B0E16` | fondo de la app | nocturno, deja que la portada sea lo mas brillante |
| `elevated` | `#121725` | paneles y barras | separa planos sin sombras grandes |
| `surface` | `#1B2232` | hover, item activo | un solo paso de elevacion visible |
| `ink` | `#F4F6FA` | texto principal | 17.8:1 sobre `bg` |
| `muted` | `#A7B3C4` | texto secundario | 9.1:1 sobre `bg` |
| `faint` | `#8B98AC` | metadatos y etiquetas | 6.6:1 sobre `bg`, sigue AA |
| `accent` | `#85AFFF` | accion primaria, foco, progreso | unico acento saturado; 8.8:1 con texto `bg` |
| `violet` | `#B0A0FF` | IPA | distingue la transcripcion de la traduccion; 8.6:1 |
| `line` | `#273044` | bordes de 1 px | profundidad por borde, no por sombra |

Contraste calculado con la formula WCAG 2.1 (luminancia relativa). Valores
comprobados sobre `#0B0E16`; todo el texto cumple AA y el principal llega a AAA.

El tema es **solo oscuro**, decision de producto documentada en
[DECISIONS.md](DECISIONS.md#adr-008-tema-oscuro-unico): es una superficie de
escucha y lectura, la portada es el elemento mas luminoso y el uso tipico es
nocturno. No se ofrece un toggle a medias.

## Tipografia

Inter variable, incluida localmente (`frontend/public/inter-variable.woff2`,
licencia OFL). Razon: x-height alto y figuras tabulares, lo que mantiene legibles
las filas densas de letra a 14 px y evita que el reloj "baile" al cambiar.

| Rol | Tamano / peso | Notas |
|---|---|---|
| Titulo de vista | 21 px / 500, tracking -0.015em | un `h1` por vista |
| Titulo de seccion | 12.5 px / 600 | sin mayusculas espaciadas |
| Cuerpo | 14 px / 400 | line-height 1.45 |
| Letra inactiva | 15-19 px / 500 | escala con el estado activo |
| Letra activa | 19 px (34 px en inmersivo) | unico foco de lectura |
| Metadatos | 11-12.5 px / 400 | `muted` o `faint` |
| Reloj y duraciones | 10.5-11 px, `tnum` | figuras tabulares |

No se usan mayusculas con tracking amplio en etiquetas de seccion: la jerarquia
la dan peso y color, no el espaciado (regla antislop R-06).

## Espaciado y radios

- Escala base 4 px; padding lateral de vista 28 px; separacion entre secciones
  28 px con la linea de senal.
- Radios: 4 px chips, 7-8 px controles y filas, 14-16 px paneles. El unico
  elemento circular es el boton de reproducir.

## Motivo de identidad

**La linea de senal**: tres barras verticales en el logotipo, la linea de 1 px que
separa secciones y la barra de progreso. La misma forma aparece en el boton de
reproducir y en el relleno de la palabra activa, de modo que "donde estas en la
cancion" se lee con un unico lenguaje visual.

## Componentes

| Componente | Regla |
|---|---|
| `Button` | `primary` solo para la accion principal de la vista; `subtle` para acciones secundarias; `ghost` para terciarias |
| `IconButton` | nombre accesible obligatorio + tooltip; tamano minimo 28x28, 36x36 en transporte |
| `Slider` | riel de 2 px, pulgar visible en hover y foco; teclado con flechas, Shift para paso grande |
| `Cover` | portada real si el proveedor la da; si no, degradado determinista derivado del titulo con iniciales: es un marcador honesto, no una portada falsa |
| `Toast` | `role="status"`, se descarta solo a los 5.2 s, nunca roba foco |
| `EmptyState` | dice que falta y cual es el siguiente paso; siempre con accion cuando existe |
| `Modal` | Radix Dialog: foco atrapado, `Escape` cierra, overlay con blur de 2 px |

Glass (backdrop-blur) aparece en exactamente dos lugares: la barra del reproductor
y el pie del modo inmersivo. Sombra elevada, solo en esos mismos dos elementos y
en dialogos.

## Movimiento

- Transiciones de 120-200 ms en color, opacidad y transform.
- La letra se desplaza con `scroll-behavior: smooth`; el resaltado de la palabra
  activa se pinta con un gradiente recortado sobre texto (solo pintura, sin
  recalcular layout).
- `prefers-reduced-motion: reduce` desactiva animaciones y transiciones.
- El autoscroll se detiene 6 s cuando el usuario hace scroll manual y ofrece
  "Seguir la letra".

## Accesibilidad

- Foco visible siempre: contorno de 2 px `accent` con 2 px de separacion.
- Navegacion completa por teclado: `Tab` recorre navegacion, filas, palabras de
  la letra activa y controles; `Espacio` reproduce, flechas buscan +/-5 s, `M`
  silencia, `I` abre el modo inmersivo, `Escape` lo cierra.
- Las palabras de la letra son `role="button"` con nombre accesible ("Consultar
  slowly") y activables con Enter/Espacio.
- Los estados no dependen solo del color: el verso activo cambia de tamano y
  opacidad; los avisos llevan icono y texto.
- En modo estudio el texto de la letra activa llega a 34 px en inmersivo.
