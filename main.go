package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Constants for key bindings and colors
const (
	KeySave              = tcell.KeyCtrlS
	KeyQuit              = tcell.KeyCtrlQ
	KeyFocusTerminal     = tcell.KeyCtrlT
	KeyFocusEditor       = tcell.KeyCtrlE
	KeyFocusFileExplorer = tcell.KeyCtrlF
	KeyCustomizeTerminal = tcell.KeyCtrlA

	ColorGreen = tcell.ColorGreen
)

// UI represents the main UI components
type UI struct {
	app          *tview.Application
	root         *tview.Flex
	fileExplorer *tview.TreeView
	editor       *tview.TextArea
	output       *tview.TextView
	terminal     *tview.TextView
}

// TerminalState represents the state of the terminal
type TerminalState struct {
	pty  *os.File
	cmd  *exec.Cmd
	done chan struct{}
}

var (
	ui          UI
	termState   TerminalState
	currentFile string
)

func main() {
	var err error
	ui.app = tview.NewApplication()

	if err = createUI(); err != nil {
		log.Fatalf("Failed to create UI: %v", err)
	}

	if err = setupKeyBindings(); err != nil {
		log.Fatalf("Failed to set up key bindings: %v", err)
	}

	if err = ui.app.SetRoot(ui.root, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

// createUI initializes and sets up the user interface components
func createUI() error {
	ui.root = tview.NewFlex().SetDirection(tview.FlexRow)

	menuBar := createMenuBar()
	ui.root.AddItem(menuBar, 1, 0, false)

	content := tview.NewFlex().SetDirection(tview.FlexColumn)

	var err error
	ui.fileExplorer, err = createFileExplorer()
	if err != nil {
		return fmt.Errorf("failed to create file explorer: %w", err)
	}
	content.AddItem(ui.fileExplorer, 30, 0, true)

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	ui.editor = createEditor()
	ui.output = createOutput()
	ui.terminal, err = createTerminal()
	if err != nil {
		return fmt.Errorf("failed to create terminal: %w", err)
	}
	rightPanel.AddItem(ui.editor, 0, 2, false)
	rightPanel.AddItem(ui.output, 0, 1, false)
	rightPanel.AddItem(ui.terminal, 0, 1, false)

	content.AddItem(rightPanel, 0, 1, false)

	ui.root.AddItem(content, 0, 1, true)

	return nil
}

// setupKeyBindings configures the global key bindings for the application
func setupKeyBindings() error {
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case KeySave:
			if err := saveFile(); err != nil {
				ui.output.SetText(fmt.Sprintf("Error saving file: %s", err))
			}
			return nil
		case KeyQuit:
			ui.app.Stop()
			return nil
		case KeyFocusTerminal:
			ui.app.SetFocus(ui.terminal)
			return nil
		case KeyFocusEditor:
			ui.app.SetFocus(ui.editor)
			return nil
		case KeyFocusFileExplorer:
			ui.app.SetFocus(ui.fileExplorer)
			return nil
		case KeyCustomizeTerminal:
			if ui.app.GetFocus() == ui.terminal {
				customizeTerminal()
				return nil
			}
		}
		return event
	})
	return nil
}

// createMenuBar creates and returns the menu bar component
func createMenuBar() *tview.TextView {
	menuBar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	menuText := `[yellow]Ctrl+S[-] Save   [yellow]Ctrl+Q[-] Quit   [yellow]Ctrl+T[-] Terminal   [yellow]Ctrl+E[-] Editor   [yellow]Ctrl+F[-] Files   [yellow]Ctrl+C[-] Customize Terminal`
	menuBar.SetText(menuText)

	return menuBar
}

// createFileExplorer creates and returns the file explorer component
func createFileExplorer() (*tview.TreeView, error) {
	root := tview.NewTreeNode(".").
		SetColor(ColorGreen)
	if err := populateTree(root, "."); err != nil {
		return nil, fmt.Errorf("failed to populate tree: %w", err)
	}

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}
		path := reference.(string)
		if err := loadFile(path); err != nil {
			ui.output.SetText(fmt.Sprintf("Error loading file: %s", err))
		}
	})

	return tree, nil
}

