package keytabstore

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	defaultKeytabLifetime int64 = 120
	maxKeytabLifetime     int64 = 86400 // Seconds (One day)
	minKeytabLifetime     int64 = 30    // Seconds
)

// Config ..
type Config struct {
	KeytabLifetime int64
}

type wrapper struct {
	principal string
	keytab    *Keytab
}

// KeytabStore ..
type KeytabStore struct {
	mutex          sync.RWMutex
	internal       map[string]*wrapper
	keytabLifetime int64
	closed         chan struct{}
	wg             sync.WaitGroup
	ticker         *time.Ticker
}

// NewKeytabStore ...
func NewKeytabStore(config *Config) *KeytabStore {

	keytabLifetime := defaultKeytabLifetime

	if config.KeytabLifetime > 0 {
		zap.L().Debug(fmt.Sprintf("KeytabLifetime is %d (config)", keytabLifetime))
		keytabLifetime = config.KeytabLifetime
	} else {
		zap.L().Debug(fmt.Sprintf("KeytabLifetime is %d (default)", keytabLifetime))
	}

	if keytabLifetime > maxKeytabLifetime {
		panic(fmt.Sprintf("KeytabLifetime %d greater then max of %d", keytabLifetime, maxKeytabLifetime))
	}

	if keytabLifetime < minKeytabLifetime {
		panic(fmt.Sprintf("KeytabLifetime %d smaller then min of %d", keytabLifetime, minKeytabLifetime))
	}

	keytabStore := &KeytabStore{
		internal:       make(map[string]*wrapper),
		keytabLifetime: keytabLifetime,
		closed:         make(chan struct{}),
		ticker:         time.NewTicker(time.Duration(keytabLifetime-30) * time.Second),
	}

	go func() {
		zap.L().Debug("Starting")
		for {
			select {
			case <-keytabStore.closed:
				zap.L().Debug("Shutting down")
				return
			case <-keytabStore.ticker.C:
				zap.L().Debug("Updater->Automatic: running")
				keytabStore.update()
				zap.L().Debug("Updater->Automatic: completed")
			}
		}
	}()

	return keytabStore
}

// NewKeytabStoreDefault ...
func NewKeytabStoreDefault() *KeytabStore {
	return NewKeytabStore(&Config{})
}

