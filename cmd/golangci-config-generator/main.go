package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
)

func main() {
	disabledDict, err := readDisabledDict()
	if err != nil {
		panic(fmt.Errorf("failed to read disabledDict: %w", err))
	}

	templateText, err := readTemplateText()
	if err != nil {
		panic(fmt.Errorf("failed to read templateText: %w", err))
	}

	linterList, err := fetchLinterList()
	if err != nil {
		panic(fmt.Errorf("failed to fetch linterList: %w", err))
	}

	filteredLinterList := make([]string, 0, len(linterList))

	for _, linter := range linterList {
		if _, ok := disabledDict[linter]; !ok {
			filteredLinterList = append(filteredLinterList, linter)
		}
	}

	templateArgs := map[string]interface{}{
		"LinterList": filteredLinterList,
	}

	err = executeAndWrite(templateText, templateArgs)
	if err != nil {
		panic(fmt.Errorf("failed to execute and write: %w", err))
	}
}

const disabledBase = `golint
interfacer
maligned
scopelint
`

func readDisabledDict() (map[string]struct{}, error) {
	data, err := readFile(".golangci-disabled.txt", []byte(disabledBase))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	dict := make(map[string]struct{})
	for _, line := range strings.Split(string(data), "\n") {
		dict[strings.TrimSpace(line)] = struct{}{}
	}

	return dict, nil
}

const templateBase = `linters:
  disable-all: true
  enable:
  {{- range .LinterList }}
    - {{ . }}
  {{- end }}
`

func readTemplateText() (string, error) {
	data, err := readFile(".golangci-template.yml", []byte(templateBase))
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(data), nil
}

// readFile reads the named file and returns the contents.
// If the file does not exist, a new file will be created.
func readFile(name string, base []byte) ([]byte, error) {
	if _, err := os.Stat(name); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to stat file: %w", err)
		}

		if err = os.WriteFile(name, base, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}

		return base, nil
	}

	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func fetchLinterList() ([]string, error) {
	command := exec.Command("golangci-lint", "linters")
	stdout := new(bytes.Buffer)
	command.Stdout = stdout
	stderr := new(bytes.Buffer)
	command.Stderr = stderr

	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	}

	lineList := strings.Split(stdout.String(), "\n")

	linterList := make([]string, 0, len(lineList))

	for _, line := range lineList {
		partList := strings.Split(line, ":")
		if len(partList) == 0 {
			continue
		}

		line = partList[0]

		partList = strings.Split(line, " ")
		if len(partList) == 0 {
			continue
		}

		linter := partList[0]

		if strings.Contains(linter, "Disabled") ||
			strings.Contains(linter, "Enabled") ||
			len(linter) == 0 {
			continue
		}

		linterList = append(linterList, linter)
	}

	sort.Strings(linterList)

	return linterList, nil
}

func executeAndWrite(text string, args interface{}) error {
	tmpl, err := template.New("root").Parse(text)
	if err != nil {
		return fmt.Errorf("failed to parse text: %w", err)
	}

	buff := new(bytes.Buffer)
	if err = tmpl.Execute(buff, args); err != nil {
		return fmt.Errorf("failed to execute tmpl: %w", err)
	}

	err = os.WriteFile(".golangci.yml", buff.Bytes(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
