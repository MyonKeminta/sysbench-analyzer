package lib

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
)

func PlotQPS(records []Record, imageFilePath string) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.X.Label.Text = "Time (sec)"
	p.Y.Label.Text = "QPS"
	p.Add(plotter.NewGrid())

	points := make(plotter.XYs, len(records))
	for i := range points {
		points[i].X = float64(records[i].Second)
		points[i].Y = records[i].QPS
	}

	plotutil.AddLinePoints(p, "", points)

	err = p.Save(1200, 800, imageFilePath)
	return err
}
