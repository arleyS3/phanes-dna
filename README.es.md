<div align="center">

<h1>Phanes DNA</h1>

<p><strong>Phanes DNA — Arquitecto de Software Residente y Plataforma de Gobernanza de Contexto para Agentes IA.</strong></p>

<p>
<a href="LICENSE"><img src="https://img.shields.io/badge/Licencia-MIT-blue.svg" alt="Licencia: MIT"></a>
<img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go 1.25+">
<img src="https://img.shields.io/badge/Servidor_MCP-stdio-purple" alt="Servidor MCP">
<img src="https://img.shields.io/badge/plataforma-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey" alt="Plataforma">
</p>

</div>

---

## ¿Qué hace Phanes DNA?

**Phanes DNA** es una plataforma sin dependencias externas que actúa como un arquitecto de software residente. Descubre, indexa y comprime reglas arquitectónicas y árboles sintácticos (AST) de repositorios multi-stack, sirviéndolos a asistentes de desarrollo basados en IA a través del protocolo estándar Model Context Protocol (MCP).

**El Problema**: Los agentes de código pierden el contexto arquitectónico entre sesiones, violando los límites de capa (por ejemplo, llamando a consultas SQL directas en un controlador de UI o salteándose la capa de servicios).

**La Solución**: Phanes DNA analiza el repositorio, aplica reglas de gobernanza mediante almacenamiento vectorial en SQLite, comprime el contexto con el motor **Caveman Filtering** (ahorrando entre un 40% y 65% de tokens) y protege el código mediante Git Hooks y GitHub Actions.

---

## Lenguajes y Características Principales

### Stacks Compatibles

| Lenguaje | Motor de Parseo AST | Características |
| --- | --- | --- |
| **Java** | `go-tree-sitter` (gramática Java) | Árbol AST completo, anotaciones Spring Boot, clasificación de capas |
| **Go** | Nativo `go/parser` | Estructuras, interfaces, funciones, métodos, cero dependencias |
| **Python** | Escáner de RegEx y AST | Clases, funciones y heurísticas de capa (`.py`) |
| **TypeScript / JS** | Escáner de RegEx y AST | Clases ES6, funciones y rutas (`.ts`, `.tsx`, `.js`, `.jsx`) |

---

## Capacidades Clave

1. **Servidor MCP Stdio (`phanes-dna serve`)**:
   - Expone de forma nativa las herramientas `get_project_dna` y `review_architecture` a Claude Code, Cursor, Antigravity, OpenCode y Codex.
2. **Motor de Compresión Caveman**:
   - Elimina cortesías, conectores y artículos innecesarios del contexto arquitectónico antes de enviarlo a los LLMs, reduciendo el costo de tokens en los prompts hasta en un 65%.
3. **Local Cache & Sync Engine (bundles `.dna`)**:
   - Exporta las reglas del proyecto a un archivo comprimido ultraliviano `.dna` (~150 bytes) para sincronizar en Git y compartir límites arquitectónicos entre todos los desarrolladores del equipo.
4. **Integración con Git Hooks**:
   - Instala hooks pre-commit y pre-push (`phanes-dna hooks install`) para detener violaciones arquitectónicas en milisegundos antes de confirmar el código.
5. **Bot de CI para GitHub Actions**:
   - Ejecuta `phanes-dna review --ci` en Pull Requests para publicar anotaciones de errores de capa directamente en la interfaz de GitHub.
6. **Instalador en Un Comando (`phanes-dna setup`)**:
   - Configura automáticamente la entrada MCP stdio en los archivos de configuración de Antigravity, OpenCode, Claude Desktop y Cursor.

---

## Guía Rápida de Inicio

### Instalación Recomendada

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/arley/phanes-dna/main/scripts/install.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/arley/phanes-dna/main/scripts/install.ps1 | iex
```

### Compilación e Instalación desde el Código Fuente

Requiere **Go 1.25+**:

```bash
git clone https://github.com/arley/phanes-dna.git
cd phanes-dna
go build -o phanes-dna ./cmd/phanes-dna
```

---

## Interfaz Interactiva de Terminal (TUI)

Simplemente ejecuta `phanes-dna` sin argumentos para abrir el menú interactivo CLI:

```bash
phanes-dna
```

---

## Referencia de Comandos CLI

| Comando | Descripción |
| --- | --- |
| `phanes-dna` | Abre la interfaz de menú interactivo |
| `phanes-dna analyze [ruta]` | Escanea e indexa archivos del proyecto en la base SQLite local |
| `phanes-dna review [--strict] [--ci]` | Evalúa el cumplimiento arquitectónico y reglas de capa |
| `phanes-dna serve` | Inicia el servidor MCP stdio para agentes de IA |
| `phanes-dna export [salida.dna]` | Exporta las reglas a un archivo comprimido `.dna` |
| `phanes-dna import <archivo.dna>` | Importa un bundle `.dna` a la base de datos local |
| `phanes-dna setup [agente]` | Autoconfigura MCP para asistentes de IA (`claude`, `cursor`, `antigravity`, `opencode`, `all`) |
| `phanes-dna hooks install` | Instala hooks de Git pre-commit y pre-push |
| `phanes-dna doctor` | Ejecuta diagnósticos de salud del entorno y ecosistema |
| `phanes-dna version` | Muestra la versión del binario y el proveedor de IA activo |

---

## Configuración de Integración MCP

Ejecuta el instalador automático:

```bash
phanes-dna setup
```

O agrega manualmente la entrada al archivo JSON de MCP de tu agente:

```json
{
  "mcpServers": {
    "phanes-dna": {
      "command": "phanes-dna",
      "args": ["serve"]
    }
  }
}
```

---

## Licencia

Distribuido bajo la [Licencia MIT](LICENSE).
