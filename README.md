# cvyml

Generador de currículums en PDF a partir de un archivo YAML, con un diseño
clásico de una sola columna al estilo Harvard.

`cvyml` lee un archivo `cv.yml`, valida sus datos y produce un PDF en formato
carta (o A4) con tipografía serif, encabezado de contacto con iconos, secciones
con reglas horizontales y entradas con fechas alineadas a la derecha. El binario
es autónomo: las fuentes y los iconos van embebidos, por lo que no requiere
archivos externos ni conexión a red para generar el PDF.

## Características

- **Entrada declarativa en YAML**: separas el contenido (`cv.yml`) del formato.
- **Salida PDF determinista**: dos ejecuciones con la misma entrada producen
  bytes idénticos, ideal para control de versiones y CI.
- **Enlaces clicables**: correos, perfiles y proyectos quedan como enlaces
  reales, mientras en papel se muestra una versión acortada y legible.
- **Iconos opcionales**: glifos monocromos (Bootstrap Icons) junto a los datos
  de contacto, activables o desactivables.
- **Orden de secciones configurable** y válido para cualquier combinación.
- **Validación estricta**: los campos desconocidos o mal escritos producen un
  error en lugar de perder datos silenciosamente.
- **Fechas con precisión parcial**: acepta `YYYY`, `YYYY-MM` y `YYYY-MM-DD`, y
  soporta rangos abiertos ("Present").
- **Fuentes embebidas**: Tinos (compatible en métricas con Times New Roman) con
  soporte completo de UTF-8.

## Requisitos

- Go 1.26 o superior.

## Instalación

Con el gestor de paquetes de Go (recomendado):

```sh
go install github.com/miguehm/yml-to-cv/cmd/cvyml@latest
```

El binario `cvyml` queda en `$(go env GOPATH)/bin` (normalmente en el `PATH`).

Desde el código fuente:

```sh
git clone https://github.com/miguehm/yml-to-cv.git
cd cvyml
make build
```

El binario queda en `bin/cvyml`.

## Uso

### Línea de comandos

```sh
# Genera cv.pdf a partir de cv.yml (valores por defecto)
./bin/cvyml

# Rutas y papel explícitos
./bin/cvyml -in cv.yml -out salida.pdf -paper a4

# Margen personalizado en milímetros
./bin/cvyml -margin 15

# Imprime el CV ya parseado/normalizado como YAML y termina
./bin/cvyml -dump
```

Banderas disponibles:

| Bandera   | Valor por defecto | Descripción                                              |
| --------- | ----------------- | -------------------------------------------------------- |
| `-in`     | `cv.yml`          | Ruta del archivo YAML de entrada.                        |
| `-out`    | `cv.pdf`          | Ruta del PDF de salida.                                   |
| `-paper`  | `letter`          | Tamaño de papel: `letter` o `a4`.                         |
| `-margin` | `0`               | Margen de página en mm (`0` usa 19.05 mm = 0.75 in).      |
| `-dump`   | `false`           | Imprime el CV parseado como YAML y sale sin generar PDF.  |

### Makefile

```sh
make build    # compila el binario en bin/cvyml
make install  # go install ./cmd/cvyml (deja cvyml en el GOPATH)
make wasm     # compila el módulo WebAssembly en bin/cvyml.wasm
make render   # go run ... -in cv.yml -out cv.pdf
make run      # compila y luego ejecuta el binario
make test     # ejecuta las pruebas
make fmt      # gofmt -w .
make vet      # go vet ./...
make check    # fmt + vet + test
make clean    # elimina bin/ y cv.pdf
```

## Funcionamiento

El núcleo (`internal/cvyml`) es independiente de la plataforma: recibe los bytes
de entrada y un `io.Writer` de salida, y no toca el sistema de archivos ni el
terminal. La CLI nativa (`internal/cli` + `cmd/cvyml`) solo aporta el parseo de
banderas y la lectura/escritura de ficheros; un futuro binario WebAssembly
(`cmd/cvyml-wasm`) reutiliza exactamente el mismo núcleo.

El flujo es lineal y consta de tres etapas:

1. **Parseo y validación** (`internal/model`): el YAML se decodifica con
   `KnownFields(true)`, de modo que cualquier clave desconocida es un error. Se
   valida que `basics.name` no esté vacío, que los nombres de las secciones de
   `section_order` sean válidos y que los iconos referenciados existan. Después
   se normalizan los datos: las URLs sin esquema reciben `https://` y el resumen
   se recorta.
2. **Renderizado** (`internal/render`): el `Renderer` de estilo Harvard calcula
   el ancho del contenido, dibuja el encabezado (nombre, etiqueta y línea de
   contacto con iconos y enlaces centrados, con salto automático de línea) y
   recorre `section_order`, saltando las secciones sin contenido. Cada tipo de
   entrada (educación, experiencia, proyecto, aptitud, logro) tiene su propia
   función de dibujo.
