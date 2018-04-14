package core

import (
	"math"

	"github.com/go-spatial/proj4go/merror"
	"github.com/go-spatial/proj4go/support"
)

// IConvertLPToXY is for 2D LP->XY conversions
type IConvertLPToXY interface {
	IOperation
	Forward(*CoordLP) (*CoordXY, error)
	Inverse(*CoordXY) (*CoordLP, error)
}

// ConvertLPToXY is the specific kind of operation
type ConvertLPToXY struct {
	Operation
	Algorithm IConvertLPToXY
}

// NewConvertLPToXY makes a new operation, and the associated alg object
func NewConvertLPToXY(sys *System, desc *OperationDescription) (IOperation, error) {

	if !desc.IsConvertLPToXY() {
		return nil, merror.New(merror.NotYetSupported)
	}

	op := &ConvertLPToXY{}
	op.Description = desc
	op.System = sys

	f := desc.creatorFunc.(ConvertLPToXYCreatorFuncType)
	obj, err := f(sys, desc)
	if err != nil {
		return nil, err
	}
	op.Algorithm = obj

	return op, nil
}

//---------------------------------------------------------------------

// Forward is the real entry point
func (op *ConvertLPToXY) Forward(lp *CoordLP) (*CoordXY, error) {

	lp, err := op.forwardPrepare(lp)
	if err != nil {
		return nil, err
	}

	xy, err := op.Algorithm.Forward(lp)
	if err != nil {
		return nil, err
	}

	xy, err = op.forwardFinalize(xy)
	if err != nil {
		return nil, err
	}

	return xy, err
}

// Inverse is the real entry point
func (op *ConvertLPToXY) Inverse(xy *CoordXY) (*CoordLP, error) {

	xy, err := op.inversePrepare(xy)
	if err != nil {
		return nil, err
	}

	lp, err := op.Algorithm.Inverse(xy)
	if err != nil {
		return nil, err
	}

	lp, err = op.inverseFinalize(lp)
	if err != nil {
		return nil, err
	}

	return lp, err
}

// ForwardPrepare is called just before calling Forward()
func (op *ConvertLPToXY) forwardPrepare(lp *CoordLP) (*CoordLP, error) {

	sys := op.System

	if math.MaxFloat64 == lp.Lam {
		return nil, merror.New(merror.ErrCoordinateError)
	}

	/* Check validity of angular input coordinates */
	if sys.Left == IOUnitsAngular {

		/* check for latitude or longitude over-range */
		var t float64
		if lp.Phi < 0 {
			t = -lp.Phi - support.PiOverTwo
		} else {
			t = lp.Phi - support.PiOverTwo
		}
		if t > epsLat || lp.Lam > 10 || lp.Lam < -10 {
			return nil, merror.New(merror.ErrLatOrLonExceededLimit)
		}

		/* Clamp latitude to -90..90 degree range */
		if lp.Phi > support.PiOverTwo {
			lp.Phi = support.PiOverTwo
		}
		if lp.Phi < -support.PiOverTwo {
			lp.Phi = -support.PiOverTwo
		}

		/* If input latitude is geocentrical, convert to geographical */
		if sys.Geoc {
			lp = GeocentricLatitude(sys, DirectionInverse, lp)
		}

		/* Ensure longitude is in the -pi:pi range */
		if !sys.Over {
			lp.Lam = support.Adjlon(lp.Lam)
		}

		if lp.Lam == math.MaxFloat64 {
			return lp, nil
		}

		/* Distance from central meridian, taking system zero meridian into account */
		lp.Lam = (lp.Lam - sys.FromGreenwich) - sys.Lam0

		/* Ensure longitude is in the -pi:pi range */
		if !sys.Over {
			lp.Lam = support.Adjlon(lp.Lam)
		}

		return lp, nil
	}

	return lp, nil
}

// ForwardFinalize is called just after calling Forward()
func (op *ConvertLPToXY) forwardFinalize(coo *CoordXY) (*CoordXY, error) {

	sys := op.System

	switch sys.Right {

	/* Handle false eastings/northings and non-metric linear units */

	/* Classic proj.4 functions return plane coordinates in units of the semimajor axis */
	case IOUnitsClassic:
		coo.X *= sys.Ellipsoid.A
		coo.Y *= sys.Ellipsoid.A
		fallthrough

	/* Falls through */ /* (<-- GCC warning silencer) */
	/* to continue processing in common with PJ_IO_UNITS_PROJECTED */
	case IOUnitsProjected:
		coo.X = sys.FromMeter * (coo.X + sys.X0)
		coo.Y = sys.FromMeter * (coo.Y + sys.Y0)
		///////////////////coo.Z = sys.VFromMeter * (coo.Z + sys.Z0)

	}

	return coo, nil
}

// InversePrepare is called just before calling Inverse()
func (op *ConvertLPToXY) inversePrepare(coo *CoordXY) (*CoordXY, error) {

	sys := op.System

	if coo.X == math.MaxFloat64 {
		return nil, merror.New(merror.ErrInvalidXOrY)
	}

	/* Handle remaining possible input types */
	switch sys.Right {

	case IOUnitsWhatever:
		return coo, nil

		/* de-scale and de-offset */
	case IOUnitsCartesian:
		coo.X = sys.ToMeter*coo.X - sys.X0
		coo.Y = sys.ToMeter*coo.Y - sys.Y0

		return coo, nil

	case IOUnitsProjected, IOUnitsClassic:

		coo.X = sys.ToMeter*coo.X - sys.X0
		coo.Y = sys.ToMeter*coo.Y - sys.Y0
		if sys.Right == IOUnitsProjected {
			return coo, nil
		}

		/* Classic proj.4 functions expect plane coordinates in units of the semimajor axis  */
		/* Multiplying by ra, rather than dividing by a because the CalCOFI projection       */
		/* stomps on a and hence (apparently) depends on this to roundtrip correctly         */
		/* (CalCOFI avoids further scaling by stomping - but a better solution is possible)  */
		coo.X *= sys.Ellipsoid.Ra
		coo.Y *= sys.Ellipsoid.Ra
		return coo, nil
	}

	/* Should not happen, so we could return pj_coord_err here */
	return coo, nil
}

// InverseFinalize is called just after calling Inverse()
func (op *ConvertLPToXY) inverseFinalize(coo *CoordLP) (*CoordLP, error) {

	sys := op.System

	if sys.Left == IOUnitsAngular {

		if sys.Right != IOUnitsAngular {
			/* Distance from central meridian, taking system zero meridian into account */
			coo.Lam = coo.Lam + sys.FromGreenwich + sys.Lam0

			/* adjust longitude to central meridian */
			if !sys.Over {
				coo.Lam = support.Adjlon(coo.Lam)
			}

			if coo.Lam == math.MaxFloat64 {
				return coo, nil
			}
		}

		/* If input latitude was geocentrical, convert back to geocentrical */
		if sys.Geoc {
			coo = GeocentricLatitude(sys, DirectionForward, coo)
		}
	}

	return coo, nil
}