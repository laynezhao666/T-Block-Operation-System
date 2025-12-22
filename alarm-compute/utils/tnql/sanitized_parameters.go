package tnql

// SanitizedParameters is a wrapper for Parameters that does sanitization as
// parameters are accessed.
type SanitizedParameters struct {
	orig Parameters
	e    *EvaluableExpression
}

// InitSanitizedParameters InitSanitizedParameters
func InitSanitizedParameters(orig Parameters, e *EvaluableExpression) *SanitizedParameters {
	return &SanitizedParameters{
		orig: orig,
		e:    e,
	}
}

// Get TODO
func (p SanitizedParameters) Get(key string) (interface{}, error) {
	value, err := p.orig.Get(key)
	if err != nil {
		return nil, err
	}
	return castToFloat64(value), nil
}

// GetExpression return current EvaluableExpression
func (p SanitizedParameters) GetExpression() *EvaluableExpression {
	return p.e
}

func castToFloat64(value interface{}) interface{} {
	switch value.(type) {
	case uint8:
		return float64(value.(uint8))
	case uint16:
		return float64(value.(uint16))
	case uint32:
		return float64(value.(uint32))
	case uint64:
		return float64(value.(uint64))
	case int8:
		return float64(value.(int8))
	case int16:
		return float64(value.(int16))
	case int32:
		return float64(value.(int32))
	case int64:
		return float64(value.(int64))
	case int:
		return float64(value.(int))
	case float32:
		return float64(value.(float32))
	}

	return value
}
