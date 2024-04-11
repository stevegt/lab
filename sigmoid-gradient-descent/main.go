package main

import (
	"fmt"
	"math"

	. "github.com/stevegt/goadapt"
)

// Updated sigmoid function with y-offset (a) and multiplier (b)
func sigmoid(a, b, k, x0, x float64) float64 {
	return a + (b / (1 + math.Exp(-k*(x-x0))))
}

// Cost function: Sum of squared errors, updated for new sigmoid parameters
func costFunction(a, b, k, x0 float64, xs, ys []float64) float64 {
	sum := 0.0
	for i, x := range xs {
		predicted := sigmoid(a, b, k, x0, x)
		error := ys[i] - predicted
		sum += error * error
	}
	return sum / float64(len(xs))
}

// Partial derivatives of the cost function with respect to a, b, k, and x0
func partialDerivativeA(b, k, x0 float64, xs, ys []float64) float64 {
	sum := 0.0
	for i, x := range xs {
		predicted := sigmoid(0, b, k, x0, x) // a is not directly used in derivative
		error := ys[i] - (0 + predicted)     // Simplify to match the function form
		sum += error
	}
	return -2 * sum / float64(len(xs))
}

func partialDerivativeB(a, b, k, x0 float64, xs, ys []float64) float64 {
	sum := 0.0
	for i, x := range xs {
		sigmoidValue := sigmoid(a, b, k, x0, x)
		error := ys[i] - sigmoidValue
		derivative := 1 / (1 + math.Exp(-k*(x-x0)))
		sum += error * derivative
	}
	return -2 * sum / float64(len(xs))
}

func partialDerivativeK(a, b, k, x0 float64, xs, ys []float64) float64 {
	sum := 0.0
	for i, x := range xs {
		sigmoidValue := sigmoid(a, b, k, x0, x)
		error := ys[i] - sigmoidValue
		derivative := b * (math.Exp(-k * (x - x0))) / math.Pow(1+math.Exp(-k*(x-x0)), 2) * (x - x0)
		sum += error * derivative
	}
	return -2 * sum / float64(len(xs))
}

func partialDerivativeX0(a, b, k, x0 float64, xs, ys []float64) float64 {
	sum := 0.0
	for i, x := range xs {
		sigmoidValue := sigmoid(a, b, k, x0, x)
		error := ys[i] - sigmoidValue
		derivative := b * k * (math.Exp(-k * (x - x0))) / math.Pow(1+math.Exp(-k*(x-x0)), 2)
		sum += error * derivative
	}
	return -2 * sum / float64(len(xs))
}

// Gradient Descent to update a, b, k, and x0
func gradientDescent(xs, ys []float64, a, b, k, x0, learningRate float64, iterations int) (float64, float64, float64, float64) {
	for i := 0; i < iterations; i++ {
		a -= learningRate * partialDerivativeA(b, k, x0, xs, ys)
		b -= learningRate * partialDerivativeB(a, b, k, x0, xs, ys)
		k -= learningRate * partialDerivativeK(a, b, k, x0, xs, ys)
		x0 -= learningRate * partialDerivativeX0(a, b, k, x0, xs, ys)
	}
	return a, b, k, x0
}

func main() {
	// Example data points
	xs := []float64{0, 1, 2, 3, 4, 5}
	ys := []float64{-1, 0, 3, 7, 10, 12} // Example data not limited to 0-1 range

	// Initial parameters for a, b, k, and x0
	a, b, k, x0 := 0.0, 1.0, 1.0, 2.5

	// Learning rate and number of iterations for Gradient Descent
	learningRate := 0.001
	iterations := 10000

	// Perform Gradient Descent
	a, b, k, x0 = gradientDescent(xs, ys, a, b, k, x0, learningRate, iterations)

	fmt.Printf("After optimization: a = %.4f, b = %.4f, k = %.4f, x0 = %.4f\n", a, b, k, x0)

	// show example predictions
	Pl("Predictions:")
	for _, x := range xs {
		Pf("x=%.2f, y=%.2f\n", x, sigmoid(a, b, k, x0, x))
	}

}