3. **Salida PDF**: el documento se escribe directamente al `io.Writer` de salida
   mediante `go-pdf/fpdf`. La fecha de creación se fija por defecto en
   `2000-01-01` y se activa `SetCatalogSort` para que el PDF resultante sea
   reproducible byte a byte.

La estructura del proyecto es:

```
cmd/cvyml/          CLI nativa (instalable con go install).
cmd/cvyml-wasm/     Punto de entrada WebAssembly (js/wasm).
internal/cli/       Banderas, I/O de ficheros y código de salida de la CLI.
internal/cvyml/     Núcleo portable: parseo, volcado y render (sin I/O).
internal/model/     Esquema del CV, parseo YAML y tipo PartialDate.
internal/render/    Renderizador Harvard y utilidades de texto/enlaces.
internal/icons/     Registro de iconos y su mapeo a puntos de código.
assets/fonts/       Fuentes Tinos embebidas (OFL).
assets/icons/       Tipo de letra Bootstrap Icons embebida (MIT).
```

## WebAssembly

El comando `cmd/cvyml-wasm` compila el mismo núcleo a un módulo `js/wasm` que
expone un objeto global `cvyml` a JavaScript:

```sh
make wasm   # genera bin/cvyml.wasm
```

En el navegador, carga `wasm_exec.js` (el que distribuye tu versión de Go, en
`$(go env GOROOT)/misc/wasm/`) y el módulo, y luego:

```js
const go = new Go();
const { instance } = await WebAssembly.instantiateStreaming(
  fetch("cvyml.wasm"), go.importObject
);
go.run(instance);

const { pdf, error } = cvyml.render(yamlText, { paper: "a4", margin: 15 });
if (error) throw new Error(error);
// `pdf` es un Uint8Array con el PDF listo para descargar.
```

API expuesta:

| Función                  | Devuelve            | Descripción                                        |
| ------------------------ | ------------------- | -------------------------------------------------- |
| `cvyml.render(yaml, opts)` | `{ pdf, error }`  | Genera el PDF. `pdf` es un `Uint8Array`.           |
| `cvyml.parse(yaml)`        | `{ yaml, error }` | Devuelve el CV validado y normalizado como YAML.   |

`opts` es opcional y puede ser una cadena (`"letter"` o `"a4"`) o un objeto
`{ paper, margin }`. Las funciones no lanzan excepciones: ante un error
devuelven el objeto con `error` establecido y el campo de carga útil a `null`.

## Esquema del YAML

### `basics` (encabezado)

| Campo      | Tipo      | Obligatorio | Descripción                                                        |
| ---------- | --------- | ----------- | ------------------------------------------------------------------ |
| `name`     | string    | sí          | Nombre completo.                                                   |
| `label`    | string    | no          | Titular o rol, mostrado bajo el nombre.                            |
| `location` | string    | no          | Ubicación.                                                          |
| `email`    | string    | no          | Correo; se convierte en enlace `mailto:`.                          |
| `phone`    | string    | no          | Teléfono.                                                           |
| `summary`  | string    | no          | Resumen profesional (acepta texto multilínea con `>-`).            |
| `icons`    | bool      | no          | `false` desactiva todos los iconos del encabezado. Por defecto `true`. |
| `profiles` | list      | no          | Lista de enlaces (ver abajo).                                      |

Cada elemento de `profiles`:

