// Copyright (c) 2020. Stefan Kiss
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

package kubeconfig

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/stefan-kiss/khg/internal/cfg"
	"github.com/stefan-kiss/khg/internal/kubesftp"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	SshProtocol  = "ssh://"
	FileProtocol = "file://"
	LocalHost    = "127.0.0.1"
)

type KubeConfig struct {
	Url    *url.URL
	Bytes  []byte
	Config clientcmdapi.Config
	Label  string
	SrcDef cfg.Source
}

func (k *KubeConfig) ReadConfig() (err error) {

	var bContent []byte
	var host string
	if k.Url.Scheme == "ssh" {
		log.Debugf("protocol: SSH, HOST: %q", k.Url.Host)
		bContent, host, _, err = kubesftp.GetFile(k.Url)
		k.SrcDef.OverrideIp = host
		if err != nil {
			return err
		}
	} else {
		var fileName string
		k.SrcDef.OverrideIp = LocalHost
		if strings.HasPrefix(k.Url.Path, "~/") {
			home, err := homedir.Dir()
			if err != nil {
				return fmt.Errorf("unable to determine home for filename: %v :%v", k.Url.Path, err)
			}
			fileName = filepath.Join(home, k.Url.Path[2:])
		} else {
			fileName = k.Url.Path
		}
		log.Debugf("protocol: FILE, PATH: %q", fileName)
		bContent, err = ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
	}

	clientConfig, err := clientcmd.NewClientConfigFromBytes(bContent)
	if err != nil {
		return err
	}
	k.Config, err = clientConfig.RawConfig()
	if err != nil {
		return err
	}
	return nil
}

func DestInit(source string) (konf *KubeConfig, err error) {
	konf = new(KubeConfig)
	konf.Url, err = url.Parse(source)
	if err != nil {
		return nil, err
	}

	err = konf.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to read current destination config file: %v: %v", source, err)
	}
	return konf, nil
}

func SourceInit(source cfg.Source, label string) (konf *KubeConfig, err error) {
	konf = new(KubeConfig)
	if !strings.HasPrefix(source.Source, FileProtocol) && !strings.HasPrefix(source.Source, SshProtocol) {
		source.Source = SshProtocol + source.Source
	}
	konf.Url, err = url.Parse(source.Source)
	if err != nil {
		return nil, err
	}
	konf.SrcDef = source
	konf.Label = label
	log.Infof("using source: %s", source.Source)
	err = konf.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to read current destination config file: %v: %v", source, err)
	}

	return konf, nil
}

func (k *KubeConfig) WriteConfig() (err error) {

	bContent, err := k.ToYaml()
	if err != nil {
		return fmt.Errorf("WriteConfig unexpected error: %v", err)
	}

	var fileName string
	if strings.HasPrefix(k.Url.Path, "~/") {
		home, err := homedir.Dir()
		if err != nil {
			return fmt.Errorf("unable to determine home for filename: %v :%v", k.Url.Path, err)
		}
		fileName = filepath.Join(home, k.Url.Path[2:])
	} else {
		fileName = k.Url.Path
	}

	backupFileName := k.Url.Path + fmt.Sprintf(".%d", time.Now().Unix())
	_ = os.Rename(fileName, backupFileName)
	err = ioutil.WriteFile(fileName, bContent, 0600)
	if err != nil {
		return err
	}
	return nil

}

func (k *KubeConfig) ToYaml() ([]byte, error) {

	json, err := runtime.Encode(clientcmdlatest.Codec, k.Config.DeepCopyObject())
	if err != nil {
		return nil, fmt.Errorf("ToYaml unexpected error: %v", err)
	}

	bContent, err := yaml.JSONToYAML(json)
	if err != nil {
		return nil, fmt.Errorf("ToYaml unexpected error: %v", err)
	}
	return bContent, err
}

