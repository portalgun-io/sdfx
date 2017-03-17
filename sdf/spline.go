//-----------------------------------------------------------------------------
/*

Splines

*/
//-----------------------------------------------------------------------------

package sdf

import (
	"fmt"
	"math"
)

//-----------------------------------------------------------------------------

// Solve the tridiagonal matrix equation m.x = d, return x
// See: https://en.wikipedia.org/wiki/Tridiagonal_matrix_algorithm
func TriDiagonal(m []V3, d []float64) []float64 {
	// Sanity checks
	n := len(m)
	if len(d) != n {
		panic("bad sizes rows(m) != rows(d)")
	}
	if m[0].X != 0 || m[n-1].Z != 0 {
		panic("bad values for tridiagonal matrix")
	}
	if m[0].Y == 0 {
		panic("m[0].Y == 0")
	}
	cp := make([]float64, n) // c-prime
	x := make([]float64, n)  // d-prime -> x solution
	// elimination
	cp[0] = m[0].Z / m[0].Y
	x[0] = d[0] / m[0].Y
	for i := 1; i < n; i++ {
		denom := m[i].Y - m[i].X*cp[i-1]
		if denom == 0 {
			panic("denom == 0")
		}
		cp[i] = m[i].Z / denom
		x[i] = (d[i] - m[i].X*x[i-1]) / denom
	}
	// back substitution
	for i := n - 2; i >= 0; i-- {
		x[i] -= cp[i] * x[i+1]
	}
	return x
}

//-----------------------------------------------------------------------------
// Interpolate using cubic splines.
// interval: y(t) = a + bt + ct^2 + dt^3 for t in [0,1]
// 1st and 2nd derivatives are equal across intervals.
// 2nd derivatives == 0 at the endpoints (natural splines).
// See: http://mathworld.wolfram.com/CubicSpline.html

type CS struct {
	x0, x1, k  float64
	a, b, c, d float64
}

type CubicSpline struct {
	xmin, xmax float64
	spline     []CS
}

// NewCubicSpline returns n-1 cubic splines for n x-ordered data points.
func NewCubicSpline(data []V2) CubicSpline {
	// Build and solve the tridiagonal matrix
	n := len(data)
	m := make([]V3, n)
	d := make([]float64, n)
	for i := 1; i < n-1; i++ {
		m[i] = V3{1, 4, 1}
		d[i] = 3 * (data[i+1].Y - data[i-1].Y)
	}
	// Special case the end splines.
	// Assume the 2nd derivative at the end points is 0.
	m[0] = V3{0, 2, 1}
	d[0] = 3 * (data[1].Y - data[0].Y)
	m[n-1] = V3{1, 2, 0}
	d[n-1] = 3 * (data[n-1].Y - data[n-2].Y)
	x := TriDiagonal(m, d)
	// The solution data are the first derivatives.
	// Reformat as the cubic polynomial coefficients.
	spline := make([]CS, n-1)
	for i := 0; i < n-1; i++ {
		x0 := data[i].X
		x1 := data[i+1].X
		y0 := data[i].Y
		y1 := data[i+1].Y
		D0 := x[i]
		D1 := x[i+1]
		spline[i].x0 = x0
		spline[i].x1 = x1
		spline[i].k = 1.0 / (x1 - x0)
		spline[i].a = y0
		spline[i].b = D0
		spline[i].c = 3*(y1-y0) - 2*D0 - D1
		spline[i].d = 2*(y0-y1) + D0 + D1
	}
	return CubicSpline{data[0].X, data[n-1].X, spline}
}

//-----------------------------------------------------------------------------
// Operations on individual splines

// Return the t value for a given x value.
func (s *CS) XtoT(x float64) float64 {
	return s.k * (x - s.x0)
}

// Return the function value for a given t value.
func (s *CS) Function(t float64) float64 {
	return s.a + t*(s.b+t*(s.c+s.d*t))
}

// Return the first derivative for a given t value.
func (s *CS) FirstDerivative(t float64) float64 {
	return s.b + t*(2*s.c+3*s.d*t)
}

// Return the second derivative for a given t value.
func (s *CS) SecondDerivative(t float64) float64 {
	return 2*s.c + 6*s.d*t
}

//-----------------------------------------------------------------------------

// Return the spline used for a given value of x.
func (ss CubicSpline) Find(x float64) *CS {
	// sanity checking
	n := len(ss.spline)
	if n == 0 {
		panic("no splines")
	}
	// check x is within the range of the data points
	if x < ss.xmin || x > ss.xmax {
		panic("x is out of range")
	}
	// find the spline corresponding to the x value
	lo := 0
	hi := n
	for hi-lo > 1 {
		mid := (lo + hi) >> 1
		if ss.spline[mid].x0 < x {
			lo = mid
		} else {
			hi = mid
		}
	}
	return &ss.spline[lo]
}

// Return the function value for x on a set of cubic splines.
func (ss CubicSpline) Function(x float64) float64 {
	s := ss.Find(x)
	return s.Function(s.XtoT(x))
}

// Return the distance squared between a point and a point on the splines curve.
func (ss *CubicSpline) Dist2(x float64, p V2) float64 {
	dx := x - p.X
	dy := ss.Function(x) - p.Y
	return dx*dx + dy*dy
}

