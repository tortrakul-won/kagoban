# 🗂️ Kagoban - Your Terminal Kanban Board

<div align="center">

![Kagoban Logo](https://github.com/user-attachments/assets/3181e1f9-0924-4500-8f42-ba73b13c0bf6)

_A terminal-based Kanban board application built in Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework and [Lip Gloss](https://github.com/charmbracelet/lipgloss) styling._ 🎨

[![Go Version](https://img.shields.io/github/go-mod/go-version/tortrakul-won/kagoban)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Purpose](https://img.shields.io/badge/purpose-education-orange.svg)]()

</div>

## Status & Roadmap

<img width="647" alt="image" src="https://github.com/user-attachments/assets/9efdfa89-8d32-4690-8dc5-bad42e8952cd" />

### 🎯 Status

**✅ Shipped & Ready**

- 📝 Full note management (CRUD)
- 📊 Full Section management (CRUD)
- ⌨️ Intuitive keyboard navigation
- 💾 Persistent storage (JSON)
- 🔀 Advanced reordering capabilities
- ↔️ Cross-section movement

**🚧 Under Construction**

- 📂 Project structure refinement
- 📜 Scrollable viewport

## Keyboard Controls

| Key           | Action                            |
| ------------- | --------------------------------- |
| `←` `h`       | Move left between sections        |
| `→` `l`       | Move right between sections       |
| `↑` `k`       | Move up within section            |
| `↓` `j`       | Move down within section          |
| `a`           | Add new note                      |
| `e`           | Edit selected note                |
| `d`           | Delete selected note              |
| `A`           | Add new section                   |
| `E`           | Edit section name                 |
| `D`           | Delete section                    |
| `Space/Enter` | Toggle note completion            |
| `Ctrl+s`      | Save current state                |
| `Alt+←`       | Move note to the previous section |
| `Alt+→`       | Move note to the next section     |
| `Alt+↑`       | Move note upward                  |
| `Alt+↓`       | Move note downward                |
| `Alt+Shift+←` | Move section to the left          |
| `Alt+Shift+→` | Move section to the rgith         |
| `q`           | Quit application                  |

## Installation

```bash
# Clone the repository
git clone https://github.com/tortrakul-won/kagoban

# Navigate to project directory
cd kagoban

# Install dependencies
go mod tidy

# just run the main file
go run .
```

## 🛠️ Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - Common UI components

## Project Structure

```
.
├── README.md
├── data
│   └── save_file.json
├── go.mod
├── go.sum
├── main.go
├── model.go
├── operation.go
├── style.go
└── utils.go
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📜 License

MIT Licensed - do your thing! See [LICENSE](LICENSE) for details.

## 👏 Acknowledgments

- 💖 [Charm](https://charm.sh) for the amazing TUI tools
- 🐹 Go community for being awesome

---

**Note**: This is an open-source project. Feel free to contribute and report issues!
