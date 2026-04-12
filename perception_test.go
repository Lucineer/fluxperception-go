package fluxperception

import "testing"

func TestNewEngine(t *testing.T) {
	e := NewEngine(0.1)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	if e.threshold != 0.1 {
		t.Fatalf("threshold = %v, want 0.1", e.threshold)
	}
}

func TestAddSensor(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	if len(e.sensors) != 1 || e.sensors[0].Id != 1 {
		t.Fatal("sensor not added")
	}
}

func TestFindSensor(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(5, 1.0, 0.0)
	s := e.FindSensor(5)
	if s == nil || s.Id != 5 {
		t.Fatal("sensor not found")
	}
	if e.FindSensor(99) != nil {
		t.Fatal("found non-existent sensor")
	}
}

func TestUpdateAndRead(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.AddSensor(2, 1.0, 0.0)
	e.Update(1, 10.0, 0.8, 100)
	e.Update(2, 20.0, 0.2, 100)
	sig := e.Read()
	// weighted avg: (10*0.8 + 20*0.2) / (0.8+0.2) = 12
	if sig.Value != 12.0 {
		t.Fatalf("value = %v, want 12", sig.Value)
	}
	if sig.SourceCount != 2 {
		t.Fatalf("count = %d, want 2", sig.SourceCount)
	}
	if sig.Timestamp != 100 {
		t.Fatalf("timestamp = %d, want 100", sig.Timestamp)
	}
}

func TestReadNoActiveSensors(t *testing.T) {
	e := NewEngine(0.0)
	sig := e.Read()
	if sig.SourceCount != 0 {
		t.Fatal("expected zero source count")
	}
}

func TestReadBelowThreshold(t *testing.T) {
	e := NewEngine(0.5)
	e.AddSensor(1, 1.0, 0.0)
	e.Update(1, 10.0, 0.1, 1) // below threshold
	sig := e.Read()
	if sig.SourceCount != 0 {
		t.Fatal("should skip low-confidence sensor")
	}
}

func TestDeactivate(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.Update(1, 42.0, 1.0, 1)
	e.Deactivate(1)
	sig := e.Read()
	if sig.SourceCount != 0 {
		t.Fatal("deactivated sensor should be skipped")
	}
}

func TestCalibrate(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.Update(1, 10.0, 1.0, 1)
	sig := e.Read()
	if sig.Value != 10.0 {
		t.Fatalf("before calibrate: %v", sig.Value)
	}
	e.Calibrate(1, 5.0)
	sig = e.Read()
	if sig.Value != 15.0 {
		t.Fatalf("after calibrate: %v, want 15", sig.Value)
	}
}

func TestUpdateInactiveIgnored(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.Deactivate(1)
	e.Update(1, 99.0, 1.0, 1)
	sig := e.Read()
	if sig.SourceCount != 0 {
		t.Fatal("should not read inactive sensor")
	}
}

func TestUpdateNonExistentIgnored(t *testing.T) {
	e := NewEngine(0.0)
	// should not panic
	e.Update(99, 1.0, 1.0, 1)
}

func TestHistory(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.Update(1, 1.0, 1.0, 1)
	e.Read()
	e.Update(1, 2.0, 1.0, 2)
	e.Read()
	e.Update(1, 3.0, 1.0, 3)
	e.Read()

	h := e.History(10)
	if len(h) != 3 {
		t.Fatalf("history len = %d, want 3", len(h))
	}
	if h[2].Value != 3.0 {
		t.Fatalf("latest = %v, want 3", h[2].Value)
	}

	h2 := e.History(2)
	if len(h2) != 2 || h2[0].Value != 2.0 {
		t.Fatal("limited history incorrect")
	}
}

func TestHistoryZeroAndNegative(t *testing.T) {
	e := NewEngine(0.0)
	if e.History(0) != nil {
		t.Fatal("expected nil")
	}
	if e.History(-1) != nil {
		t.Fatal("expected nil")
	}
}

func TestAgreementPerfect(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.AddSensor(2, 1.0, 0.0)
	e.Update(1, 10.0, 1.0, 1)
	e.Update(2, 10.0, 1.0, 1)
	if a := e.Agreement(); a != 1.0 {
		t.Fatalf("agreement = %v, want 1.0", a)
	}
}

func TestAgreementDisagree(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.AddSensor(2, 1.0, 0.0)
	e.Update(1, 0.0, 1.0, 1)
	e.Update(2, 100.0, 1.0, 1)
	a := e.Agreement()
	if a >= 0.5 {
		t.Fatalf("agreement = %v, want < 0.5", a)
	}
}

func TestAgreementSingleSensor(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.Update(1, 42.0, 1.0, 1)
	if a := e.Agreement(); a != 1.0 {
		t.Fatalf("single sensor agreement = %v, want 1.0", a)
	}
}

func TestAgreementNoActive(t *testing.T) {
	e := NewEngine(0.0)
	if a := e.Agreement(); a != 1.0 {
		t.Fatalf("no sensor agreement = %v, want 1.0", a)
	}
}

func TestVariance(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0)
	e.AddSensor(2, 1.0, 0.0)
	e.Update(1, 0.0, 1.0, 1)
	e.Update(2, 10.0, 1.0, 1)
	sig := e.Read()
	if sig.Variance <= 0 {
		t.Fatalf("variance = %v, want > 0", sig.Variance)
	}
}

func TestConfidenceBlending(t *testing.T) {
	e := NewEngine(0.0)
	e.AddSensor(1, 1.0, 0.0) // high confidence sensor
	e.AddSensor(2, 1.0, 0.0) // low confidence sensor
	e.Update(1, 100.0, 1.0, 1)
	e.Update(2, 0.0, 0.01, 1)
	sig := e.Read()
	if sig.Value < 50 {
		t.Fatalf("value = %v, should be closer to high-confidence sensor", sig.Value)
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct{ in, want float64 }{
		{0, 0}, {1, 1}, {4, 2}, {9, 3}, {2, 1.4142135623730951},
	}
	for _, tc := range tests {
		got := sqrt(tc.in)
		if abs(got-tc.want) > 1e-10 {
			t.Fatalf("sqrt(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
