package scripting

import (
	core "github.com/rcarmo/go-joker/v42/core"
	coretypes "github.com/rcarmo/go-joker/v42/core/types"
	corecollections "github.com/rcarmo/go-joker/v42/core/types/collections"
)

// keywordizeJokerBridge mirrors walk/keywordize-keys for JSON-reader trees.
// Keep the reader's scalar objects unchanged, and build maps through Assoc:
// the reader's :keywords? option alone leaves wide maps as ArrayMaps, whereas
// the previous interpreted walk promotes them to HashMaps.
func keywordizeJokerBridge(value coretypes.Object) coretypes.Object {
	switch value := value.(type) {
	case core.Nil: // Nil implements Map, but is not an empty JSON object.
		return value
	case coretypes.Map:
		var result coretypes.Map = corecollections.EmptyArrayMap()
		for iter := value.Iter(); iter.HasNext(); {
			pair := iter.Next()
			key := pair.Key
			if text, ok := key.(coretypes.String); ok {
				key = coretypes.MakeKeyword(core.STRINGS.Intern, text.S)
			}
			result = result.Assoc(key, keywordizeJokerBridge(pair.Value)).(coretypes.Map)
		}
		return result
	case *corecollections.Vector:
		result := corecollections.EmptyVector()
		for seq := value.Seq(); !seq.IsEmpty(); seq = seq.Rest() {
			result = result.Conjoin(keywordizeJokerBridge(seq.First()))
		}
		return result
	default:
		return value
	}
}
