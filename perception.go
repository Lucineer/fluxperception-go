package fluxperception

type Sensor struct {
	Id         uint8
	Value      float64
	Confidence float64
	Weight     float64
	Bias       float64
	Timestamp  uint64
	Active     bool
}

type FusedSignal struct {
	Value       float64
	Confidence  float64
	Variance    float64
	SourceCount uint8
	Timestamp   uint64
}

type Engine struct {
	sensors   []Sensor
	history   []FusedSignal
	threshold float64
}

func NewEngine(threshold float64) *Engine {
	return &Engine{
		sensors:   make([]Sensor, 0),
		history:   make([]FusedSignal, 0),
		threshold: threshold,
	}
}

func (e *Engine) AddSensor(id uint8, weight float64, bias float64) {
	e.sensors = append(e.sensors, Sensor{
		Id:       id,
		Weight:   weight,
		Bias:     bias,
		Active:   true,
	})
}

func (e *Engine) FindSensor(id uint8) *Sensor {
	for i := range e.sensors {
		if e.sensors[i].Id == id {
			return &e.sensors[i]
		}
	}
	return nil
}

func (e *Engine) Update(sensorId uint8, value float64, confidence float64, now uint64) {
	s := e.FindSensor(sensorId)
	if s == nil || !s.Active {
		return
	}
	s.Value = value
	s.Confidence = confidence
	s.Timestamp = now
}

func (e *Engine) Read() FusedSignal {
	var totalWeight, weightedSum, weightedConf float64
	var count uint8
	var latest uint64

	for i := range e.sensors {
		s := &e.sensors[i]
		if !s.Active || s.Confidence < e.threshold {
			continue
		}
		w := s.Weight * s.Confidence
		adjusted := s.Value + s.Bias
		totalWeight += w
		weightedSum += adjusted * w
		weightedConf += s.Confidence * w
		count++
		if s.Timestamp > latest {
			latest = s.Timestamp
		}
	}

	if count == 0 || totalWeight == 0 {
		return FusedSignal{}
	}

	avgValue := weightedSum / totalWeight
	avgConf := weightedConf / totalWeight

	// Variance: weighted sum of squared deviations
	var variance float64
	for i := range e.sensors {
		s := &e.sensors[i]
		if !s.Active || s.Confidence < e.threshold {
			continue
		}
		w := s.Weight * s.Confidence / totalWeight
		diff := (s.Value + s.Bias) - avgValue
		variance += w * diff * diff
	}

	signal := FusedSignal{
		Value:       avgValue,
		Confidence:  avgConf,
		Variance:    variance,
		SourceCount: count,
		Timestamp:   latest,
	}
	e.history = append(e.history, signal)
	return signal
}

func (e *Engine) History(n int) []FusedSignal {
	if n <= 0 || len(e.history) == 0 {
		return nil
	}
	start := len(e.history) - n
	if start < 0 {
		start = 0
	}
	out := make([]FusedSignal, len(e.history)-start)
	copy(out, e.history[start:])
	return out
}

func (e *Engine) Agreement() float64 {
	var values []float64
	for i := range e.sensors {
		s := &e.sensors[i]
		if s.Active && s.Confidence >= e.threshold {
			values = append(values, s.Value+s.Bias)
		}
	}
	if len(values) < 2 {
		return 1.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	variance := 0.0
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(values))
	if variance == 0 {
		return 1.0
	}
	// Normalize by max possible spread heuristic: agreement = 1/(1+sqrt(variance))
	return 1.0 / (1.0 + sqrt(variance))
}

func sqrt(x float64) float64 {
	// Newton's method
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func (e *Engine) Deactivate(id uint8) {
	s := e.FindSensor(id)
	if s != nil {
		s.Active = false
	}
}

func (e *Engine) Calibrate(id uint8, bias float64) {
	s := e.FindSensor(id)
	if s != nil {
		s.Bias = bias
	}
}
