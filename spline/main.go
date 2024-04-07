package main

import (
	"fmt"
	// . "github.com/stevegt/goadapt"
)

// Spline represents a cubic spline.
type Spline struct {
	xs []float64
	ys []float64
	ks []float64 // Second derivatives
}

// NewSpline initializes and returns a new Spline.
func NewSpline(xs, ys []float64) *Spline {
	n := len(xs) - 1
	ks := make([]float64, n+1)

	// Solve tridiagonal system for second derivatives, ks.
	a := make([]float64, n+1)
	b := make([]float64, n+1)
	c := make([]float64, n+1)
	d := make([]float64, n+1)

	for i := 1; i < n; i++ {
		a[i] = (xs[i] - xs[i-1]) / 6
		b[i] = (xs[i+1] - xs[i-1]) / 3
		c[i] = (xs[i+1] - xs[i]) / 6
		d[i] = (ys[i+1]-ys[i])/(xs[i+1]-xs[i]) - (ys[i]-ys[i-1])/(xs[i]-xs[i-1])
	}

	// Forward elimination
	for i := 1; i < n; i++ {
		m := a[i+1] / b[i]
		b[i+1] -= m * c[i]
		d[i+1] -= m * d[i]
	}

	// Back substitution
	ks[n] = d[n] / b[n]
	for i := n - 1; i >= 0; i-- {
		ks[i] = (d[i] - c[i]*ks[i+1]) / b[i]
	}

	return &Spline{xs, ys, ks}
}

// Interpolate finds an interpolated value at x.
func (s *Spline) Interpolate(x float64) float64 {
	i := searchSorted(s.xs, x)
	h := s.xs[i+1] - s.xs[i]
	a := (s.xs[i+1] - x) / h
	b := (x - s.xs[i]) / h
	return a*s.ys[i] + b*s.ys[i+1] + ((a*a*a-a)*s.ks[i]+(b*b*b-b)*s.ks[i+1])*(h*h)/6
}

// searchSorted finds the index of the last element in xs that is less than or equal to x.
func searchSorted(xs []float64, x float64) int {
	n := len(xs)
	if x < xs[0] {
		return 0
	}
	for i := 1; i < n; i++ {
		if x < xs[i] {
			return i - 1
		}
	}
	return n - 2
}

func main() {
	// Example data points
	xs := []float64{0, 1, 2, 3, 4, 5}
	ys := []float64{-1, 0, 3, 7, 10, 12}

	spline := NewSpline(xs, ys)

	fmt.Println("Interpolated values:")
	for x := 0.0; x <= 5; x += 0.5 {
		y := spline.Interpolate(x)
		fmt.Printf("x=%.2f, y=%.2f\n", x, y)
	}
}
