// this package super shortcut util functions
package u

func Pt[T any](in T) *T {
	var out T = in
	return &out
}
