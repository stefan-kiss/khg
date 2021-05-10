// Copyright (c) 2021. Stefan Kiss
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package kubesftp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/kevinburke/ssh_config"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const DEFAULT_SSH_PORT = 22

var (
	DEFAULT_KEY_PATH = "~/.ssh/id_rsa"
)

func publicKey(path string) (ssh.AuthMethod, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := homedir.Dir()
		if err != nil {
			return nil, fmt.Errorf("key path contains \"~\" and unable to find home dir: %v", err)
		}
		path = filepath.Join(home, path[2:])
	}
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read key path: %v: %v", path, err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %v: %v", path, err)
	}
	return ssh.PublicKeys(signer), nil
}

func LoadSshConfig(url *url.URL) (host string, port string, sshConfig *ssh.ClientConfig, err error) {
	var username string
	var urlHostName string

	hostPort := strings.Split(url.Host, ":")
	if len(hostPort) < 2 {
		port = ssh_config.Get(url.Host, "Port")
	} else {
		port = hostPort[1]
	}
	urlHostName = hostPort[0]

	if host = ssh_config.Get(urlHostName, "HostName"); host != "" {
		log.Debugf("ssh config HostName: %q", host)
	} else {
		host = hostPort[0]
	}

	if url.User.Username() != "" {
		username = url.User.Username()
	} else {
		username = ssh_config.Get(url.Host, "User")
	}

	keyPath := ssh_config.Get(url.Host, "IdentityFile")
	// Hard-coded default from library. Should open issue.
	// github.com/kevinburke/ssh_config/validators.go:120
	if keyPath == "~/.ssh/identity" {
		keyPath = DEFAULT_KEY_PATH
	}

	key, err := publicKey(keyPath)
	if err != nil {
		return "", "", nil, fmt.Errorf("unable to load private key: %v", err)
	}
	sshConfig = &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			key,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	log.Debugf("host: %s:%s, config: %#v", host, port, sshConfig)
	return host, port, sshConfig, nil
}

func GetFile(url *url.URL) (contents []byte, host string, port string, err error) {

	var fileName string
	log.Debugf("url: %#v", url)

	if url.Host != "" && url.Path == "" {
		url.Path = viper.GetString("defaultsourcepath")
		log.Debugf("empty path, using default: %q", url.Path)
	}

	if url.Path == "" {
		log.Fatalf("unable to determine source path: empty string")
	}

	if strings.HasPrefix(url.Path, "/./") || strings.HasPrefix(url.Path, "/~/") {
		fileName = url.Path[3:]
	} else if strings.HasPrefix(url.Path, "./") || strings.HasPrefix(url.Path, "~/") {
		fileName = url.Path[2:]
	} else {
		fileName = url.Path
	}

	host, port, config, err := LoadSshConfig(url)

	// host
	log.Debugf("connecting to: %q", host)
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// create new SFTP client
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var bytesContent bytes.Buffer
	bytesWriter := bufio.NewWriter(&bytesContent)

	// open source file
	log.Debugf("opening file: %q", fileName)
	srcFile, err := client.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// copy source file to destination file
	bRead, err := io.Copy(bytesWriter, srcFile)
	if err != nil {
		log.Fatal(err)
	}
	bytesWriter.Flush()
	log.Infof("%d bytes copied\n", bRead)
	return bytesContent.Bytes(), host, port, nil
}
