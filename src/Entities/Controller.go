package Entities

import (
	"fmt"
	"main/Shared"

	//	"math"
	"os"
	//"main.go/shared"
)

type AStarController struct {
	Info Shared.ControllerParams
}

type RatioController struct {
	info Shared.ControllerParams
}

type PIDController struct {
	info Shared.ControllerParams
}

type MRACAdaptativeController struct {
	info   Shared.ControllerParams
	yNext  float64
	ymNext float64
}

func (h *MRACAdaptativeController) Initialise(p ...float64) {
	fmt.Println("Initializing MRAC Controller")
	if len(p) < 7 {
		panic("MRAC needs 7 initialization parameters")
	}
	h.info.Min = p[0]
	h.info.Max = p[1]
	//gamma = 0.01
	h.info.Gamma = p[2]
	h.info.Am = p[3]
	h.info.Bm = p[4]
	h.info.A = p[5]
	h.info.B = p[6]
	h.ymNext = 0.1 //condição inicial do sistema modelado
	h.info.Theta1 = 0.01
	h.info.Theta2 = 0.01

}

func (h *MRACAdaptativeController) SetPreviousValue(value float64) {

	h.info.PreviousOut = value
}

func (h *MRACAdaptativeController) Update(p ...float64) float64 {
	//(y, r, ym float64) (u, yNext, ymNext float64) {
	// Modelo de referência

	r := p[0]
	y := p[1]

	h.ymNext = h.info.Am*h.ymNext + h.info.Bm*r

	// Controle adaptativo
	u := h.info.Theta1*y + h.info.Theta2*r

	// Saída da planta
	//h.yNext = h.info.A*y + h.info.B*u

	// Erro e atualização (MIT Rule)
	e := y - h.ymNext
	//isSaturated := u < h.info.Min || u > h.info.Max
	isSaturated := u > h.info.Max
	if !isSaturated {
		fmt.Println("MRAC: Controller is not saturated")
		h.info.Theta1 = h.info.Theta1 - h.info.Gamma*e*y
		h.info.Theta2 = h.info.Theta2 - h.info.Gamma*e*r
	}
	fmt.Println("MRAC: Controller output before clip", u)
	u = clip(u, h.info.Min, h.info.Max)
	fmt.Println("MRAC: Controller output after clip", u)

	return u

}

