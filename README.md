# fluxperception-go 👁️

**Go sensor fusion for agent fleets.** Confidence-weighted blending of multiple noisy sensors into a single coherent perception. Microservice-ready, zero external deps.

```go
e := fluxperception.NewEngine(0.3) // minimum confidence threshold
e.AddSensor(1, 0.5, 0.0)
e.AddSensor(2, 0.3, 0.0)

e.Update(1, 48.5, 0.95, now)
e.Update(2, 52.0, 0.60, now)

fs := e.Read()
fmt.Printf("Fused: %.2f (conf: %.2f, agreement: %.2f)\n",
    fs.Value, fs.Confidence, e.Agreement())
```

## API

```go
// Create engine
e := fluxperception.NewEngine(0.3) // confidence threshold

// Register sensors
e.AddSensor(1, 0.5, 0.0) // id, weight, bias

// Update readings
e.Update(1, 48.5, 0.95, now)

// Fuse
fs := e.Read()         // FusedSignal{Value, Confidence, Variance}
hist := e.History(10)  // last N fused signals
agree := e.Agreement() // current sensor agreement [0-1]

// Sensor management
sensor := e.FindSensor(1)
e.Deactivate(1)
e.Calibrate(1, 0.5)   // adjust bias
```

### Fusion

1. Filter sensors below confidence threshold
2. Weight by `weight × confidence`
3. Fused = normalized weighted sum
4. Confidence = mean × agreement

## Install

```bash
go get github.com/Lucineer/fluxperception-go
```

## Fleet Context

Part of the Lucineer/Cocapn fleet. Go variant of [flux-perception](https://github.com/Lucineer/flux-perception) (Rust). Pairs with [fluxtrust-go](https://github.com/Lucineer/fluxtrust-go) for trust-weighted sensor prioritization.
