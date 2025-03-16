package util

import (
	"encoding/json"
	"fmt"

	"github.com/common-nighthawk/go-figure"
)

func PrintSpace() {
	fmt.Println("")
}

func PrintBanner() {
	myFigure := figure.NewColorFigure("Architects That Code", "", "blue", false)
	myFigure.Scroll(1400, 100, "left")
	myFigure.Print()
	PrintSpace()
}

func ToJSON[T any](data T) ([]byte, error) {
	return json.Marshal(data)
}
