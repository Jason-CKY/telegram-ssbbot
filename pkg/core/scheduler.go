package core

import (
	"fmt"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vicanso/go-charts/v2"
)

func writeFile(buf []byte) error {
	tmpPath := "./tmp"
	err := os.MkdirAll(tmpPath, 0700)
	if err != nil {
		return err
	}

	file := filepath.Join(tmpPath, "line-chart.png")
	err = os.WriteFile(file, buf, 0600)
	if err != nil {
		return err
	}
	return nil
}

func ScheduleUpdate(bot *tgbotapi.BotAPI) {
	chartOption := charts.ChartOption{
		Width:  1000,
		Height: 400,
		SeriesList: []charts.Series{
			{
				Type: charts.ChartTypeLine,
				Data: charts.NewSeriesDataFromValues([]float64{
					2.0,
					4.9,
					3.0,
					3.2,
					2.6,
					2.7,
					3.6,
					2.2,
					2.6,
					2.0,
					3.4,
					3.3,
				}),
				Label: charts.SeriesLabel{Show: *charts.TrueFlag()},
			}},
		Title: charts.TitleOption{
			Text: "Singapore Savings Bonds Interest Rates",
		},
		Padding: charts.Box{
			Top:    20,
			Left:   20,
			Right:  20,
			Bottom: 20,
		},
		Legend: charts.NewLegendOption([]string{
			"Interest Rates",
		}, charts.PositionRight),
		XAxis: charts.NewXAxisOption([]string{
			"01/25",
			"02/25",
			"03/25",
			"04/25",
			"05/25",
			"06/25",
			"07/25",
			"08/25",
			"09/25",
			"10/25",
			"11/25",
			"12/25",
		}),
		ValueFormatter: func(f float64) string {
			return fmt.Sprintf("%.0f", f) + "%"
		},
	}
	p, err := charts.Render(chartOption)

	if err != nil {
		panic(err)
	}

	buf, err := p.Bytes()
	if err != nil {
		panic(err)
	}
	err = writeFile(buf)
	if err != nil {
		panic(err)
	}
}
