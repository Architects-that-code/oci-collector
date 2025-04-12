package util

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/common-nighthawk/go-figure"
	"gopkg.in/yaml.v2"
)

func PrintSpace() {
	fmt.Println("")
}

func PrintBanner() {
	myFigure := figure.NewColorFigure("Architects That Code", "", "blue", false)
	//myFigure.Scroll(800, 100, "left")
	myFigure.Print()
	PrintSpace()
}

func ToJSON[T any](data T) ([]byte, error) {
	return json.Marshal(data)
}

func ToYAML[T any](data T) ([]byte, error) {

	return yaml.Marshal(data)

}

func WriteToFile(filename string, data []byte) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Errorf("failed to get home directory: %w", err)
	}
	filename = fmt.Sprintf("%s/%s", homedir, filename)
	fmt.Printf("Writing to file: %s\n", filename)
	os.WriteFile(filename, data, 0777)

	return nil
}
