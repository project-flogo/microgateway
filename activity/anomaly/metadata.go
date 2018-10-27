package anomaly

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Depth int `md:"depth"`
}

type Input struct {
	Payload interface{} `md:"payload"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	r.Payload = values["payload"]
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"payload": r.Payload,
	}
}

type Output struct {
	Complexity float32 `md:"complexity"`
	Count      int     `"md:"count`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	complexity, err := coerce.ToFloat32(values["complexity"])
	if err != nil {
		return err
	}
	o.Complexity = complexity
	count, err := coerce.ToInt(values["count"])
	if err != nil {
		return err
	}
	o.Count = count
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"complexity": o.Complexity,
		"count":      o.Count,
	}
}
