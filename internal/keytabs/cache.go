package keytabs

import (
	"fmt"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

var (
	principalRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

const (
	defaultCacheRefreshInterval int = 30
	minCacheRefreshInterval         = 15
	maxCacheRefreshInterval         = 3600

	defaultLifetime = 300
	minLifetime     = 30
	maxLifetime     = 86400
)

// Config ..
type Config struct {
	CacheRefreshInterval int      `json:"cacheRefreshInterval,omitempty" yaml:"cacheRefreshInterval,omitempty"`
	Lifetime             int      `json:"lifetime,omitempty" yaml:"lifetime,omitempty"`
	Principals           []string `json:"principals,omitempty" yaml:"principals,omitempty"`
}

// TODO: This is stale

// Cache holds valid keytabs, creates valid keytabs as necessary
// and invalidates old keytabs when required. Keytabs hold encrypted passwords
// and creating new keytabs invalidates old ones. When a keytab is requested
// that already exist we must check to see if it is still valid. If so then
// we hand out the exisitng keytab. If the keytab is expired we generate a
// new one and replace the old one and return it. We need to handle the
// situation where a keytab is expired but no one has requested a new one.
// This is done by incremeting a counter when a keytab is requested.
// Periodically this count is checked and if the keytab is no longer valid
// and the checked out counter is greter then zero then a new keytab is
// generated and the counter is set back to zero. This means that it is
// possible for a keytab to be valid for a short time after it has expired
// but before it is renewed.
//
// When a keytab is requested that is valid is is returned. It is possible
// that a keytab that is close to expiration perhaps only by a second will
// be returned. This is a difficult situation to deal with. If we isssue a
// new keytab we are breaking the contract on the old one. If the expiration
// is very close we could possibly block. At this time we are leaving it to
// the client to handle this situation. Even if the keytab has technically
// expired it should still work until the cleanup job runs or a request for
// the same keytab is made again.
//
// Keytabs are held in a wrapper struct. At start time a wrapper is created
// for each principal and added to a map. The map is written to my a single
// writer and only once hence we do not need to lock the map. Wrappers do
// need to be locked as we will read and write to them by multiple readers
// and writers
type Cache struct {
	internal map[string]*wrapper
	lifetime int64
	closed   chan struct{}
	wg       sync.WaitGroup
	ticker   *time.Ticker
	mutex    sync.RWMutex
}

type wrapper struct {
	principal string
	keytab    *Keytab
	mutex     sync.Mutex
}

func (t *wrapper) keytabClone() *Keytab {
	clone := &Keytab{}
	err := copier.Copy(&clone, &t.keytab)
	if err != nil {
		panic(err)
	}
	return clone
}

// Build Returns a new Keytab Cache
func (config *Config) Build() (*Cache, error) {

	zap.L().Debug("Starting Keytab Store")

	cacheRefreshInterval := defaultCacheRefreshInterval
	lifetime := defaultLifetime

	if config.CacheRefreshInterval > 0 {
		cacheRefreshInterval = config.CacheRefreshInterval
	}

	if config.Lifetime > 0 {
		lifetime = config.Lifetime
	}

	if cacheRefreshInterval < minCacheRefreshInterval || cacheRefreshInterval > maxCacheRefreshInterval {
		return nil, fmt.Errorf(fmt.Sprintf("%s must be greater then %d and less then %d", "CacheRefreshInterval", minCacheRefreshInterval, maxCacheRefreshInterval))
	}

	if lifetime > maxLifetime || lifetime < minLifetime {
		return nil, fmt.Errorf(fmt.Sprintf("%s must be greater then %d and less then %d", "Lifetime", minLifetime, maxLifetime))
	}

	t := &Cache{
		internal: make(map[string]*wrapper),
		lifetime: int64(lifetime),
		closed:   make(chan struct{}),
		ticker:   time.NewTicker(time.Duration(cacheRefreshInterval) * time.Second),
	}

	err := t.loadPrincipals(config.Principals)
	if err != nil {
		return nil, err
	}

	go func() {

		for {
			select {
			case <-t.closed:
				zap.L().Debug("Shutting down")
				return
			case <-t.ticker.C:
				zap.L().Debug("cleanupCache: running")
				t.cleanupCache()
				zap.L().Debug("cleanupCache: completed")
			}
		}
	}()

	return t, nil
}

// Load only once and before we start
func (t *Cache) loadPrincipals(principals []string) error {

	if principals == nil {
		zap.L().Warn("principals is nil")
		return nil
	}

	if len(principals) <= 0 {
		zap.L().Warn("principals is empty")
		return nil
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, principal := range principals {

		if principal == "" {
			zap.L().Warn("Ignoring empty principal")
			continue
		}

		if len(principal) < 3 && len(principal) > 254 {
			if len(principal) < 3 {
				return fmt.Errorf("Principal %s is to short", principal)
			}
			return fmt.Errorf("Principal %s is to long", principal)
		}

		if !principalRegex.MatchString(principal) {
			return fmt.Errorf("Principal %s is invalid", principal)
		}

		keytab, err := t.newKeytab(principal)
		if err != nil {
			zap.L().Warn(fmt.Sprintf("Error on loading principal %s; error=%s", principal, err))
			return nil
		}

		t.internal[principal] = &wrapper{
			principal: principal,
			keytab:    keytab,
		}

		zap.L().Debug(fmt.Sprintf("Loaded principal %s", principal))

	}

	return nil
}

// GetKeytab returns keytab If wrapper does not exist then principal does not exist
// If the wrapper does exist then we check if it has a valid
// keytab and if it does we return it. If it does not then we
// generate a new keytab and return it. We set the flag dirty
// to true so that we know someone has the keytab
func (t *Cache) GetKeytab(principal string) (*Keytab, error) {

	if principal == "" {
		zap.L().Debug(fmt.Sprintf("Principal is empty"))
		return nil, fmt.Errorf("Principal does not exist")
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if wrapper, ok := t.internal[principal]; ok {

		wrapper.mutex.Lock()
		defer wrapper.mutex.Unlock()

		var err error

		if wrapper.keytab == nil {
			wrapper.keytab, err = t.newKeytab(wrapper.principal)
			if err != nil {
				zap.L().Error(fmt.Sprintf("Error creating keytab for prinvipal %s; err=%s", wrapper.principal, err))
				return nil, fmt.Errorf("Unable to create keytab; please talk to your system administrator")
			}
			// Func is exported. Return clone to untrusted outsiders
			return wrapper.keytabClone(), nil
		}

		if wrapper.keytab.Exp == 0 {
			wrapper.keytab.Exp = time.Now().Unix() + t.lifetime
			zap.L().Debug(fmt.Sprintf("Principal %s changed from clean to dirty; expiration is now set", wrapper.principal))

			// Func is exported. Return clone to untrusted outsiders
			return wrapper.keytabClone(), nil
		}

		wrapper.keytab.Exp = time.Now().Unix() + t.lifetime
		zap.L().Debug(fmt.Sprintf("Principal %s is dirty; expiration incremented", wrapper.principal))

		// Func is exported. Return clone to untrusted outsiders
		return wrapper.keytabClone(), nil
	}

	return nil, fmt.Errorf("Principal not found")
}

func (t *Cache) cleanupCache() {

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, wrapper := range t.internal {
		wrapper.mutex.Lock()
		defer wrapper.mutex.Unlock()

		var err error

		if wrapper.keytab == nil {
			zap.L().Debug(fmt.Sprintf("Principal %s does not have a keytab; creating new", wrapper.principal))
			wrapper.keytab, err = t.newKeytab(wrapper.principal)
			if err != nil {
				zap.L().Error(fmt.Sprintf("Unable to create Keytab for Principal %s, err:%s ", wrapper.principal, err))
			}
			continue
		}

		if wrapper.keytab.Exp == 0 {
			zap.L().Debug(fmt.Sprintf("Principal %s is clean; nothing to do", wrapper.principal))
			continue
		}

		if wrapper.keytab.Valid() {
			zap.L().Debug(fmt.Sprintf("Principal %s is dirty but still valid; nothing to do", wrapper.principal))
			continue
		}

		zap.L().Debug(fmt.Sprintf("Principal %s is dirty and invalid; creating new", wrapper.principal))

		wrapper.keytab, err = t.newKeytab(wrapper.principal)
		if err != nil {
			zap.L().Error(fmt.Sprintf("Unable to create Keytab for Principal %s, err:%s ", wrapper.principal, err))
		}

	}

}

func (t *Cache) newKeytab(principal string) (*Keytab, error) {

	//TODO

	if runtime.GOOS == "windows" {
		base64File, err := windowsNewKeytab(principal)
		if err != nil {
			return nil, err
		}

		return &Keytab{
			Principal:  "HTTP/" + principal,
			Base64File: base64File,
		}, nil
	}

	zap.L().Warn(fmt.Sprintf("This OS is not supported. Real keytabs will NOT be generated"))

	base64File, err := unixNewKeytab(principal)
	if err != nil {
		return nil, err
	}

	return &Keytab{
		Principal:  "HTTP/" + principal,
		Base64File: base64File,
	}, nil
}

// Shutdown Cache
func (t *Cache) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}
