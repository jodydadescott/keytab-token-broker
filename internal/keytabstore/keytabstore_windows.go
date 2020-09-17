// +build windows

package keytabstore

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

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
func osSpecificNewKeytab(principal string) (string, error) {

	tmpFile := tmpFile()

	exe := "C:\\Windows\\System32\\ktpass"
	args := []string{}
	args = append(args, "-out")
	args = append(args, tmpFile)
	args = append(args, "-mapUser")
	args = append(args, principal)
	args = append(args, "+rndPass")
	args = append(args, "-mapOp")
	args = append(args, "set")
	args = append(args, "-crypto")
	args = append(args, "AES256-SHA1")
	args = append(args, "-ptype")
	args = append(args, "KRB5_NT_PRINCIPAL")
	args = append(args, "-princ")
	args = append(args, "HTTP/"+principal)

	cmd := exec.Command(exe, args...)

	err := cmd.Run()
	if err != nil {
		zap.L().Error(fmt.Sprintf("exec.Command(%s, %s)", exe, args))
		return "", err
	}

	f, err := os.Open(tmpFile)
	if err != nil {
		return "", err
	}

	defer f.Close()
	defer os.Remove(f.Name())

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(content), nil
}
