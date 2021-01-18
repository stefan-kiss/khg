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
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
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

func loadSshConfig(url *url.URL) (connect string, sshConfig *ssh.ClientConfig, err error) {
	var username string

	hostPort := strings.Split(url.Host, ":")

	if len(hostPort) < 2 {
		port := ssh_config.Get(url.Host, "Port")
		connect = fmt.Sprintf("%s:%s", url.Host, port)
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
		return "", nil, fmt.Errorf("unable to load private key: %v", err)
	}
	sshConfig = &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			key,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	return connect, sshConfig, nil
}

func GetFile(url *url.URL) (contents []byte, err error) {

	var fileName string

	if strings.HasPrefix(url.Path, "/./") || strings.HasPrefix(url.Path, "/~/") {
		fileName = url.Path[3:]
	}

	connect, config, err := loadSshConfig(url)

	// connect
	conn, err := ssh.Dial("tcp", connect, config)
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
	srcFile, err := client.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// copy source file to destination file
	bytes, err := io.Copy(bytesWriter, srcFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d bytes copied\n", bytes)

	return bytesContent.Bytes(), nil
}
