package internal

func SlicesConcat[T any](ins ...[]T) []T {
	result := []T{}
	for _, in := range ins {
		result = append(result, in...)
	}
	return result
}

func SlicesMap[I any, O any](in []I, fn func(elem I) O) []O {
	result := []O{}
	for _, elem := range in {
		result = append(result, fn(elem))
	}
	return result
}

func SlicesFlatMap[I any, O any](in []I, fn func(elem I) []O) []O {
	result := []O{}
	for _, elem := range in {
		result = append(result, fn(elem)...)
	}
	return result
}

func SlicesCompare[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ValuePointer[T any](value T) *T {
	return &value
}

func ValuePointerCompare[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	} else if a != nil && b != nil {
		return *a == *b
	} else {
		return false
	}
}

// func CompareHcloudFirewallRules(a, b []hcloud.FirewallRule) bool {
// 	if len(a) != len(b) {
// 		return false
// 	}

// 	for i := range a {
// 		ar := a[i]
// 		br := b[i]

// 		fmt.Printf("%+v\n", ar)
// 		fmt.Printf("%+v\n", br)

// 		if ar.Direction != br.Direction {
// 			return false
// 		}
// 		if !SlicesCompare(SlicesMap(ar.SourceIPs, func(x net.IPNet) string { return x.String() }), SlicesMap(br.SourceIPs, func(x net.IPNet) string { return x.String() })) {
// 			return false
// 		}
// 		if !SlicesCompare(SlicesMap(ar.DestinationIPs, func(x net.IPNet) string { return x.String() }), SlicesMap(br.DestinationIPs, func(x net.IPNet) string { return x.String() })) {
// 			return false
// 		}
// 		if ar.Protocol != br.Protocol {
// 			return false
// 		}
// 		if !ValuePointerCompare(ar.Port, br.Port) {
// 			return false
// 		}
// 		if !ValuePointerCompare(ar.Description, br.Description) {
// 			return false
// 		}
// 	}

// 	return true
// }
