package interp

type array struct {
	// items []arrayItem // hash buckets, linearly probed
	// len  int           // number of active items in items slice
	m map[string]value
}

// type arrayItem struct {
//     key string
//     val value
// }

type arrayIter struct {
	keys []string
	i    int
}

const (
	// FNV-1 64-bit constants from hash/fnv.
	offset64 = 14695981039346656037
	prime64  = 1099511628211

	arrayDefaultSize = 32
)

func newArray() *array {
	return newArraySize(arrayDefaultSize)
}

func newArraySize(initialSize int) *array {
	if initialSize < arrayDefaultSize {
		initialSize = arrayDefaultSize
	}
	return &array{m: make(map[string]value, initialSize)}
	// return &array{
	// 	items: make([]arrayItem, initialSize),
	// }
}

func (a *array) Len() int {
	return len(a.m)
}

func (a *array) Contains(key string) bool {
	_, ok := a.m[key]
	return ok
}

func (a *array) Get(key string) value {
	return a.m[key]
}

func (a *array) GetOrCreate(key string) value {
	// Strangely, per the POSIX spec, "Any other reference to a
	// nonexistent array element [apart from "in" expressions]
	// shall automatically create it."
	v, ok := a.m[key]
	if !ok {
		a.m[key] = v
	}
	return v
}

func (a *array) Set(key string, val value) {
	a.m[key] = val
}

func (a *array) Delete(key string) {
	delete(a.m, key)
}

func (a *array) DeleteAll() {
	for k := range a.m {
		delete(a.m, k)
	}
}

func (a *array) Iterator() *arrayIter {
	keys := make([]string, 0, len(a.m))
	for k := range a.m {
		keys = append(keys, k)
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

/*
func (c *Counter) Inc(key []byte, n int) {
    // Like hash/fnv New64, Write, Sum64 -- but inlined without extra code.
    hash := uint64(offset64)
    for _, c := range key {
        hash *= prime64
        hash ^= uint64(c)
    }

    // Make 64-bit hash in range for items slice.
    index := int(hash & uint64(len(c.items)-1))

    // If current items more than half full, double length and reinsert items.
    if c.size >= len(c.items)/2 {
        newLen := len(c.items) * 2
        if newLen == 0 {
            newLen = initialLen
        }
        newC := Counter{items: make([]CounterItem, newLen)}
        for _, item := range c.items {
            if item.Key != nil {
                newC.Inc(item.Key, item.Count)
            }
        }
        c.items = newC.items
        index = int(hash & uint64(len(c.items)-1))
    }

    // Look up key, using direct match and linear probing if not found.
    for {
        if c.items[index].Key == nil {
            // Found empty slot, add new item (copying key).
            keyCopy := make([]byte, len(key))
            copy(keyCopy, key)
            c.items[index] = CounterItem{keyCopy, n}
            c.size++
            return
        }
        if bytes.Equal(c.items[index].Key, key) {
            // Found matching slot, increment existing count.
            c.items[index].Count += n
            return
        }
        // Slot already holds another key, try next slot (linear probe).
        index++
        if index >= len(c.items) {
            index = 0
        }
    }
}

// Items returns a copy of the incremented items.
func (c *Counter) Items() []CounterItem {
    var items []CounterItem
    for _, item := range c.items {
        if item.Key != nil {
            items = append(items, item)
        }
    }
    return items
}
*/
