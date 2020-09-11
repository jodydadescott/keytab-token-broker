package cachemap

// Entity Interface type held in cache. Entity must implement Valid() so that cache
// map knows when to eject entity. JSON() is used for logging transactions.
type Entity interface {
	Valid() bool
	JSON() string
}
