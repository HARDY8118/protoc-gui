package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

var options struct {
	inputDir   string
	inputFiles map[string]bool
	outputDir  string
	outputLang string
}

func fileList() *fyne.Container {
	var checkboxes []fyne.CanvasObject

	for fileName, _ := range options.inputFiles {
		cb := widget.NewCheck(fileName, func(b bool) {
			options.inputFiles[fileName] = b
		})
		cb.SetChecked(true)

		checkboxes = append(checkboxes, cb)
	}

	return container.NewVBox(checkboxes...)
}

func inputDirectorySelector() *fyne.Container {
	headLabel := widget.NewLabel("Protocol Buffer directory")
	headLabel.TextStyle = fyne.TextStyle{
		Underline: true,
		Bold:      true,
	}

	protocDirectoryPath := widget.NewLabel("(Select proto files directory)")

	fileContainer := container.NewVBox()
	errorLabel := widget.NewLabel("")

	fileSelectButton := widget.NewButton("Select input directory", func() {
		directory, err := dialog.Directory().Title("Protoc files directory").Browse()

		if err != nil && err.Error() != "Cancelled" {
			panic(err)
		}

		files, err := os.ReadDir(directory)
		if err != nil {
			errorLabel.SetText(err.Error())
		}

		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".proto") {
				options.inputFiles[f.Name()] = true
			}
		}

		protocDirectoryPath.SetText(directory)
		options.inputDir = directory

		fileContainer.Add(fileList())
	})

	return container.New(
		layout.NewVBoxLayout(),
		headLabel,
		container.New(layout.NewHBoxLayout(),
			protocDirectoryPath,
			layout.NewSpacer(),
			widget.NewSeparator(),
			fileSelectButton,
		),
		fileContainer,
		errorLabel,
	)
}

func outputDirectorySelector() *fyne.Container {
	headLabel := widget.NewLabel("Output directory")
	headLabel.TextStyle = fyne.TextStyle{
		Underline: true,
		Bold:      true,
	}

	protocDirectoryPath := widget.NewLabel("(Select output directory)")

	fileSelectButton := widget.NewButton("Select output directory", func() {
		directory, err := dialog.Directory().Title("Output files files directory").Browse()

		if err != nil && err.Error() != "Cancelled" {
			panic(err)
		}

		protocDirectoryPath.SetText(directory)
		options.outputDir = directory
	})

	languageDropDown := widget.NewSelect([]string{"C++", "C#", "Java", "JavaScript", "Objective C", "PHP", "Python", "Ruby", "Golang"}, func(s string) {
		options.outputLang = s
	})
	// languageDropDown.Selected = "Golang"

	return container.New(
		layout.NewVBoxLayout(),
		headLabel,
		container.New(layout.NewHBoxLayout(),
			protocDirectoryPath,
			layout.NewSpacer(),
			widget.NewSeparator(),
			fileSelectButton,
		),

		container.New(layout.NewHBoxLayout(),
			widget.NewLabel("Output language"),
			layout.NewSpacer(),
			languageDropDown,
		),
	)
}

func getLanguageArg() string {
	switch options.outputLang {
	case "C++":
		return "--cpp_out"
	case "C#":
		return "--csharp_out"
	case "Java":
		return "--java_out"
	case "JavaScript":
		return "--js_out"
	case "Objective C":
		return "--objc_out"
	case "PHP":
		return "--php_out"
	case "Python":
		return "--python_out"
	case "Ruby":
		return "--ruby_out"
	case "Golang":
		return "--go_out"
	default:
		return ""
	}
}

func resizeStringWidth(s string, w int) string {
	lines := []string{}

	for len(s) > w {
		subS := s[:w]
		lines = append(lines, subS)
		s = s[w:]
	}

	return strings.Join(lines, "\n")
}

func submitButton() *fyne.Container {
	errorLabel := widget.NewLabel("")

	button := widget.NewButton("Generate", func() {
		errorText := ""
		if options.inputDir == "" {
			errorText += "[ERROR] Input directory not selected\n"
		}
		if len(options.inputFiles) == 0 {
			errorText += "[ERROR] Select at least 1 input file\n"
		}
		if options.outputDir == "" {
			errorText += "[ERROR] Output directory not selected\n"
		}
		if options.outputLang == "" {
			errorText += "[ERROR] Output language not selected"
		}

		command := []string{}

		filesSelected := 0
		for file, selected := range options.inputFiles {
			if selected {
				filesSelected += 1
				// absPath := strings.ReplaceAll(options.inputDir+"/"+file, " ", "\\ ")
				absPath := options.inputDir + "/" + file
				command = append(command, absPath)
			}
		}
		if filesSelected < 1 {
			errorText += "[ERROR] Select at least 1 input file\n"
		}

		if errorText != "" {
			fmt.Println(errorText)
			errorLabel.SetText(errorText)
			return
		}

		command = append(command, "--proto_path", options.inputDir)

		command = append(command, getLanguageArg(), options.outputDir)

		fmt.Println("protoc", strings.Join(command, " "))

		cmd := exec.Command("protoc", command...)

		var out bytes.Buffer
		cmd.Stdout = &out
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()

		if err != nil {
			fmt.Println(err.Error())
			fmt.Println(stderr.String())
			fmt.Println(out.String())
		}
		errorLabel.SetText(resizeStringWidth(stderr.String(), 600) + "\n" + out.String())
	})

	return container.New(
		layout.NewVBoxLayout(),
		errorLabel,
		button,
	)
}

func main() {
	options.inputFiles = make(map[string]bool)
	app := app.New()

	w := app.NewWindow("Proto GUI")

	w.SetContent(
		container.New(
			layout.NewVBoxLayout(),
			inputDirectorySelector(),
			widget.NewSeparator(),
			outputDirectorySelector(),
			widget.NewSeparator(),
			layout.NewSpacer(),
			submitButton(),
		),
	)
	w.Resize(fyne.NewSize(700, 600))
	w.ShowAndRun()
}
