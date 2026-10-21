# cvyml

Generador de currículums en PDF a partir de un archivo YAML, con un diseño
clásico de una sola columna al estilo Harvard.

`cvyml` lee un archivo `cv.yml`, valida sus datos y produce un PDF en formato
carta (o A4) con tipografía serif, encabezado de contacto con iconos, resumen
profesional opcional, secciones con reglas horizontales y entradas con fechas
alineadas a la derecha. El binario es autónomo: las fuentes y los iconos van
embebidos, por lo que no requiere archivos externos ni conexión a red para
generar el PDF.

## Características

- **Entrada declarativa en YAML**: separas el contenido (`cv.yml`) del formato.
- **Resumen profesional**: `basics.summary` se dibuja como párrafo justificado
  bajo los datos de contacto.
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
cd yml-to-cv
make build
```

El binario queda en `bin/cvyml`.

## Inicio rápido

Crea un `cv.yml` mínimo:

```yaml
basics:
  name: "Jane Doe"
  label: "Backend Developer"
  email: "jane@example.com"
  summary: >-
    Desarrolladora backend con experiencia en APIs REST y PostgreSQL.
```

Valida e inspecciona la entrada normalizada, y después genera el PDF:

```sh
cvyml -in cv.yml -dump
cvyml -in cv.yml -out cv.pdf
```

El segundo comando imprime `PDF generado: cv.pdf`. Para cambiar el papel o el
margen:

```sh
cvyml -in cv.yml -out cv.pdf -paper a4 -margin 15
```

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
   el ancho del contenido, dibuja el encabezado (nombre, etiqueta, línea de
   contacto con iconos y enlaces centrados, con salto automático de línea, y el
   resumen profesional como párrafo justificado) y recorre `section_order`,
   saltando las secciones sin contenido. Cada tipo de entrada (educación,
   experiencia, proyecto, aptitud, logro) tiene su propia función de dibujo.
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

## Del YAML al PDF

Cada bloque del YAML se traduce a una zona concreta del documento:

| Bloque YAML                                   | Resultado en el PDF                                                                 |
| --------------------------------------------- | ----------------------------------------------------------------------------------- |
| `basics.name`                                 | Nombre completo centrado y en negrita.                                              |
| `basics.label`                                | Titular o rol centrado en cursiva bajo el nombre.                                   |
| `basics.phone`, `email`, `location`, `profiles` | Línea de contacto centrada con iconos y enlaces; se reparte en varias líneas si no cabe. |
| `basics.summary`                              | Párrafo justificado bajo la línea de contacto.                                      |
| `section_order`                               | Orden y presencia de las secciones; cada una se dibuja con su título en mayúsculas y una regla horizontal. |
| `education`                                   | Entrada con grado y área, fechas a la derecha, institución/ubicación en cursiva y viñetas. |
| `experience`                                  | Entrada con puesto, fechas a la derecha, empresa/ubicación en cursiva y viñetas.     |
| `projects`                                    | Nombre con enlace acortado, fecha, resumen en párrafo y viñetas.                     |
| `skills`                                      | Viñeta con etiqueta en negrita seguida de los detalles.                             |
| `achievements`                                | Nombre del logro, fecha a la derecha y viñetas.                                     |

Las secciones que no aparecen en `section_order`, o que existen pero están
vacías, se omiten sin dejar huecos. Los títulos de sección se muestran siempre en
mayúsculas.

### Valores por defecto del PDF

| Ajuste            | Valor por defecto                                  | Cómo cambiarlo                    |
| ----------------- | -------------------------------------------------- | --------------------------------- |
| Papel             | `letter`                                           | `-paper a4`                       |
| Margen            | 19.05 mm (0.75 in)                                 | `-margin <mm>`                    |
| Tipografía        | Tinos (métricas compatibles con Times New Roman)  | No configurable desde la CLI.     |
| Cuerpo            | 10 pt                                              | No configurable desde la CLI.     |
| Nombre            | 20 pt                                              | No configurable desde la CLI.     |
| Títulos de sección | 11.5 pt                                           | No configurable desde la CLI.     |
| Fechas            | 9.5 pt                                             | No configurable desde la CLI.     |
| Metadatos         | Título `<nombre> - CV`, autor y fecha fija         | Automático a partir de `basics`.  |

La fecha de creación se fija en `2000-01-01` para que el PDF sea reproducible
byte a byte. Las fuentes y los iconos van embebidos en el binario.

## Esquema del YAML

### Plantilla completa

Esta plantilla reúne todas las claves que acepta el programa. Copia solo las
secciones que necesites; las que elimines se omiten del PDF:

```yaml
basics:
  name: "Jane Doe"                    # obligatorio
  label: "Backend Developer"
  location: "Puebla, Mexico"
  email: "jane@example.com"
  phone: "221-400-9504"
  summary: >-
    Desarrolladora backend con experiencia en APIs REST y PostgreSQL.
  icons: true                         # false oculta los iconos del encabezado
  profiles:
    - network: "GitHub"
      username: "jane"
      url: "github.com/jane"          # se normaliza a https://
      icon: "github"                  # opcional; ver sección Iconos