// AddPrincipal ...
func (t *KeytabStore) AddPrincipal(principal string) error {

	if principal == "" {
		panic("String 'principal' is empty")
	}

	if len(principal) < 3 && len(principal) > 254 {
		if len(principal) < 3 {
			return fmt.Errorf("Principal %s is to short", principal)
		}
		return fmt.Errorf("Principal %s is to long", principal)
	}

	if !principalRegex.MatchString(principal) {
		err := fmt.Errorf("Principal is invalid")
		zap.L().Error(fmt.Sprintf("AddPrincipal(%s)->[err=%s]", principal, err))
		return err
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.internal[principal] = &wrapper{
		principal: principal,
		keytab:    &Keytab{},
	}

	zap.L().Debug(fmt.Sprintf("AddPrincipal(%s)->[ok]", principal))
	return nil
}

// RemovePrincipal ...
func (t *KeytabStore) RemovePrincipal(principal string) error {

	if principal == "" {
		panic("String 'principal' is empty")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exist := t.internal[principal]; exist {
		delete(t.internal, principal)
		zap.L().Error(fmt.Sprintf("RemovePrincipal(%s)->[err=nil]", principal))
		return nil
	}

	err := fmt.Errorf("Principal not found")
	zap.L().Error(fmt.Sprintf("RemovePrincipal(%s)->[err=%s]", principal, err))
	return err
}

// // GetPrincipal ...
// func (t *KeytabStore) GetPrincipal(principal string) (*Principal, error) {

// 	if principal == "" {
// 		panic("String 'principal' is empty")
// 	}

// 	t.mutex.RLock()
// 	defer t.mutex.RUnlock()

// 	if principal, ok := t.internal[principal]; ok {
// 		return principal, nil
// 	}

// 	err := fmt.Errorf("Principal not found")
// 	zap.L().Error(fmt.Sprintf("GetPrincipal(%s)->[err=%s]", principal, err))
// 	return nil, err
// }

// GetKeytab ...
func (t *KeytabStore) GetKeytab(principal string) (*Keytab, error) {

	if principal == "" {
		panic("String 'principal' is empty")
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if wrapper, ok := t.internal[principal]; ok {
		if wrapper.keytab.Valid() {
			return wrapper.keytab, nil
		}
		// Try to get updated
		t.updateKeytab(wrapper)
		if wrapper.keytab.Valid() {
			return wrapper.keytab, nil
		}
		return nil, fmt.Errorf("Please talk to you system administrator")
	}

	zap.L().Debug(fmt.Sprintf("Principal %s not found", principal))
	return nil, fmt.Errorf("Principal not found")
}

// UpdateNow Update Keytab cache now
func (t *KeytabStore) UpdateNow() {
	zap.L().Debug("Updater->Manual: running")
	t.update()
	zap.L().Debug("Updater->Manual: completed")
}

func (t *KeytabStore) update() {

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, wrapper := range t.internal {
		if wrapper.keytab.Valid() {
			zap.L().Debug(fmt.Sprintf("Keytab for principal %s is valid", wrapper.principal))
		} else {
			zap.L().Debug(fmt.Sprintf("Keytab for principal %s is invalid; refreshing", wrapper.principal))
			t.updateKeytab(wrapper)
		}
	}

}

func (t *KeytabStore) updateKeytab(wrapper *wrapper) {

	if runtime.GOOS != "windows" {

		zap.L().Debug(fmt.Sprintf("OS is %s not windows. Keytab for principal %s is not real", runtime.GOOS, wrapper.principal))

		base64File := fmt.Sprintf("This is not a valid keytab file because the OS %s is not supported. Only Windows is supported at this time", runtime.GOOS)
		wrapper.keytab = &Keytab{
			Base64File: base64File,
			Exp:        time.Now().Unix() + t.keytabLifetime,
		}
		return
	}

	// Use the Windows ktpass utility to generate a keytab. This must be
	// executed on a Windows OS that is either a domain controller or a member
	// of the desired domain. The password will be randomly created by the
	// ktpass utility and unknown to us.
	// ktpass
	// Executable file locationL C:\Windows\System32\ktpass
	// Documentation: https://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2012-r2-and-2012/cc753771(v=ws.11)
	// [/out <FileName>]
	// [/princ <PrincipalName>]
	// [/mapuser <UserAccount>]
	// [/mapop {add|set}] [{-|+}desonly] [/in <FileName>]
	// [/pass {Password|*|{-|+}rndpass}]
	// [/minpass]
	// [/maxpass]
	// [/crypto {DES-CBC-CRC|DES-CBC-MD5|RC4-HMAC-NT|AES256-SHA1|AES128-SHA1|All}]
	// [/itercount]
	// [/ptype {KRB5_NT_PRINCIPAL|KRB5_NT_SRV_INST|KRB5_NT_SRV_HST}]
	// [/kvno <KeyVersionNum>]
	// [/answer {-|+}]
	// [/target]
	// [/rawsalt] [{-|+}dumpsalt] [{-|+}setupn] [{-|+}setpass <Password>]  [/?|/h|/help]
	//
	// Use +DumpSalt to dump MIT Salt to output

	// Our syntax
	// ktpass -out $file -mapUser $principal +rndPass -mapOp set -crypto AES256-SHA1 -ptype KRB5_NT_PRINCIPAL -princ HTTP/$principal

	zap.L().Debug(fmt.Sprintf("Creating new keytab for principal %s", wrapper.principal))

	tmpFile := tmpFile()

	exe := "C:\\Windows\\System32\\ktpass"
	args := []string{}
	args = append(args, "-out")
	args = append(args, tmpFile)
	args = append(args, "-mapUser")
	args = append(args, wrapper.principal)
	args = append(args, "+rndPass")
	args = append(args, "-mapOp")
	args = append(args, "set")
	args = append(args, "-crypto")
	args = append(args, "AES256-SHA1")
	args = append(args, "-ptype")
	args = append(args, "KRB5_NT_PRINCIPAL")
	args = append(args, "-princ")
	args = append(args, "HTTP/"+wrapper.principal)

	cmd := exec.Command(exe, args...)

	zap.L().Debug(fmt.Sprintf("exec.Command(%s, %s)", exe, args))

	err := cmd.Run()
	if err != nil {
		//TODO: Handle error better
		zap.L().Error(fmt.Sprintf("Error getting keytab: cmd=%s %s, err=%s", exe, args, err))
		wrapper.keytab = &Keytab{}
		return
	}

	f, err := os.Open(tmpFile)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Error getting keytab %s", err))
		wrapper.keytab = &Keytab{}
		return
	}

	defer f.Close()
	defer os.Remove(f.Name())

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Error getting keytab %s", err))
		wrapper.keytab = &Keytab{}
		return
	}

	wrapper.keytab = &Keytab{
		Base64File: base64.StdEncoding.EncodeToString(content),
		Exp:        time.Now().Unix() + t.keytabLifetime,
	}

	zap.L().Debug(fmt.Sprintf("Keytab for principal %s loaded into cache", wrapper.principal))
}

// Shutdown Currently does nothing. Keeping as an option in case we want to
// cleanup later.
func (t *KeytabStore) Shutdown() {
	close(t.closed)
	t.wg.Wait()
}

// Create a tmp file and then delete it. This way we know we can write to
// the temp location. Then return the full path to the file as a string.
func tmpFile() string {
	f, err := ioutil.TempFile("", "keytab")
	if err != nil {
		panic(err.Error())
	}
	f.Close()
	os.Remove(f.Name())
	return f.Name()
}
