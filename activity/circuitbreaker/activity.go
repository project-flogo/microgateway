package circuitbreaker

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
)

const (
	// CircuitBreakerModeA triggers the circuit breaker when there are contiguous errors
	CircuitBreakerModeA = "a"
	// CircuitBreakerModeB triggers the circuit breaker when there are errors over time
	CircuitBreakerModeB = "b"
	// CircuitBreakerModeC triggers the circuit breaker when there are contiguous errors over time
	CircuitBreakerModeC = "c"
	// CircuitBreakerModeD is a probabilistic smart circuit breaker
	CircuitBreakerModeD = "d"
	// CircuitBreakerFailure is a failure
	CircuitBreakerFailure = -1.0
	// CircuitBreakerUnknown is an onknown status
	CircuitBreakerUnknown = 0.0
	// CircuitBreakerSuccess is a success
	CircuitBreakerSuccess = 1.0
)

func init() {
	activity.Register(&Activity{}, New)
}

var (
	// ErrorCircuitBreakerTripped happens when the circuit breaker has tripped
	ErrorCircuitBreakerTripped = errors.New("circuit breaker tripped")
	activityMetadata           = activity.ToMetadata(&Settings{}, &Input{}, &Output{})
	Now                        = time.Now
)

func New(ctx activity.InitContext) (activity.Activity, error) {
	settings := Settings{
		Mode:      CircuitBreakerModeA,
		Threshold: 5,
		Period:    60,
		Timeout:   60,
	}
	err := metadata.MapToStruct(ctx.Settings(), &settings, true)
	if err != nil {
		return nil, err
	}

	logger := ctx.Logger()
	logger.Debugf("Setting: %b", settings)

	buffer := make([]Record, settings.Threshold)
	for i := range buffer {
		buffer[i].Weight = CircuitBreakerSuccess
	}
	act := &Activity{
		mode:      settings.Mode,
		threshold: settings.Threshold,
		period:    time.Duration(settings.Period) * time.Second,
		timeout:   time.Duration(settings.Timeout) * time.Second,
		context: Context{
			buffer: buffer,
		},
	}

	return act, nil
}

// Record is a record of a request
type Record struct {
	Weight float64
	Stamp  time.Time
}

// CircuitBreakerContext is a circuit breaker context
type Context struct {
	counter   int
	processed uint64
	timeout   time.Time
	index     int
	buffer    []Record
	tripped   bool
	sync.RWMutex
}

type Activity struct {
	mode      string        `json:"mode"`
	threshold int           `json:"threshold"`
	period    time.Duration `json:"period"`
	timeout   time.Duration `json:"timeout"`
	context   Context
}

// Metadata return the metadata for the activity
func (a *Activity) Metadata() *activity.Metadata {
	return activityMetadata
}

// Eval executes the activity
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	input := Input{}
	err = ctx.GetInputObject(&input)
	if err != nil {
		return false, err
	}

	context, now, tripped := &a.context, Now(), false
	switch input.Operation {
	case "counter":
		context.Lock()
		if context.timeout.Sub(now) > 0 {
			context.Unlock()
			break
		}
		context.counter++
		context.AddRecord(CircuitBreakerFailure, now)
		if context.tripped {
			context.Trip(now, a.timeout)
			context.Unlock()
			break
		}
		switch a.mode {
		case CircuitBreakerModeA:
			if context.counter >= a.threshold {
				context.Trip(now, a.timeout)
			}
		case CircuitBreakerModeB:
			if context.processed < uint64(a.threshold) {
				break
			}
			if now.Sub(context.buffer[context.index].Stamp) < a.period {
				context.Trip(now, a.timeout)
			}
		case CircuitBreakerModeC:
			if context.processed < uint64(a.threshold) {
				break
			}
			if context.counter >= a.threshold &&
				now.Sub(context.buffer[context.index].Stamp) < a.period {
				context.Trip(now, a.timeout)
			}
		}
		context.Unlock()
	case "reset":
		context.Lock()
		switch a.mode {
		case CircuitBreakerModeA, CircuitBreakerModeB, CircuitBreakerModeC:
			if context.timeout.Sub(now) <= 0 {
				context.counter = 0
				context.tripped = false
			}
		case CircuitBreakerModeD:
			context.AddRecord(CircuitBreakerSuccess, now)
		}
		context.Unlock()
	default:
		switch a.mode {
		case CircuitBreakerModeA, CircuitBreakerModeB, CircuitBreakerModeC:
			context.RLock()
			timeout := context.timeout
			context.RUnlock()
			if timeout.Sub(now) > 0 {
				tripped = true
			}
		case CircuitBreakerModeD:
			context.RLock()
			p := context.Probability(now)
			context.RUnlock()
			if rand.Float64()*1000 < math.Floor(p*1000) {
				context.Lock()
				context.AddRecord(CircuitBreakerUnknown, now)
				context.Unlock()
				tripped = true
			}
		}
	}

	output := Output{Tripped: tripped}
	err = ctx.SetOutputObject(&output)
	if err != nil {
		return false, err
	}

	if tripped {
		return true, ErrorCircuitBreakerTripped
	}

	return true, nil
}

// Trip trips the circuit breaker
func (c *Context) Trip(now time.Time, timeout time.Duration) {
	c.timeout = now.Add(timeout)
	c.counter = 0
	c.tripped = true
}

func (c *Context) AddRecord(weight float64, now time.Time) {
	c.processed++
	c.buffer[c.index].Weight = weight
	c.buffer[c.index].Stamp = now
	c.index = (c.index + 1) % len(c.buffer)
}

// Probability computes the probability for mode d
func (c *Context) Probability(now time.Time) float64 {
	records, factor, sum := c.buffer, 0.0, 0.0
	max := float64(now.Sub(records[c.index].Stamp))
	for _, record := range records {
		a := math.Exp(-float64(now.Sub(record.Stamp)) / max)
		factor += a
		sum += record.Weight * a
	}
	sum /= factor
	return 1 / (1 + math.Exp(8*sum))
}
