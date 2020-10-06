/*
Copyright © 2020 Jody Scott <jody@thescottsweb.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package keytab

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"go.uber.org/zap"
)

func newKeytab(principal, password string) (string, error) {
	if runtime.GOOS == "windows" {
		return windowsNewKeytab(principal, password)
	}
	return unixNewKeytab(principal, password)
}

// Windows Kerberos Implementation (Active Directory) allows for the creation
// of principals that are mapped to a user account. Only one principal may be
// mapped to a user account at a time. Once a keytab is created it will remain
// valid until the principal is removed  or the password is changed or a new
// keytab is created. The windows utility ktpass is used to create the keytabs.
// The ktpass command is executed directly on the host. Therefore this should
// be ran on a Windows system that is a member of the target domain. It must
// also be ran with privileges to allow the creation of keytabs. Generally this
// is a Domain Admin. If running as a service it is necessary that it be
// configured to run as a domain admin or user with the privileges necessary
// to create keytabs.
//
// Information about the ktpass utility is as follows
// Exe: C:\Windows\System32\ktpass
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
//
// Notes about ktpass failure functionality
// Testing on Windows Server 2019 reveals that if the user lacks the
// privileges to create keytabs the ktpass utility does not create the
// keytab but also still exits with 0 and nothing is sent to the stdout
// This was with a service account and stderr was not checked. For this
// reason we will return an auth err if the file does not exist. This
// should be refined in the future.
//
// ktpass -mapUser bob@EXAMPLE.COM -pass ** -mapOp set -crypto AES256-SHA1 -ptype KRB5_NT_PRINCIPAL -princ HTTP/bob@EXAMPLE.COM -out keytab
//
func windowsNewKeytab(principal, password string) (string, error) {

	dir, err := ioutil.TempDir("", "kt")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	filename := dir + `\file.keytab`

	exe := "C:\\Windows\\System32\\ktpass"
	args := []string{}

	args = append(args, "-mapUser")
	args = append(args, principal)
	args = append(args, "-pass")
	args = append(args, password)
	args = append(args, "-mapOp")
	args = append(args, "set")
	args = append(args, "-crypto")
	args = append(args, "AES256-SHA1")
	args = append(args, "-ptype")
	args = append(args, "KRB5_NT_PRINCIPAL")
	args = append(args, "-princ")
	args = append(args, "HTTP/"+principal)
	args = append(args, "-kvno")
	args = append(args, "1")
	args = append(args, "-out")
	args = append(args, filename)

	logarg := exe
	for _, arg := range args {
		logarg = logarg + " " + arg
	}

	//zap.L().Debug(fmt.Sprintf("command->%s", logarg))

	cmd := exec.Command(exe, args...)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err = cmd.Run()
	if err != nil {
		zap.L().Error(fmt.Sprintf("exec.Command(%s, %s)", exe, args))
		return "", err
	}

	zap.L().Info(fmt.Sprintf("command->%s, output->%s", logarg, string(cmdOutput.Bytes())))

	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

func unixNewKeytab(principal, password string) (string, error) {
	return "this is not a valid keytab, it is fake", nil
}