func clip(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (h *RatioController) Initialise(p ...float64) {
	h.info.Min = p[0]
	h.info.Max = p[1]
	h.info.PreviousOut = h.info.Min

}

func (h *RatioController) Update(p ...float64) float64 {

	setpoint := p[0]
	currentMetricValue := p[1]

	u := h.info.PreviousOut * (currentMetricValue / setpoint) //HPA function

	if u < h.info.Min {
		u = h.info.Min
	}
	if u > h.info.Max {
		u = h.info.Max
	}

	h.info.PreviousOut = u
	//h.info.PreviousRate = currentMetricValue

	fmt.Println("Ratio controller:", "u=", u)
	return u
}

func (h *RatioController) SetPreviousValue(value float64) {

	h.info.PreviousOut = value
}
func (c *AStarController) Initialise(p ...float64) {

	if len(p) < 3 {
		fmt.Printf("Error: '%s' controller requires 6 info (direction,min,max,kp,ki,kd) \n")
		os.Exit(0)
	}

	c.Info.Min = p[0]
	c.Info.Max = p[1]
	c.Info.HysteresisBand = p[2]
	c.Info.PC = 1 // TODO
}

func (c *AStarController) Update(p ...float64) float64 {
	u := 0.0
	setpoint := p[0]
	y := p[1]

	if y < (setpoint - c.Info.HysteresisBand) { // The system is bellow the goal  TODO
		if y > c.Info.PreviousRate {
			u = c.Info.PreviousOut * 1.1
			//fmt.Printf("Accelerating+ [%.4f][%.4f][%.4f][%.4f]\n", y, setpoint, c.Info.PreviousOut, u)
		} else {
			u = c.Info.PreviousOut * 0.9
			//fmt.Printf("Reducing-- [%.4f][%.4f][%.4f][%.4f]\n", y, setpoint, c.Info.PreviousOut, u)
		}
		//} else if y > (setpoint + c.Info.HysteresisBand) { // The system is above the goal TODO
	} else if y > (setpoint + c.Info.HysteresisBand) { // The system is above the goal
		if y < c.Info.PreviousRate {
			u = c.Info.PreviousOut * 1.3
			//fmt.Printf("Reducing- [%.4f][%.4f][%.4f][%.4f]\n", y, setpoint, c.Info.PreviousOut, u)
		} else {
			u = c.Info.PreviousOut * 2.5
			//fmt.Printf("Accelerating++ [%.4f][%.4f][%.4f][%.4f]\n", y, setpoint, c.Info.PreviousOut, u)
		}
	} else { // The system is at Optimum state, no action required
		u = c.Info.PreviousOut
		//fmt.Printf("Optimum Level \n")
	}

	// final check of rnew
	if u < c.Info.Min {
		u = c.Info.Min
	}
	if u > c.Info.Max {
		u = c.Info.Max
	}

	//fmt.Printf("[Rate=%.4f -> %.4f], [PC=%.4f -> %.4f]\n", c.Info.PreviousRate, y, c.Info.PreviousOut, u)

	c.Info.PreviousOut = u
	c.Info.PreviousRate = y

	return u
}

func (c *AStarController) UpdateOld(p ...float64) float64 {
	u := 0.0
	setpoint := p[0]
	y := p[1] // measured output

	//	if y < (setpoint - c.Info.HysteresisBand) { // The system is bellow the goal  TODO
	if y < (setpoint - setpoint*0.20) { // The system is bellow the goal  TODO
		if y > c.Info.PreviousRate {
			u = c.Info.PreviousOut + 1
			//fmt.Printf("Below the goal (Accelerating) [%.4f]", c.Info.OptimumLevel-c.Info.HysteresisBand)
		} else {
			u = c.Info.PreviousOut * 2
			//fmt.Printf("Below the goal (Accelerating fast) [%.4f]", c.Info.OptimumLevel-c.Info.HysteresisBand)
		}
		//} else if y > (setpoint + c.Info.HysteresisBand) { // The system is above the goal TODO
	} else if y > (setpoint + c.Info.HysteresisBand*0.20) { // The system is above the goal
		if y < c.Info.PreviousRate {
			u = c.Info.PreviousOut - 1
			//fmt.Printf("Above the goal (Reducing) [%.4f]", c.Info.OptimumLevel+c.Info.HysteresisBand)
		} else {
			u = c.Info.PreviousOut / 2
			//fmt.Printf("Above the goal (Reducing fast) [%.4f]", c.Info.OptimumLevel+c.Info.HysteresisBand)
		}
	} else { // The system is at Optimum state, no action required
		u = c.Info.PreviousOut
		//fmt.Printf("Optimum Level ")
	}

	// final check of rnew
	if u < c.Info.Min {
		u = c.Info.Min
	}
	if u > c.Info.Max {
		u = c.Info.Max
	}

	//fmt.Printf("[Rate=%.4f -> %.4f], [PC=%.4f -> %.4f]\n", c.Info.PreviousRate, y, c.Info.PreviousOut, u)

	c.Info.PreviousOut = u
	c.Info.PreviousRate = y

	return u
}

func (c *AStarController) SetGains(p ...float64) {
}

func (c *AStarController) SetPreviousValue(value float64) {

	c.Info.PreviousOut = value
}

// PID

func (c *PIDController) Initialise(p ...float64) {
	fmt.Println("PID Controller")
	if len(p) < 6 {
		fmt.Printf("Error: '%s' PID controller requires 6 info (min,max,kp,ki,kd, deltaTime) \n")
		os.Exit(0)
	}

	//c.info.Direction = p[0] -- Memory and CPU always are positive.
	c.info.DeltaTime = p[5]
	c.info.Min = p[0]
	c.info.Max = p[1]

	c.info.Kp = p[2]
	c.info.Ki = p[3]
	c.info.Kd = p[4]

	c.info.PreviousOut = p[0]
	c.info.Integrator = 0.0
	c.info.PreviousError = 0.0
	c.info.PreviousPreviousError = 0.0
	c.info.SumPrevErrors = 0.0
	c.info.Out = 0.0
	c.info.PreviousDifferentiator = 0.0
}

func (c *PIDController) SetPreviousValue(value float64) {

	c.info.PreviousOut = value
}

func (c *PIDController) Update(p ...float64) float64 {

	setpoint := p[0] // goal
	y := p[1]        // plant output

	// errors
	err := setpoint - y

	//err := math.Sqrt(math.Pow((y - setpoint), 2))      //RMSE
	//direction := (setpoint-y) / math.Abs(y-setpoint) //if y > setpoint -> increase output
	//err=err*direction

	// Proportional
	proportional := c.info.Kp * err

	// Integrator (page 49)
	integrator := (c.info.SumPrevErrors + err) * c.info.Ki * c.info.DeltaTime

	//c.info.SumPrevErrors += err * c.info.DeltaTime
	//integrator := c.info.Ki * c.info.SumPrevErrors

	// Differentiator (page 49)
	differentiator := c.info.Kd * (err - c.info.PreviousError) / c.info.DeltaTime

	// control law
	//c.info.Out = c.info.PreviousOut + (proportional+integrator+differentiator)*direction -- before
	c.info.Out = c.info.PreviousOut + (proportional + integrator + differentiator)

	fmt.Println("PID Update Params", c.info.Out)

	if c.info.Out > c.info.Max {
		c.info.Out = c.info.Max
	} else if c.info.Out < c.info.Min {
		c.info.Out = c.info.Min
	}

	c.info.PreviousError = err
	c.info.SumPrevErrors += err
	c.info.PreviousOut = c.info.Out

	return c.info.Out
}

func (c *PIDController) SetGains(p ...float64) {
	c.info.Kp = p[0]
	c.info.Ki = p[1]
	c.info.Kd = p[2]
}
