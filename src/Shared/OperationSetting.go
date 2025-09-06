package Shared

type OperationSettings struct {
	Kp                     *float64
	Ki                     *float64
	Kd                     *float64
	MinCPU                 *float64
	MaxCPU                 *float64
	ControllerType         *string
	SetPointInSeconds      *float64
	AdaptionInterval       int
	AdaptionType           string
	ModelingPercentageStep float64
	PerformanceMeasurement bool
	SetpointList           []float64
	InitialValue           *float64
}
