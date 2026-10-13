package print

import (
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type Printer struct {
	title     string
	startTime time.Time
}

func StartPrinter(title string) *Printer {
	p := &Printer{title: title, startTime: time.Now()}
	p.printSection(p.title)
	return p
}

func (p *Printer) EndPrinter() {
	endTime := time.Now()
	elapsedTime := endTime.Sub(p.startTime)
	p.printSection(p.title + "完成 耗时 time: " + elapsedTime.String())
}

func (p *Printer) Info(v ...interface{}) {
	hlog.Info(v)
}

func (p *Printer) printSection(title string) {
	const lineLength = 50
	padding := (lineLength - len(title) - 2) / 2
	if padding < 0 {
		padding = 0
	}
	leftPadding := padding
	rightPadding := lineLength - len(title) - 2 - padding
	if rightPadding < 0 {
		rightPadding = 0
	}

	hlog.Info(strings.Repeat("=", lineLength))
	hlog.Info(strings.Repeat(" ", leftPadding) + title + strings.Repeat(" ", rightPadding))
	hlog.Info(strings.Repeat("=", lineLength))
}