| Campo      | Tipo   | Obligatorio | Descripción                                                          |
| ---------- | ------ | ----------- | -------------------------------------------------------------------- |
| `network`  | string | sí          | Nombre de la red o sitio (`GitHub`, `LinkedIn`, `Website`, …).       |
| `username` | string | no          | Usuario (informativo).                                               |
| `url`      | string | no          | Enlace; se le añade `https://` si falta el esquema.                  |
| `icon`     | string | no          | Fuerza un icono concreto (ver [Iconos](#iconos)).                     |

### `section_order`

Lista que define el orden y la presencia de las secciones. Los valores válidos
son `education`, `experience`, `projects`, `skills` y `achievements`. Si se
omite, se usa este orden por defecto:

```yaml
section_order: [education, experience, projects, skills, achievements]
```

Las secciones sin datos se omiten automáticamente.

### `education`

| Campo         | Tipo     | Descripción                                      |
| ------------- | -------- | ------------------------------------------------ |
| `institution` | string   | Institución.                                     |
| `degree`      | string   | Grado (p. ej. `BS`).                             |
| `area`        | string   | Área de estudio. Se combina como `degree in area`. |
| `location`    | string   | Ubicación.                                        |
| `start`       | fecha    | Inicio.                                           |
| `end`         | fecha    | Fin (vacío = "Present").                          |
| `highlights`  | [string] | Viñetas descriptivas.                            |

### `experience`

| Campo        | Tipo     | Descripción                |
| ------------ | -------- | -------------------------- |
| `company`    | string   | Empresa.                   |
| `position`   | string   | Puesto.                    |
| `location`   | string   | Ubicación.                 |
| `start`      | fecha    | Inicio.                    |
| `end`        | fecha    | Fin (vacío = "Present").   |
| `highlights` | [string] | Viñetas de responsabilidades. |

### `projects`

| Campo        | Tipo     | Descripción                                            |
| ------------ | -------- | ------------------------------------------------------ |
| `name`       | string   | Nombre del proyecto.                                   |
| `url`        | string   | Enlace; se añade `https://` si falta el esquema.       |
| `date`       | fecha    | Fecha del proyecto.                                    |
| `summary`    | string   | Descripción en párrafo.                                |
| `highlights` | [string] | Viñetas adicionales.                                   |

### `skills`

| Campo     | Tipo   | Descripción                                     |
| --------- | ------ | ----------------------------------------------- |
| `label`   | string | Etiqueta del grupo (se le añade `:` al final).  |
| `details` | string | Tecnologías o competencias, envueltas si hace falta. |

### `achievements`

| Campo        | Tipo     | Descripción                    |
| ------------ | -------- | ------------------------------ |
| `name`       | string   | Nombre del logro o premio.     |
| `date`       | fecha    | Fecha.                         |
| `highlights` | [string] | Viñetas descriptivas.          |

## Fechas

El campo `PartialDate` acepta precisión parcial:

| Formato       | Ejemplo      | Se muestra como     |
| ------------- | ------------ | ------------------- |
| `YYYY`        | `2020`       | `2020`              |
| `YYYY-MM`     | `2020-08`    | `Aug 2020`          |
| `YYYY-MM-DD`  | `2022-08-11` | `Aug 11, 2022`      |

Reglas de los rangos:

- `start` y `end` iguales colapsan a una sola fecha.
- Un `end` vacío se renderiza como `– Present`.
- Si tanto `start` como `end` comparten el mismo año y no tienen mes, se
  muestra el año una sola vez.
- Las combinaciones imposibles (p. ej. `2024-02-31`) son un error.

## Enlaces y URLs

- Los enlaces en el cuerpo usan sintaxis Markdown `[texto](url)`, tanto en los
  títulos de las entradas como en los nombres de proyectos.
- Si un proyecto tiene `url` y su nombre no contiene ya un enlace, se añade
  automáticamente un enlace visible con la URL acortada.
- En el papel se muestra la URL sin esquema ni `www.` (p. ej.
  `github.com/usuario`), pero el destino del enlace sigue siendo la URL
  completa, de modo que el PDF es clicable en pantalla y legible impreso.

## Iconos

Los iconos del encabezado se eligen a partir de `network`, con reserva a un
globo terráqueo para redes desconocidas. Puedes forzar uno distinto con el campo
`icon` del perfil. Nombres disponibles:

| Nombre      | Uso típico         |
| ----------- | ------------------ |
| `github`    | GitHub             |
| `linkedin`  | LinkedIn           |
| `globe`     | Sitio web / otros  |
| `envelope`  | Correo             |
| `telephone` | Teléfono           |
| `geo-alt`   | Ubicación          |
| `link`      | Enlace genérico    |

Ejemplo de anulación de icono:

```yaml
profiles:
  - network: Website
    url: "miguehm.com"
    icon: "link"
```

## Ejemplo

```yaml
basics:
  name: "Jane Doe"
  label: "Backend Developer"
  location: "Puebla, Mexico"
  email: "jane@example.com"
  phone: "221-400-9504"
  summary: >-
    Desarrolladora backend con experiencia en APIs REST y PostgreSQL.
  profiles:
    - network: GitHub
      username: "jane"
      url: "github.com/jane"

section_order: [experience, education, skills]

experience:
  - company: "Acme Corp"
    position: "Backend Developer"
    location: "Remote"
    start: "2024-06"
    end: "2025-01"
    highlights:
      - "Construyó APIs REST con Go y PostgreSQL."

education:
  - institution: "Example University"
    degree: "BS"
    area: "Computer Science"
    start: "2020-08"
    end: "2024-05"
    highlights:
      - "Cursos de algoritmos y estructuras de datos."

skills:
  - label: "Languages"
    details: "Go, Python, JavaScript"
```

Genera el PDF con:

```sh
make render
```

## Desarrollo

```sh
go test ./...   # pruebas de modelo, iconos y renderizado
make check      # formato, vet y pruebas
```

Las pruebas del paquete `internal/render` descomprimen los flujos de contenido
del PDF y verifican propiedades concretas: que la salida sea determinista, que
las fechas queden alineadas a la derecha, que los iconos compartan línea base
con su texto y que la línea de contacto haga salto de línea cuando no cabe.

## Licencia y atribuciones

- **Tinos** (`assets/fonts`): SIL Open Font License. Compatible en métricas con
  Times New Roman.
- **Bootstrap Icons** (`assets/icons`): MIT License; ver
  `assets/icons/LICENSE`. El archivo TrueType se convirtió desde la
  distribución oficial WOFF de `bootstrap-icons@1.13.2`.
