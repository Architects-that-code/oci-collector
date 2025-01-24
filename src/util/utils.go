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
	myFigure := figure.NewFigure("Architects That Code", "", true)
	myFigure.Print()
	PrintSpace()
}

func ToJSON[T any](data T) ([]byte, error) {
	return json.Marshal(data)
}
