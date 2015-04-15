package goshin

import (
	"bytes"
	"io/ioutil"
	"log"
	"strconv"
)

type Temp struct {
	sensor string
}

func (t *Temp) Collect(queue chan *Metric) {

	metric := NewMetric()
	metric.Service = "temp"

	temp := 0
	buf, err := ioutil.ReadFile(t.sensor)
	if err == nil {
		temp, _ = strconv.Atoi(string(bytes.TrimSpace(buf)))
		log.Println("[TEMP] ", temp)
	}
	metric.Value = temp
	queue <- metric
}

func NewTemp(src string) *Temp {
	return &Temp{
		sensor: src,
	}
}