//-----------------------------------------------------------------------------

const N_SAMPLES = 1000

// Return a 2D polygon approximating the cubic spline.
func (ss *CubicSpline) Polygonize() SDF2 {
	p := NewPolygon()
	p.Add(ss.xmin, 0)
	p.Add(ss.xmax, 0)
	dx := (ss.xmax - ss.xmin) / float64(N_SAMPLES-1)
	x := ss.xmax
	for i := 0; i < N_SAMPLES; i++ {
		p.Add(x, ss.Function(x))
		x -= dx
	}
	p.Render("spline.dxf")
	return Polygon2D(p.Vertices())
}

//-----------------------------------------------------------------------------

// Dumb search for the minimum point/spline distance
func (ss *CubicSpline) Min1(p V2) float64 {
	delta := (ss.xmax - ss.xmin) / float64(N_SAMPLES)
	x := ss.xmin
	xmin := x
	dmin2 := ss.Dist2(ss.xmin, p)
	for i := 0; i < N_SAMPLES; i++ {
		d2 := ss.Dist2(x, p)
		if d2 < dmin2 {
			dmin2 = d2
			xmin = x
		}
		x += delta
	}
	fmt.Printf("min point %v\n", V2{xmin, ss.Function(xmin)})
	return math.Sqrt(dmin2)
}

//-----------------------------------------------------------------------------

// Return a new t estimate for minimum distance using the Newton Raphson method.
func (s *CS) NR_Iterate(x0, y0, x float64) float64 {
	// We are minimising the distance squared function.
	// We are looking for the zeroes of the first derivative of this function.
	// d0 := dx * dx + dy * dy // distance * distance
	// d1 := 2 * (dx + y1*dy) // first derivative
	// d2 := 2 * (1 + y1*y1 + y2*dy) // second derivative
	// xnew = x - d1 / d2

	t := s.XtoT(x)
	dy := s.Function(t) - y0
	dx := x - x0
	y1 := s.k * s.FirstDerivative(t)
	y2 := s.k * s.k * s.SecondDerivative(t)

	return x - (dx+y1*dy)/(1+y1*y1+y2*dy)
}

const NR_TOLERANCE = 0.000001
const NR_MAX_ITERATIONS = 10

// Newton Raphson search for the minimum point/spline distance
func (ss *CubicSpline) Min2(p V2) float64 {
	// initial estimate
	s := ss.Find(p.X)
	x := p.X
	for i := 0; i < NR_MAX_ITERATIONS; i++ {
		xold := x
		x = s.NR_Iterate(p.X, p.Y, x)
		if x >= s.x0 && x <= s.x1 {
			if Abs(x-xold) < NR_TOLERANCE*Abs(x) {
				// The x estimate is within tolerance
				break
			}
		} else {
			// we are out of the existing spline
			s = ss.Find(x)
		}
		fmt.Printf("min point %v\n", V2{x, ss.Function(x)})
	}
	return math.Sqrt(ss.Dist2(x, p))
}

//-----------------------------------------------------------------------------

type CubicSplineSDF2 struct {
	xmin, xmax float64
	spline     []CS
	bb         Box2 // bounding box
}

func CubicSpline2D(knot []V2) SDF2 {

	s := CubicSplineSDF2{}

	// Build and solve the tridiagonal matrix
	n := len(knot)
	m := make([]V3, n)
	d := make([]float64, n)
	for i := 1; i < n-1; i++ {
		m[i] = V3{1, 4, 1}
		d[i] = 3 * (knot[i+1].Y - knot[i-1].Y)
	}
	// Special case the end splines.
	// Assume the 2nd derivative at the end points is 0.
	m[0] = V3{0, 2, 1}
	d[0] = 3 * (knot[1].Y - knot[0].Y)
	m[n-1] = V3{1, 2, 0}
	d[n-1] = 3 * (knot[n-1].Y - knot[n-2].Y)
	x := TriDiagonal(m, d)

	// The solution data are the first derivatives.
	// Reformat as the cubic polynomial coefficients.
	s.spline = make([]CS, n-1)
	for i := 0; i < n-1; i++ {
		x0 := knot[i].X
		x1 := knot[i+1].X
		y0 := knot[i].Y
		y1 := knot[i+1].Y
		D0 := x[i]
		D1 := x[i+1]
		s.spline[i].x0 = x0
		s.spline[i].x1 = x1
		s.spline[i].k = 1.0 / (x1 - x0)
		s.spline[i].a = y0
		s.spline[i].b = D0
		s.spline[i].c = 3*(y1-y0) - 2*D0 - D1
		s.spline[i].d = 2*(y0-y1) + D0 + D1
	}

	// set x min/max
	s.xmin = knot[0].X
	s.xmax = knot[n-1].X

	// work out the bounding box
	ymax := 0.0
	for i := 0; i < n-1; i++ {
	}
	s.bb = Box2{V2{s.xmin, 0}, V2{s.xmax, ymax}}

	return &s
}

func (s *CubicSplineSDF2) Evaluate(p V2) float64 {
	return 0
}

func (s *CubicSplineSDF2) BoundingBox() Box2 {
	return s.bb
}

//-----------------------------------------------------------------------------