section_order: [education, experience, projects, skills, achievements]

education:
  - institution: "Example University"
    degree: "BS"
    area: "Computer Science"
    location: "Puebla, Mexico"
    start: "2020-08"
    end: "2024-05"
    highlights:
      - "Cursos de algoritmos y estructuras de datos."

experience:
  - company: "Acme Corp"
    position: "Backend Developer"
    location: "Remote"
    start: "2024-06"
    end: ""                           # vacío = Present
    highlights:
      - "Construyó APIs REST con Go y PostgreSQL."

projects:
  - name: "ContraScripts"
    url: "contrascripts.example.com"
    date: "2026-06"
    summary: "Aplicación web para escribir guiones cinematográficos."
    highlights:
      - "Usa el formato de texto plano Fountain."

skills:
  - label: "Languages"
    details: "Go, Python, JavaScript"

achievements:
  - name: "Segundo lugar - Concurso de Ciencia Ficción"
    date: "2022-08-11"
    highlights:
      - "Reconocimiento por una narración presentada en un concurso universitario."
```

### `basics` (encabezado)

| Campo      | Tipo      | Obligatorio | Descripción                                                        |
| ---------- | --------- | ----------- | ------------------------------------------------------------------ |
| `name`     | string    | sí          | Nombre completo.                                                   |
| `label`    | string    | no          | Titular o rol, mostrado bajo el nombre.                            |
| `location` | string    | no          | Ubicación.                                                          |
| `email`    | string    | no          | Correo; se convierte en enlace `mailto:`.                          |
| `phone`    | string    | no          | Teléfono.                                                           |
| `summary`  | string    | no          | Párrafo justificado bajo el contacto (acepta multilínea con `>-`). |
| `icons`    | bool      | no          | `false` desactiva todos los iconos del encabezado. Por defecto `true`. |
| `profiles` | list      | no          | Lista de enlaces (ver abajo).                                      |

Cada elemento de `profiles`:

| Campo      | Tipo   | Obligatorio | Descripción                                                          |
| ---------- | ------ | ----------- | -------------------------------------------------------------------- |
| `network`  | string | sí          | Nombre de la red o sitio (`GitHub`, `LinkedIn`, `Website`, …).       |
| `username` | string | no          | Usuario (informativo).                                               |
| `url`      | string | no          | Enlace; se le añade `https://` si falta el esquema.                  |
| `icon`     | string | no          | Fuerza un icono concreto (ver [Iconos](#iconos)).                     |

La validación estricta solo exige `basics.name`; el resto de campos marcados
como obligatorios son necesarios para que cada entrada se muestre con sentido.
Los campos opcionales vacíos se omiten del PDF.

### `section_order`

Lista que define el orden y la presencia de las secciones. Los valores válidos
son `education`, `experience`, `projects`, `skills` y `achievements`. Si se
omite, se usa este orden por defecto:

```yaml
section_order: [education, experience, projects, skills, achievements]
```

Las secciones sin datos se omiten automáticamente.

### `education`

| Campo         | Tipo     | Obligatorio | Descripción / efecto en el PDF                        |
| ------------- | -------- | ----------- | ----------------------------------------------------- |
| `institution` | string   | sí          | Institución, dibujada en cursiva.                     |
| `degree`      | string   | no          | Grado (p. ej. `BS`).                                  |
| `area`        | string   | no          | Área de estudio. Se combina como `degree in area`.    |
| `location`    | string   | no          | Ubicación; se añade a la institución separada por coma. |
| `start`       | fecha    | no          | Inicio del rango de fechas.                           |
| `end`         | fecha    | no          | Fin del rango (vacío = "Present").                    |
| `highlights`  | [string] | no          | Viñetas descriptivas bajo la entrada.                 |

### `experience`

| Campo        | Tipo     | Obligatorio | Descripción / efecto en el PDF                |
| ------------ | -------- | ----------- | --------------------------------------------- |
| `company`    | string   | sí          | Empresa, dibujada en cursiva.                 |
| `position`   | string   | sí          | Puesto, usado como título de la entrada.      |
| `location`   | string   | no          | Ubicación; se añade a la empresa con una coma. |
| `start`      | fecha    | no          | Inicio del rango de fechas.                   |
| `end`        | fecha    | no          | Fin del rango (vacío = "Present").            |
| `highlights` | [string] | no          | Viñetas de responsabilidades.                 |

### `projects`

| Campo        | Tipo     | Obligatorio | Descripción / efecto en el PDF                        |
| ------------ | -------- | ----------- | ----------------------------------------------------- |
| `name`       | string   | sí          | Nombre del proyecto; se muestra con su enlace.        |
| `url`        | string   | no          | Enlace; se añade `https://` si falta el esquema.      |
| `date`       | fecha    | no          | Fecha del proyecto, alineada a la derecha.            |
| `summary`    | string   | no          | Descripción en párrafo justificado.                   |
| `highlights` | [string] | no          | Viñetas adicionales.                                  |

### `skills`

| Campo     | Tipo   | Obligatorio | Descripción / efecto en el PDF                        |
| --------- | ------ | ----------- | ----------------------------------------------------- |
| `label`   | string | sí          | Etiqueta del grupo (se le añade `:` al final).        |
| `details` | string | sí          | Tecnologías o competencias; se envuelven si hace falta. |

### `achievements`

| Campo        | Tipo     | Obligatorio | Descripción / efecto en el PDF        |
| ------------ | -------- | ----------- | ------------------------------------- |
| `name`       | string   | sí          | Nombre del logro o premio.            |
| `date`       | fecha    | no          | Fecha, alineada a la derecha.         |
| `highlights` | [string] | no          | Viñetas descriptivas.                 |

### Reglas de formato YAML

- **Indentación**: usa dos espacios y mantén alineados los campos hermanos. No
  mezcles tabuladores con espacios.
- **Cadenas**: entre comillas cuando contengan `:`, `#`, corchetes, llaves o
  caracteres que YAML podría interpretar como sintaxis.
- **Texto multilínea**: `>-` pliega los saltos de línea en un solo párrafo, ideal
  para `summary`, `details` o `highlights` largos.
- **Listas**: cada elemento empieza con `-` y conserva la misma indentación.
- **Fechas**: escríbelas entre comillas (`"2020-08"`) para evitar ambigüedades.
- **URLs**: las comillas también son recomendables; si falta el esquema se añade
  `https://` automáticamente.
- **Claves desconocidas**: cualquier clave no documentada aquí es un error, no se
  ignora silenciosamente.
- **Secciones vacías**: si una lista no tiene elementos o no se incluye en el
  YAML, su sección no se dibuja.

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
    Desarrolladora backend con experiencia en APIs REST y PostgreSQL. Combina
    el desarrollo de servicios con la automatización de despliegues.
  icons: true
  profiles:
    - network: GitHub
      username: "jane"
      url: "github.com/jane"
    - network: Website
      url: "jane.dev"
      icon: "globe"

section_order: [experience, education, projects, skills, achievements]

experience:
  - company: "Acme Corp"
    position: "Backend Developer"
    location: "Remote"
    start: "2024-06"
    end: "2025-01"
    highlights:
      - "Construyó APIs REST con Go y PostgreSQL."
      - "Automatizó el despliegue con GitHub Actions."

education:
  - institution: "Example University"
    degree: "BS"
    area: "Computer Science"
    location: "Puebla, Mexico"
    start: "2020-08"
    end: "2024-05"
    highlights:
      - "Cursos de algoritmos y estructuras de datos."

projects:
  - name: "ContraScripts"
    url: "contrascripts.example.com"
    date: "2026-06"
    summary: "Aplicación web para escribir guiones cinematográficos."
    highlights:
      - "Usa el formato de texto plano Fountain."

skills:
  - label: "Languages"
    details: "Go, Python, JavaScript"
  - label: "Tools"
    details: "Git, Docker, Linux, GitHub Actions"

achievements:
  - name: "Segundo lugar - Concurso de Ciencia Ficción"
    date: "2022-08-11"
    highlights:
      - "Reconocimiento por una narración presentada en un concurso universitario."
```

Guarda el contenido en `cv.yml` y genera el PDF con:

```sh
cvyml -in cv.yml -out cv.pdf
```

## Desarrollo

```sh
go test ./...   # pruebas de modelo, iconos y renderizado
make check      # formato, vet y pruebas
```

Las pruebas del paquete `internal/render` descomprimen los flujos de contenido
del PDF y verifican propiedades concretas: que la salida sea determinista, que
el resumen profesional se dibuje en el documento, que las fechas queden
alineadas a la derecha, que los iconos compartan línea base con su texto y que
la línea de contacto haga salto de línea cuando no cabe.

## Licencia y atribuciones

- **Tinos** (`assets/fonts`): SIL Open Font License. Compatible en métricas con
  Times New Roman.
- **Bootstrap Icons** (`assets/icons`): MIT License; ver
  `assets/icons/LICENSE`. El archivo TrueType se convirtió desde la
  distribución oficial WOFF de `bootstrap-icons@1.13.2`.
