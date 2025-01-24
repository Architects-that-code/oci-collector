package limits

type LimitsCollector struct {
	Region    string `json:"region"`
	Service   string `json:"service"`
	Limitname string `json:"limitname"`
	Avail     int64  `json:"avail"`
	Used      int64  `json:"used"`
}
