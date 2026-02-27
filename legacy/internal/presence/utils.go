package presence

import (
	"encoding/gob"
	"hash/fnv"
)

func generateCode(id string) (uint32, error) {
	h := fnv.New32a()
	if err := gob.
		NewEncoder(h).
		Encode(id); err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}
