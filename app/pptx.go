package app

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	animexport "github.com/23dd/animation/export"
)

func pptxFrameTimes(o options, duration float64) ([]float64, error) {
	if duration < 0 || math.IsNaN(duration) || math.IsInf(duration, 0) {
		return nil, errors.New("PPTX timeline duration must be non-negative and finite")
	}
	switch o.pptxSelection {
	case "times":
		parts := strings.Split(o.pptxTimes, ",")
		if len(parts) > animexport.MaxPPTXSlides {
			return nil, fmt.Errorf("PPTX supports at most %d slides", animexport.MaxPPTXSlides)
		}
		times := make([]float64, len(parts))
		for i, part := range parts {
			at, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
			if err != nil || at < 0 || at > duration || math.IsNaN(at) || math.IsInf(at, 0) {
				return nil, fmt.Errorf("PPTX -times item %d must be a finite number between 0 and %g", i+1, duration)
			}
			times[i] = at
		}
		return times, nil
	case "interval":
		step := o.pptxInterval
		if step <= 0 || math.IsNaN(step) || math.IsInf(step, 0) {
			return nil, errors.New("PPTX -interval must be positive and finite")
		}
		if duration/step >= animexport.MaxPPTXSlides {
			return nil, fmt.Errorf("PPTX interval produces more than %d slides; increase -interval", animexport.MaxPPTXSlides)
		}
		times := []float64{0}
		for i := 1; float64(i)*step < duration; i++ {
			times = append(times, float64(i)*step)
		}
		if duration > 0 {
			if math.Abs(times[len(times)-1]-duration) <= 1e-12*math.Abs(duration) {
				times[len(times)-1] = duration
			} else {
				times = append(times, duration)
			}
		}
		if len(times) > animexport.MaxPPTXSlides {
			return nil, fmt.Errorf("PPTX supports at most %d slides; increase -interval", animexport.MaxPPTXSlides)
		}
		return times, nil
	default:
		if o.at < 0 || o.at > duration || math.IsNaN(o.at) || math.IsInf(o.at, 0) {
			return nil, fmt.Errorf("PPTX -time must be finite and between 0 and %g", duration)
		}
		return []float64{o.at}, nil
	}
}
