package multipolyfit

import (
	. "github.com/stevegt/goadapt"
	
	// from numpy import linalg, zeros, ones, hstack, asarray
	// equivalents from gonum
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/optimize"



)

// BasisVector returns an array like [0, 0, ..., 1, ..., 0, 0] where the 1 is
// at the i-th position.
func BasisVector(n, i int) []int {
	x := make([]int, n)
	x[i] = 1
	return x
}

// MultiPolyFit fits a polynomial like y = a**2 + 3a - 2ab + 4b**2 - 1 with
// many covariates a, b, c, ...
//
// xs is an MxK matrix where M is the number of sample points and K is the
// number of covariates.  y is an Mx1 matrix of the sample points.  deg is the
// degree of the fitting polynomial.
//
// MultiPolyFit returns a callable function and an array of coefficients.
// The callable function takes in many x values and returns an approximate y
// value.
func MultiPolyFit(xsRaw [][]float64, yRaw []float64, deg int) (func(...float64) float64, []float64) {
    // python: y = asarray(y).squeeze()
    y := mat.NewDense(len(yRaw), 1, yRaw)
    // python: rows = y.shape[0]
    rows := len(xsRaw)

    // python: xs = asarray(xs)
    xs1 := mat.NewDense(rows, len(xsRaw[0]), xsRaw)

    // python: num_covariates = xs.shape[1]
    numCovariates := xs.RawMatrix().Cols


    // python: xs = hstack((ones((xs.shape[0], 1), dtype=xs.dtype) , xs))
    //
    // This line of code in Python is using the `hstack` function from the
    // NumPy library to horizontally stack two arrays. The function
    // `ones((xs.shape[0], 1), dtype=xs.dtype)` generates an array of ones
    // with the same number of rows as `xs` and 1 column, and the same
    // datatype as `xs`. Then, it concatenates this array with the `xs`
    // array along the column axis (i.e., horizontally).
    //
    // So essentially, it's inserting a column of ones as the first column of
    // the `xs` array. This is a common operation in machine learning and
    // statistics when working with models that include a constant term,
    // such as linear regression. By inserting a column of ones, it allows
    // the model to include an intercept term.
    //
    // We can create a column of ones in Go by creating a new matrix with
    // the same number of rows as the `xs` matrix, but only 1 column. 
    ones := mat.NewDense(rows, 1, nil)
    // 
    // the gonum/mat package does not have a hstack function, but
    // it does have a Augment() method that does the same thing.
    // The Augment() method takes two matrices and returns a new
    // matrix that is the horizontal concatenation of the two
    // matrices.  The first matrix is the receiver, and the second
    // matrix is the argument.  
    xs := mat.NewDense(rows, numCovariates+1, nil)
    xs.Augment(ones, xs1)

    // python: generators = [basis_vector(num_covariates+1, i) for i in range(num_covariates+1)]
    //
    // In the Python code, a list comprehension is being used to
    // create a list of generators. Each generator is created by
    // calling the `basis_vector` function with arguments
    // `num_covariates+1` and `i`. The range function generates `i`
    // from 0 up to but not including `num_covariates+1`. Therefore,
    // there will be `num_covariates+1` generators in the list. The
    // `basis_vector` function presumably generates a basis vector in
    // a space with `num_covariates+1` dimensions, with the `i`th
    // vector set to 1 and all other vectors set to 0.
    //
    // In Go, we can create a slice of generators by using a for
    // loop to iterate from 0 up to but not including
    // `num_covariates+1`.  For each iteration, we call the
    // `BasisVector` function with arguments `num_covariates+1` and
    // `i`.  The `BasisVector` function returns a slice of ints
    // representing a basis vector in a space with
    // `num_covariates+1` dimensions.  We append this slice to the
    // generators slice.
    generators := make([][]int, 0, numCovariates+1)
    for i := 0; i < numCovariates+1; i++ {
        generators = append(generators, BasisVector(numCovariates+1, i))
    }


    // # All combinations of degrees
    // python: powers = [sum(x) for x in itertools.combinations_with_replacement(generators, deg)]
    //
    // The python code calculates all combinations of a certain degree
    // using a collection of generators. The code uses functions from
    // `itertools`, a Python module for creating iterators for
    // efficient looping. Specifically, it uses
    // `itertools.combinations_with_replacement()`, which generates
    // all possible combinations that include repeated elements. 
    //
    // The specified degree (given by `deg`) controls the number of
    // elements in each combination. Then, for each combination, it
    // sums the elements and adds the result to `powers`.
    //
    // For example, if `generators` is `[1, 2, 3]` and `deg` is `2`,
    // the expression would result in a list that includes sums of all
    // possible combinations of two elements from `generators` (with
    // repetition), like so: `[2, 3, 4, 4, 5, 6]`.
    //
    // In Go, we can use the `combinations` function from the
    // `github.com/mxschmitt/golang-combinations` package to
    // calculate all combinations of a certain degree using a
    // collection of generators.  The `combinations` function takes a
    // slice of ints and an int representing the degree.  It returns
    // a slice of slices of ints, where each slice of ints is a
    // combination of the specified degree.  For example, if the
    // slice of ints is `[1, 2, 3]` and the degree is `2`, the
    // function would return `[[1, 1], [1, 2], [1, 3], [2, 2], [2, 3],
    // [3, 3]]`.  We can then use a for loop to iterate over the
    // combinations and sum the elements in each combination.  We
    // append the result to the powers slice.
    powers := make([]int, 0)
    for _, combination := range combinations(generators, deg) {
        sum := 0
        for _, element := range combination {
            sum += element
        }
        powers = append(powers, sum)
    }


    // # Raise data to specified degree pattern, stack in order
    // A = hstack(asarray([as_tall((xs**p).prod(1)) for p in powers]))
    //
    // is equivalent to:
    //
    // A = asarray([])
    // for p in powers:
    //     xsp = xs**p
    //     xsp_prod = xsp.prod(1) # product along the column axis
    //     xsp_prod_tall = as_tall(xsp_prod) # convert to tall matrix
    //     A = hstack(A, xsp_prod_tall)
    //
    // A = asarray([])
    // figure out the shape of A
    aRows := xs.RawMatrix().Rows
    aCols := len(powers)
    A := mat.NewDense(aRows, aCols, nil)
    for _, p := range powers {
        // xsp = xs**p
        xsp := mat.NewDense(xs.RawMatrix().Rows, xs.RawMatrix().Cols, nil)
        xsp.Pow(xs, p)
        // xsp_prod = xsp.prod(1) # product along the column axis
        xspProd := mat.NewDense(xsp.RawMatrix().Rows, 1, nil)
        xspProd.Product(xsp, mat.NewDense(xsp.RawMatrix().Cols, 1, nil))
        // xsp_prod_tall = as_tall(xsp_prod) # convert to tall matrix
        xspProdTall := xspProd.T()
        // A = hstack(A, xsp_prod_tall)
        A.Augment(A, xspProdTall)
    }



// beta = linalg.lstsq(A, y)[0]

This line of code is using least squares method to solve a linear matrix equation. Here, `linalg.lstsq(A, y)[0]` is a method from a linear algebra library (usually numpy or scipy in Python), which takes two parameters: `A` is a matrix, and `y` is usually a single column matrix (or vector). 

The function is trying to solve for `beta` in the equation `A*beta=y`. Geometrically, this is trying to fit a line (in two dimensions) or a plane (in three) or a hyperplane (in more than three) in such a way that the sum of the squares of the distances from all given points to this plane is the least possible, hence the term least squares.

The `[0]` at the end denotes that we're only interested in the first element of what `linalg.lstsq()` returns. This is because least squares methods can return additional information about the solution (like the residuals, rank, and singular values), but in this case we only want the solution vector `beta`.

	// if model_out:
	//     return mk_model(beta, powers)

	// if powers_out:
	//     return beta, powers
	// return beta
}


