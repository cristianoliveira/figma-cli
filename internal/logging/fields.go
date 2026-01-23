package logging

// String returns a Field with a string value.
func String(key, value string) Field {
	return NewField(key, value)
}

// Int returns a Field with an int value.
func Int(key string, value int) Field {
	return NewField(key, value)
}

// Int64 returns a Field with an int64 value.
func Int64(key string, value int64) Field {
	return NewField(key, value)
}

// Float64 returns a Field with a float64 value.
func Float64(key string, value float64) Field {
	return NewField(key, value)
}

// Bool returns a Field with a bool value.
func Bool(key string, value bool) Field {
	return NewField(key, value)
}

// Err returns a Field with an error value.
func Err(err error) Field {
	return NewField("error", err)
}

// Any returns a Field with any value.
func Any(key string, value interface{}) Field {
	return NewField(key, value)
}
