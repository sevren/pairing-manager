package cache

type Cache struct {
	MagicKeys map[string]Codes
}

type Codes struct {
	Code       string
	IP         string
	Created    int64
	Expiration int64
}

// New - Creates a new cache
func New() *Cache {
	var cache = Cache{}
	cache.MagicKeys = make(map[string]Codes)
	return &cache
}

// Get - Gets the item stored in the cache based on the key
func (c *Cache) Get(key string) Codes {
	return c.MagicKeys[key]
}

// Insert - Inserts the appropriate key and value
func (c *Cache) Insert(key string, val Codes) {
	c.MagicKeys[key] = val
}

// Delete - Deletes the key and the corresponding value
func (c *Cache) Delete(key string) {
	delete(c.MagicKeys, key)
}

// IsExpired - Given the key it will check if the current cached item is expired
func (c *Cache) IsExpired(key string, current int64) bool {
	return (current > c.MagicKeys[key].Expiration)
}

// Exists - tests for presence of a key in the map
func (c *Cache) Exists(key string) bool {
	_, ok := c.MagicKeys[key]
	return ok
}
