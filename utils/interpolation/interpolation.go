package interpolation

import (
	"math"

	"github.com/elweday/go-subtitles/utils"
)

func Linear(start, end float64) utils.Interpolator {
    return func (t float64) float64 { return start*(1-t) + end*t }
}


func Spring(from, to float64, options utils.SpringOptions) utils.Interpolator {
    stiffness := options.Stiffness
    damping := options.Damping
    mass := options.Mass

    // Calculate critical damping coefficient
    criticalDamping := 2 * math.Sqrt(stiffness*mass)

    // Calculate angular frequency
    omega := math.Sqrt(stiffness / mass)

    // Calculate damping ratio
    zeta := damping / criticalDamping

    // Calculate natural frequency
    wn := omega / math.Sqrt(1-zeta*zeta)

    return func(t float64) float64 {
        if t < 0 {
            return from
        } else if t > 1 {
            return to
        }

        // Calculate current displacement
        A := to - from
        c1 := from
        c2 := (1 - zeta) * A
        c3 := (math.Exp(-zeta*omega*t) / math.Sqrt(1-zeta*zeta)) * A
        c4 := math.Sin(wn*t) + zeta/math.Sqrt(1-zeta*zeta)*math.Cos(wn*t)

        return c1 + c2*(1-c3) + c4*c3
    }
}