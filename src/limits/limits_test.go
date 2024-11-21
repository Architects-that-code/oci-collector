package limits

import (
	utils "check-limits/util"
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {

	var Datapile []LimitsCollector

	var r = LimitsCollector{
		Region:    "myRegoin",
		Service:   "something HEre",
		Limitname: "banana",
		Avail:     500,
		Used:      22,
	}

	Datapile = append(Datapile, r)

	jsonData, _ := utils.ToJSON(Datapile)
	fmt.Printf("Datapile: %v\n", Datapile)
	utils.PrintSpace()
	fmt.Println(string(jsonData))

}
