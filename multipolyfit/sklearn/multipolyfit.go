package multipolyfit

import (
	. "github.com/stevegt/goadapt"

	// from sklearn.preprocessing import PolynomialFeatures
	// from sklearn import linear_model

	"github.com/pa-m/sklearn/linear_model"
	"github.com/pa-m/sklearn/preprocessing"
)

func foo() {
	X := [][]float64{{0.44, 0.68}, {0.99, 0.23}}
	vector := []float64{109.85, 155.72}
	predict := []float64{0.49, 0.18}

	poly := preprocessing.PolynomialFeatures{Degree: 2}
	X_ := poly.FitTransform(X)
	predict_ := poly.Transform(predict)

	clf := linear_model.LinearRegression{}
	clf.Fit(X_, vector)

	Pl("predict", clf.Predict(X_))

}
