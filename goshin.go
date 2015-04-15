package goshin

import (
	"fmt"
	"time"

	"github.com/bigdatadev/goryman"
)

type Metric struct {
	Service, Description, State string
	Value                       interface{}
}

func NewMetric() *Metric {
	return &Metric{State: "ok"}
}

type Threshold struct {
	Warning, Critical float64
}

func NewThreshold() *Threshold {
	return &Threshold{}
}

type Goshin struct {
	Address      string
	EventHost    string
	Interval     int
	Tag          []string
	Ttl          float32
	Ifaces       map[string]bool
	IgnoreIfaces map[string]bool
	Thresholds   map[string]*Threshold
	Checks       map[string]bool
	TempSensor   string
}

func NewGoshin() *Goshin {
	return &Goshin{
		Thresholds: make(map[string]*Threshold),
	}
}

func (g *Goshin) Start() {
	//fmt.Print("Gare aux goriiillllleeeees!\n\n\n")

	cputime := NewCPUTime()
	memoryusage := NewMemoryUsage()
	loadaverage := NewLoadAverage()
	netstats := NewNetStats(g.Ifaces, g.IgnoreIfaces)
	temp := NewTemp(g.TempSensor)

	fmt.Printf("Goshin will report each %d seconds\n", g.Interval)

	// channel size has to be large enough
	// to allow Goshin send all metrics to Riemann
	// in g.Interval
	var collectQueue chan *Metric = make(chan *Metric, 100)

	ticker := time.NewTicker(time.Second * time.Duration(g.Interval))

	for _ = range ticker.C {
		//fmt.Println("Tick at ", t)

		// TODO find a better  way
		// to check if a collector type
		// is active
		if g.Checks["cpu"] {
			go cputime.Collect(collectQueue)
		}
		if g.Checks["memory"] {
			go memoryusage.Collect(collectQueue)
		}
		if g.Checks["load"] {
			go loadaverage.Collect(collectQueue)
		}
		if g.Checks["net"] {
			go netstats.Collect(collectQueue)
		}
		if g.Checks["temp"] {
			go temp.Collect(collectQueue)
		}

		go g.Report(collectQueue)
	}
}

func (g *Goshin) EnforceState(metric *Metric) {

	threshold, present := g.Thresholds[metric.Service]

	if present {
		value := metric.Value

		// TODO threshold checking
		// only for int and float type
		switch {
		case value.(float64) > threshold.Critical:
			metric.State = "critical"
		case value.(float64) > threshold.Warning:
			metric.State = "warning"
		default:
			metric.State = "ok"
		}
	}
}

func (g *Goshin) Report(reportQueue chan *Metric) {

	c := goryman.NewGorymanClient(g.Address)
	err := c.Connect()

	if err != nil {
		fmt.Println("Can not connect to host")
	} else {

		more := true

		for more {
			select {
			case metric := <-reportQueue:
				g.EnforceState(metric)
				err := c.SendEvent(&goryman.Event{
					Metric:      metric.Value,
					Ttl:         g.Ttl,
					Service:     metric.Service,
					Description: metric.Description,
					Tags:        g.Tag,
					Host:        g.EventHost,
					State:       metric.State})

				if err != nil {
					fmt.Println("something does wrong:", err)
				}
			default:
				more = false
			}
		}
	}

	defer c.Close()
}