def multipolyfit(xs, y, deg, full=False, model_out=False, powers_out=False):
    """
    Least squares multivariate polynomial fit

    Fit a polynomial like ``y = a**2 + 3a - 2ab + 4b**2 - 1``
    with many covariates a, b, c, ...

    Parameters
    ----------

    xs : array_like, shape (M, k)
         x-coordinates of the k covariates over the M sample points
    y :  array_like, shape(M,)
         y-coordinates of the sample points.
    deg : int
         Degree o fthe fitting polynomial
    model_out : bool (defaults to True)
         If True return a callable function
         If False return an array of coefficients
    powers_out : bool (defaults to False)
         Returns the meaning of each of the coefficients in the form of an
         iterator that gives the powers over the inputs and 1
         For example if xs corresponds to the covariates a,b,c then the array
         [1, 2, 1, 0] corresponds to 1**1 * a**2 * b**1 * c**0

    See Also
    --------
        numpy.polyfit

    """
    y = asarray(y).squeeze()
    rows = y.shape[0]
    xs = asarray(xs)
    num_covariates = xs.shape[1]
    xs = hstack((ones((xs.shape[0], 1), dtype=xs.dtype) , xs))

    generators = [basis_vector(num_covariates+1, i)
                            for i in range(num_covariates+1)]

    # All combinations of degrees
    powers = [sum(x) for x in itertools.combinations_with_replacement(generators, deg)]

    # Raise data to specified degree pattern, stack in order
    A = hstack(asarray([as_tall((xs**p).prod(1)) for p in powers]))

    beta = linalg.lstsq(A, y)[0]

    if model_out:
        return mk_model(beta, powers)

    if powers_out:
        return beta, powers
    return beta

def mk_model(beta, powers):
    """ Create a callable python function out of beta/powers from multipolyfit

    This function is callable from within multipolyfit using the model_out flag
    """
    # Create a function that takes in many x values
    # and returns an approximate y value
    def model(*args):
        num_covariates = len(powers[0]) - 1
        if len(args)!=(num_covariates):
            raise ValueError("Expected %d inputs"%num_covariates)
        xs = asarray((1,) + args)
        return sum([coeff * (xs**p).prod()
                             for p, coeff in zip(powers, beta)])
    return model

def mk_sympy_function(beta, powers):
    from sympy import symbols, Add, Mul, S
    num_covariates = len(powers[0]) - 1
    xs = (S.One,) + symbols('x0:%d'%num_covariates)
    return Add(*[coeff * Mul(*[x**deg for x, deg in zip(xs, power)])
                        for power, coeff in zip(powers, beta)])