func (k *KubeConfig) CopyCurrentContext(from *KubeConfig) error {
	kubeContextName := from.Config.CurrentContext
	if kubeContextName == "" {
		return fmt.Errorf("unable to find current context : %v", from.Config)
	}

	if _, ok := from.Config.Contexts[kubeContextName]; !ok {
		return fmt.Errorf("unable to find context details: %v", from.Config)
	}
	kubeContext := from.Config.Contexts[kubeContextName]

	if kubeContext.Cluster == "" {
		return fmt.Errorf("cluster name empty: %v", from.Config)
	}

	if kubeContext.AuthInfo == "" {
		return fmt.Errorf("auth empty: %v", from.Config)
	}

	if _, ok := from.Config.Clusters[kubeContext.Cluster]; !ok {
		return fmt.Errorf("unable to find current cluster: %v", from.Config)
	}

	if _, ok := from.Config.AuthInfos[kubeContext.AuthInfo]; !ok {
		return fmt.Errorf("unable to find auth: %v", from.Config)
	}

	translatedContext := fmt.Sprintf("%s@%s", kubeContextName, from.Label)
	translatedCluster := fmt.Sprintf("%s@%s", kubeContext.Cluster, from.Label)
	translatedAuth := fmt.Sprintf("%s@%s", kubeContext.AuthInfo, from.Label)

	k.Config.Clusters[translatedCluster] = from.Config.Clusters[kubeContext.Cluster]
	k.Config.AuthInfos[translatedAuth] = from.Config.AuthInfos[kubeContext.AuthInfo]
	k.Config.Contexts[translatedContext] = from.Config.Contexts[kubeContextName]

	k.Config.Contexts[translatedContext].Cluster = translatedCluster
	k.Config.Contexts[translatedContext].AuthInfo = translatedAuth

	if !from.SrcDef.AutodetectApi && from.SrcDef.ApiAddress != "" {
		k.Config.Clusters[translatedCluster].Server = from.SrcDef.ApiAddress
	}

	// if we need to autodetect we will use the same ip used to connect by ssh
	// or 127.0.0.1 for localhost
	// and the same port as the one from the api string. We cant autodetect a different port.
	if from.SrcDef.AutodetectApi {
		apiUrl, err := url.Parse(k.Config.Clusters[translatedCluster].Server)
		if err != nil {
			return fmt.Errorf("unable to parse existing API url. autodetecting api failed: %v", err)
		}
		hostStrings := strings.Split(apiUrl.Host, ":")
		if len(hostStrings) < 2 {
			return fmt.Errorf("unable to parse existing API url. unable to find current port")
		}
		var kPort string
		if from.SrcDef.OverridePort != "" {
			kPort = from.SrcDef.OverridePort
		} else {
			kPort = hostStrings[1]
		}
		from.SrcDef.ApiAddress = fmt.Sprintf("https://%s:%s", from.SrcDef.OverrideIp, kPort)
		k.Config.Clusters[translatedCluster].Server = from.SrcDef.ApiAddress
	}

	if from.SrcDef.Insecure {
		k.Config.Clusters[translatedCluster].CertificateAuthority = ""
		k.Config.Clusters[translatedCluster].CertificateAuthorityData = nil
		k.Config.Clusters[translatedCluster].InsecureSkipTLSVerify = true
	}
	if k.Config.CurrentContext == "" {
		k.Config.CurrentContext = translatedContext
	}
	return nil
}

func TruncateDestination(path string) error {
	dest, err := DestInit(path)
	if err != nil {
		return fmt.Errorf("unable to parse destination url: %v: %v", path, err)
	}

	err = dest.WriteConfig()
	if err != nil {
		return fmt.Errorf("unable to write destination config file: %v: %v", path, err)
	}
	return nil
}

func (k *KubeConfig) MergeOne(sourceKonfig *KubeConfig) error {

	err := k.CopyCurrentContext(sourceKonfig)
	if err != nil {
		return fmt.Errorf("unable merge config: %v: %v", sourceKonfig.Url, err)
	}

	err = k.WriteConfig()
	if err != nil {
		return fmt.Errorf("unable write config: %v: %v", k.Url, err)
	}
	return nil
}

func (k *KubeConfig) Delete(label string) error {

	err := k.ReadConfig()
	if err != nil {
		log.Fatalf("unable to parse destination: %v: %v", k.Url, err)
	}

	var removeCluster string
	var removeUser string
	if cluster, ok := k.Config.Contexts[label]; ok {
		removeCluster = cluster.Cluster
		removeUser = cluster.AuthInfo
		delete(k.Config.Contexts, label)
	} else {
		return fmt.Errorf("cluster named: %s not found", label)
	}

	if _, ok := k.Config.AuthInfos[removeUser]; ok {
		delete(k.Config.AuthInfos, removeUser)
	} else {
		log.Warnf("user %s not found. continuing", label)
	}

	if _, ok := k.Config.Clusters[removeCluster]; ok {
		delete(k.Config.Clusters, removeCluster)
	} else {
		log.Warnf("cluster %s not found. continuing", label)
	}

	err = k.WriteConfig()
	if err != nil {
		return fmt.Errorf("unable write config: %v: %v", k.Url, err)
	}
	return nil
}

func (k *KubeConfig) List(label string, source cfg.Source) error {

	sourceKonfig, err := DestInit(source.Source)
	if err != nil {
		return fmt.Errorf("unable to parse source: %v: %v", source.Source, err)
	}
	sourceKonfig.Label = label
	sourceKonfig.SrcDef = source

	err = sourceKonfig.ReadConfig()
	if err != nil {
		log.Fatalf("unable to parse destination: %v: %v", sourceKonfig.Url, err)
	}
	err = k.CopyCurrentContext(sourceKonfig)
	if err != nil {
		return fmt.Errorf("unable merge config: %v: %v", sourceKonfig.Url, err)
	}

	err = k.WriteConfig()
	if err != nil {
		return fmt.Errorf("unable write config: %v: %v", k.Url, err)
	}
	return nil
}
