# Terminal-based Text Editor with Integrated File Explorer and Terminal

This project is a terminal-based text editor written in Go. It provides a simple interface for file exploration, text editing, and terminal operations, all within a single application.

## Features

- File Explorer: Navigate through your project's directory structure
- Text Editor: Edit files with basic text editing capabilities
- Output Window: View program output and messages
- Integrated Terminal: Execute commands directly within the application
- Customizable Terminal: Adjust terminal colors to your preference

## Key Bindings

- `Ctrl+S`: Save the current file
- `Ctrl+Q`: Quit the application
- `Ctrl+T`: Focus on the terminal
- `Ctrl+E`: Focus on the editor
- `Ctrl+F`: Focus on the file explorer
- `Ctrl+C`: Customize terminal colors (when terminal is focused)

## Installation

1. Ensure you have Go installed on your system.
2. Clone this repository:
   ```
   git clone https://github.com/yourusername/terminal-text-editor.git
   ```
3. Navigate to the project directory:
   ```
   cd terminal-text-editor
   ```
4. Install the required dependencies:
   ```
   go mod tidy
   ```
5. Build the application:
   ```
   go build
   ```

## Usage

1. Run the application:
   ```
   ./terminal-text-editor
   ```
2. Use the file explorer to navigate and select files.
3. Edit files in the text editor.
4. Use the integrated terminal for command execution.
5. Customize the terminal appearance using the terminal customization feature.

## Dependencies

This project uses the following external libraries:

- [github.com/creack/pty](https://github.com/creack/pty)
- [github.com/gdamore/tcell/v2](https://github.com/gdamore/tcell)
- [github.com/rivo/tview](https://github.com/rivo/tview)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).
