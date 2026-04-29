package core

// transient.go — Clojure-style transient vectors for mutation-heavy loops.
//
// A TransientVector wraps an ArrayVector and allows in-place mutation.
// It is created via (transient v) and frozen back with (persistent! v).
// Only the IR uses transients internally — they are not exposed as
// user-facing Joker primitives in this implementation.
//
// This follows Clojure's transient semantics: single-owner, single-thread.

// TransientVector is a mutable wrapper around an array of Objects.
type TransientVector struct {
	arr []Object
}

func (tv *TransientVector) ToString(escape bool) string   { return "#<transient-vector>" }
func (tv *TransientVector) Equals(other interface{}) bool { return tv == other }
func (tv *TransientVector) GetInfo() *ObjectInfo          { return nil }
func (tv *TransientVector) WithInfo(*ObjectInfo) Object   { return tv }
func (tv *TransientVector) GetType() *Type                { return TYPE.ArrayVector }
func (tv *TransientVector) Hash() uint32                  { return 0 }

// Assoc mutates in place and returns self.
func (tv *TransientVector) AssocInPlace(key, val Object) *TransientVector {
	idx := key.(Int).I
	if idx >= 0 && idx < len(tv.arr) {
		tv.arr[idx] = val
	}
	return tv
}

// Nth returns the element at index.
func (tv *TransientVector) Nth(i int) Object {
	if i >= 0 && i < len(tv.arr) {
		return tv.arr[i]
	}
	return NIL
}

// Count returns the length.
func (tv *TransientVector) Count() int {
	return len(tv.arr)
}

// ToPersistent freezes back to an ArrayVector.
func (tv *TransientVector) ToPersistent() *ArrayVector {
	arr := make([]Object, len(tv.arr))
	copy(arr, tv.arr)
	return &ArrayVector{arr: arr}
}

// ToTransient creates a mutable copy from an ArrayVector.
func ToTransient(v *ArrayVector) *TransientVector {
	arr := make([]Object, len(v.arr))
	copy(arr, v.arr)
	return &TransientVector{arr: arr}
}
