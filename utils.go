package main

// combineListsToMap combines two lists into a map
func combineListsToMap(keys []string, valuesPtr *[]string) (map[string]*string, bool) {
	// dereference the pointer
	values := *valuesPtr
	// check if both lists have the same length
	if len(keys) != len(values) {
		return nil, false
	}

	result := make(map[string]*string)
	for i := 0; i < len(keys); i++ {
		result[keys[i]] = &values[i]
	}

	return result, true
}