// populateTre recursively populates the file explorer tree
func populateTree(node *tview.TreeNode, path string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	for _, file := range files {
		child := tview.NewTreeNode(file.Name()).
			SetSelectable(true)
		if file.IsDir() {
			child.SetColor(ColorGreen)
			if err := populateTree(child, filepath.Join(path, file.Name())); err != nil {
				return err
			}
		} else {
			child.SetReference(filepath.Join(path, file.Name()))
		}
		node.AddChild(child)
	}
	return nil
}

// createEditor creates and returns the text editor component
func createEditor() *tview.TextArea {
	return tview.NewTextArea().
		SetPlaceholder("No file loaded.")
}

// createOutput creates and returns the output view component
func createOutput() *tview.TextView {
	output := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)

	output.SetBorder(true).SetTitle("Output")

	return output
}

// createTerminal creates and returns the terminal component
func createTerminal() (*tview.TextView, error) {
	terminal := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)

	terminal.SetBorder(true).SetTitle("Terminal")

	termState.cmd = exec.Command("bash")
	var err error
	termState.pty, err = pty.Start(termState.cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}

	termState.done = make(chan struct{})
	go func() {
		defer close(termState.done)
		for {
			buf := make([]byte, 1024)
			n, err := termState.pty.Read(buf)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("Error reading from pty: %v", err)
				return
			}
			processedOutput := processANSI(buf[:n])
			ui.app.QueueUpdateDraw(func() {
				terminal.Write(processedOutput)
			})
		}
	}()

	terminal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		handleTerminalInput(event)
		return nil
	})

	return terminal, nil
}

// handleTerminalInput handles input to the terminal
func handleTerminalInput(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyRune:
		_, _ = termState.pty.Write([]byte(string(event.Rune())))
	case tcell.KeyEnter:
		_, _ = termState.pty.Write([]byte("\n"))
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		_, _ = termState.pty.Write([]byte{0x7f})
	case tcell.KeyTab:
		_, _ = termState.pty.Write([]byte{0x09})
	case tcell.KeyEscape:
		_, _ = termState.pty.Write([]byte{0x1b})
	default:
		if event.Key() >= tcell.KeyCtrlA && event.Key() <= tcell.KeyCtrlZ {
			_, _ = termState.pty.Write([]byte{byte(event.Key() - tcell.KeyCtrlA + 1)})
		}
	}
}

// processANSI processes ANSI escape sequences and returns cleaned output
func processANSI(input []byte) []byte {
	var output []byte
	inEscapeSeq := false
	for _, b := range input {
		if b == 0x1b { // ESC character
			inEscapeSeq = true
			continue
		}
		if inEscapeSeq {
			if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
				inEscapeSeq = false
			}
			continue
		}
		if b >= 32 && b != 127 { // Printable ASCII characters
			output = append(output, b)
		}
	}
	return output
}

// customizeTerminal creates and displays a form for customizing the terminal colors
func customizeTerminal() {
	bgInput := tview.NewInputField().SetLabel("Background Color")
	textInput := tview.NewInputField().SetLabel("Text Color")

	form := tview.NewForm().
		AddFormItem(bgInput).
		AddFormItem(textInput).
		AddButton("Save", func() {
			bgColor := bgInput.GetText()
			textColor := textInput.GetText()
			ui.terminal.SetBackgroundColor(tcell.GetColor(bgColor))
			ui.terminal.SetTextColor(tcell.GetColor(textColor))
			ui.app.SetRoot(ui.root, true)
			ui.app.SetFocus(ui.terminal)
		}).
		AddButton("Cancel", func() {
			ui.app.SetRoot(ui.root, true)
			ui.app.SetFocus(ui.terminal)
		})

	form.SetBorder(true).SetTitle("Customize Terminal")
	
	formFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 10, 1, true).
			AddItem(nil, 0, 1, false), 40, 1, true).
		AddItem(nil, 0, 1, false)

	ui.app.SetRoot(formFlex, true)
}

// loadFile loads the content of a file into the editor
func loadFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	ui.editor.SetText(string(content), true)
	currentFile = path
	ui.output.SetText(fmt.Sprintf("Loaded file: %s", path))
	return nil
}

// saveFile saves the content of the editor to the current file
func saveFile() error {
	if currentFile == "" {
		return fmt.Errorf("no file loaded")
	}
	content := ui.editor.GetText()
	err := os.WriteFile(currentFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	ui.output.SetText(fmt.Sprintf("File saved: %s", currentFile))
	return nil
}
