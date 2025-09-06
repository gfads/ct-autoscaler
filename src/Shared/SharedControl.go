/*********************************************************************************
Author: Nelson S Rosa
Description: This program includes all information that makes up a controller.
			 Different controller types may use specific fields.
Date: 04/02/2023
*********************************************************************************/

package Shared

type FuzzySet struct {
	Center float64
	Width  float64
}

type ControllerParams struct {
	TypeName               string // Controller type name
	Direction              float64
	Kp                     float64  // kp constant used by PID controllers
	Ki                     float64  // ki constant used by PID controllers
	Kd                     float64  // kd constant used by PID controllers
	Min                    float64  // Minimum value of the controller output
	Max                    float64  // Maximum value of the controller output
	Integrator             float64  // Integrator component
	SumPrevErrors          float64  // Sum of previous errors -- used by some controllers
	PreviousOut            float64  // Previous output -- used by some controllers
	PreviousError          float64  // Last error -- used by some controllers
	PreviousPreviousError  float64  // Penultimate error -- used by some controllers
	PreviousDifferentiator float64  // Ante penultimate error -- used by some controllers
	DeadZone               float64  // Dead zone band used by some controllers
	HysteresisBand         float64  // Hysteresis band used by some controllers
	Out                    float64  // Controller output
	GainTrigger            float64  // Gain trigger based on two level gains
	PC                     float64  // for HPA only
	PreviousRate           float64  // for AsTAR
	Alfa                   float64  // For controllers Setpoint weighting
	Beta                   float64  // for controller with two degrees of freedom/Setpointweighting
	InputSet               FuzzySet // for Fuzzy controllers
	OutputSet              FuzzySet // for Fuzzy controllers
	DeltaTime              float64
	A, B                   float64 // For adaptative controller -- model
	Am, Bm                 float64 // Reference model
	Theta1                 float64 // For adaptative controller
	Theta2                 float64 //For adaptative controller
	Gamma                  float64 //For adaptative controller -- learning factor
	Ts                     float64 // sampling
}
