// NOTE: very hacky right now, just a quick speed test

package interp

type array struct {
	items []arrayItem // hash buckets, linearly probed
	len  int           // number of active items in items slice
}

type arrayItem struct {
    key string
    val value
}

type arrayIter struct {
	keys []string
	i    int
}

const (
	// FNV-1 64-bit constants from hash/fnv.
	offset64 = 14695981039346656037
	prime64  = 1099511628211

	arrayDefaultSize = 64
)

func newArray() *array {
	return newArraySize(arrayDefaultSize)
}

func newArraySize(initialSize int) *array {
	if initialSize < arrayDefaultSize {
		initialSize = arrayDefaultSize
	}
	return &array{
		items: make([]arrayItem, initialSize), // TODO: should this be 2*initialSize
	}
}

func (a *array) Len() int {
	return a.len
}

func (a *array) Contains(key string) bool {
	v := a.Get(key)
	return v != null() // TODO: hmmm, probably need a real "contains"
}

func (a *array) Get(key string) value {
    // Like hash/fnv New64, Write, Sum64 -- but inlined without extra code.
    hash := uint64(offset64)
    for i := 0; i < len(key); i++ {
        hash *= prime64
        hash ^= uint64(key[i])
    }

    // Make 64-bit hash in range for items slice.
    index := int(hash & uint64(len(a.items)-1))

    for {
    	item := a.items[index]
        if item.key == key {
			// Found matching slot, return value.
			return item.val
        }
        if item.val == null() {
            // Found empty slot, key not present.
            return null()
        }
        // Slot already holds another key, try next slot (linear probe).
        index++
        if index >= len(a.items) {
            index = 0
        }
    }
}

func (a *array) GetOrCreate(key string) value {
	// Strangely, per the POSIX spec, "Any other reference to a
	// nonexistent array element [apart from "in" expressions]
	// shall automatically create it."
	v := a.Get(key)
	if v == null() {
		a.Set(key, v)
	}
	return v
}

func (a *array) Set(key string, val value) {
    // Like hash/fnv New64, Write, Sum64 -- but inlined without extra code.
    hash := uint64(offset64)
    for i := 0; i < len(key); i++ {
        hash *= prime64
        hash ^= uint64(key[i])
    }

    // Make 64-bit hash in range for items slice.
    index := int(hash & uint64(len(a.items)-1))

    // If current items more than half full, double length and reinsert items.
    if a.len >= len(a.items)/2 {
        newSize := len(a.items) * 2
        if newSize == 0 {
            newSize = arrayDefaultSize
        }
        newA := array{items: make([]arrayItem, newSize)}
        for _, item := range a.items {
            if item.val != null() {
                newA.Set(item.key, item.val)
            }
        }
        a.items = newA.items
        index = int(hash & uint64(len(a.items)-1))
    }

    // Look up key, using direct match and linear probing if not found.
    for {
    	item := a.items[index]
        if item.val == null() {
            // Found empty slot, add new item.
            a.items[index] = arrayItem{key: key, val: val}
            a.len++
			// fmt.Println("inserted", a.items)
            return
        }
        if item.key == key {
        	// Found this key, update item
            a.items[index] = arrayItem{key: key, val: val}
			// fmt.Println("updated", a.items)
            return
        }
        // Slot already holds a key, try next slot (linear probe).
        index++
        if index >= len(a.items) {
            index = 0
        }
    }
}

func (a *array) Delete(key string) {
	// TODO
}

func (a *array) DeleteAll() {
	a.items = make([]arrayItem, arrayDefaultSize)
	a.len = 0
}

func (a *array) Iterator() *arrayIter {
	// TODO: avoid keys slice, point directly into items
	keys := make([]string, 0, a.len)
	for _, item := range a.items {
		if item.val != null() {
			keys = append(keys, item.key)
		}
	}
	return &arrayIter{keys: keys, i: -1}
}

func (it *arrayIter) Next() bool {
	it.i++
	return it.i < len(it.keys)
}

func (it *arrayIter) Key() string {
	return it.keys[it.i]
}
