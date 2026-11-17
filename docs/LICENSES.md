# Licencias

Resumen de las dependencias y de las obligaciones que implican para distribuir el
binario. Verificado sobre las versiones realmente instaladas (`go.mod`,
`frontend/package.json`).

## Aplicacion

| Componente | Version | Licencia | Obligacion |
|---|---|---|---|
| Wails | v3.1.0 | MIT | incluir aviso de copyright |
| Go standard library | 1.27 | BSD-3-Clause | incluir aviso |
| modernc.org/sqlite | v1.62.0 | BSD-3-Clause | incluir aviso |
| github.com/zalando/go-keyring | v0.2.8 | MIT | incluir aviso |
| github.com/dhowden/tag | ultima | BSD-2-Clause | incluir aviso |
| github.com/stretchr/testify | v1.12 (solo pruebas) | MIT | no se distribuye |

## Frontend

| Componente | Licencia |
|---|---|
| React / React DOM 19 | MIT |
| Vite 8 | MIT |
| Tailwind CSS 4 | MIT |
| Radix UI (`react-dialog`, `react-tooltip`) | MIT |
| Zustand 5 | MIT |
| Motion 12 | MIT |
| Lucide React | ISC |
| Zod 4 | MIT |
| clsx | MIT |

## Fuentes

| Recurso | Licencia | Nota |
|---|---|---|
| Inter variable (`inter-variable.woff2`) | SIL Open Font License 1.1 | se incluye en `frontend/public`; el aviso esta en la carpeta y debe acompanar al binario |

## Contenido de demostracion

Las tres canciones, sus letras y las transcripciones foneticas fueron escritas
para STEFANI y se distribuyen con el proyecto. El audio se sintetiza en tiempo de
ejecucion: no se incluye ninguna grabacion de terceros.

## Servicios externos (no redistribuidos)

| Servicio | Condiciones | Uso en la app |
|---|---|---|
| LRCLIB | API publica con politica de uso razonable; pide User-Agent identificable | una consulta por pista, cache local |
| YouTube Music (InnerTube) | terminos de servicio de YouTube; el cliente es no oficial | solo metadatos publicos; sin descarga de audio ni evasion de restricciones |
| Proveedor de IA | depende del proveedor elegido por el usuario | el usuario aporta su clave y acepta sus terminos |

## Licencias descartadas

| Componente | Motivo |
|---|---|
| **eSpeak NG** | GPLv3: redistribuir el binario obligaria a liberar la obra completa o aislar el proceso; se prefirio no incluirlo y generar el IPA con el proveedor de IA |
| **Braccato** | MIT y compatible, pero se opto por un renderizador propio para controlar tipografia, accesibilidad y estados; la dependencia no se incorporo |
| **Liner** | referencia visual unicamente: no se copio codigo, shaders, iconos ni branding |

## Marcas

STEFANI es un nombre de trabajo. No se usa ninguna marca de terceros en la
interfaz ni en los instaladores. El icono de la aplicacion es el que trae la
plantilla de Wails y queda pendiente de diseno propio: se documenta como
provisional, no como identidad final.

## yt-dlp (herramienta externa, no distribuida)

`yt-dlp` es software libre (Unlicense). STEFANI **no** lo incluye en el
instalador ni lo descarga: solo lo invoca si el usuario lo tiene instalado. Por
eso no se agrega a la lista de dependencias distribuidas, pero se documenta aqui
porque el usuario debe instalarlo para la musica en linea.

## CMU Pronouncing Dictionary

`internal/phonetics/data/cmudict.dict` se distribuye dentro del binario.

Copyright (C) 1993-2015 Carnegie Mellon University. All rights reserved.
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the copyright notice and this
permission notice are retained. El archivo embebido conserva su cabecera
original.
