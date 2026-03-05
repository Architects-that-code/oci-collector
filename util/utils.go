package util

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/common-nighthawk/go-figure"
	"gopkg.in/yaml.v2"
)

// PrintSpace prints an empty line.
func PrintSpace() {
	fmt.Println("")
}

// PrintBanner prints the project banner using figlet-style text.
func PrintBanner() {
	myFigure := figure.NewColorFigure("Architects That Code", "", "blue", false)
	//myFigure.Scroll(800, 100, "left")
	myFigure.Print()
	PrintSpace()
}

// ToJSON marshals data to JSON bytes.
func ToJSON[T any](data T) ([]byte, error) {
	return json.Marshal(data)
}

// ToYAML marshals data to YAML bytes.
func ToYAML[T any](data T) ([]byte, error) {

	return yaml.Marshal(data)

}

// WriteToFile writes data to a file in the user's home directory.
func WriteToFile(filename string, data []byte) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	filename = fmt.Sprintf("%s/%s", homedir, filename)
	fmt.Printf("\nWriting to file: %s\n", filename)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
