package helper

func ToRancherMap(fm *map[string]string) map[string]interface{} {
	m := make(map[string]interface{})
	if fm != nil {
		for k, v := range *fm {
			m[k] = v
		}
	}

	return m
}

func ToFaasMap(rm map[string]interface{}) *map[string]string {
	if len(rm) == 0 {
		return nil
	}

	m := make(map[string]string)
	for k, v := range rm {
		if val, ok := v.(string); ok {
			m[k] = val
		}
	}

	if len(m) > 0 {
		return &m
	}

	return nil
}
