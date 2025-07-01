package templates

func dict(v ...interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(v)/2)
	for i := 0; i < len(v); i += 2 {
		m[v[i].(string)] = v[i+1]
	}
	return m
}
