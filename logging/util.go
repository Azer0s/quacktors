package logging

func toMap(values ...interface{}) map[string]interface{} {
	if (len(values) % 2) != 0 {
		panic("invalid logging parameters")
	}

	vals := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		if k, ok := values[i].(string); ok {
			vals[k] = values[i+1]
			continue
		}
		panic("expected key to be a string")
	}

	return vals
}
