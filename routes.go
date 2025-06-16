package main

var RouteSearchOptions = map[string]string{
	"North":           "in ('A-NCW','B-NAC')",
	"South":           "in ('C-SCW','D-SAC')",
	"500":             "in ('E-5CW','F-5AC')",
	"North clockwise": "= 'A-NCW'",
	"North anticlock": "= 'B-NAC'",
	"South clockwise": "= 'C-SCW'",
	"South anticlock": "= 'D-SAC'",
}

// RouteClass takes a string representing a route from
// the CSV downloaded from Wufoo and returns the
// associated route code used throughtout Alys.
func RouteClass(rc string) string {

	RC := map[string]string{
		"A": "A-NCW",
		"B": "B-NAC",
		"C": "C-SCW",
		"D": "D-SAC",
		"E": "E-5CW",
		"F": "F-5AC",
	}
	rca := rc[0:1]
	val, ok := RC[rca]
	if !ok {
		return RC["A"]
	}
	return val
}

// DisplayRoute returns only the three character code
// used by us humans to identify the route
func DisplayRoute(rc string) string {

	if len(rc) < 3 {
		return rc
	}
	return rc[2:]
}
