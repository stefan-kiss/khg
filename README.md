[![Build Status](https://app.travis-ci.com/stefan-kiss/khg.svg?branch=master)](https://travis-ci.com/stefan-kiss/khg)

# khg
Kubernetes (config file) hunter gatherer

## raison d'etre
This utility combines several operations that might be very common in some workflows.
A classic example is when you are testing your kubernetes install scripts.
In such example you might have one or several machines / vm's (whatever really) where you do consecutive kubernetes installs. 

Typical workflow would mean you will either ssh to the install machine where you have your kubeconfig file or transfer your kubeconfig file localy and then use clever bash scripts to merge it with your existing local kubeconfig file.

khg automates all of this:

- transferring the file
- merging into existing kubeconfig
- persistent configuration for existing sources (if you want to repeat the operation)
- some transformations to the source kubeconfig (needed in some cases where the source kubeconfig would have an internal api address as opposed to the external accessible api address)

## example usage

```shell
20:00   0[skiss@86dfj12 ~/khg]# khg get
Using config file: /Users/skiss/.khg.yaml
Error: accepts 1 arg(s), received 0
Usage:
  khg get <source> [flags]

Flags:
  -a, --api-address string   Use api address (usually external ip) instead of the one found in the source file.
  -h, --help                 help for get
  -i, --insecure             Will remove the CA from cluster and add the 'insecure-skip-tls-verify' flag.
  -l, --label string         Label for the entry. Will overwrite entry if exists.
  -r, --rewrite-api          Will rewrite api address using the host from the url and default port. Use api-address flag to overwrite this option and specify a custom one.

Global Flags:
      --config string      config file (default is $HOME/.khg.yaml)
  -L, --log-level string   Log Level. Default INFO (default "INFO")
  -p, --persistent         persist any changes to config file

accepts 1 arg(s), received 0
20:00   0[skiss@86dfj12 ~/khg]# khg get testvm/etc/rancher/k3s/k3s.yaml -r -p
Using config file: /Users/skiss/.khg.yaml
WARN[0000] rewrite-api flag is set. we will try to autodetect and rewrite api address. insecure is implied.
INFO[0000] using source: ssh://testvm/etc/rancher/k3s/k3s.yaml
INFO[0000] 2960 bytes copied
20:00   0[skiss@86dfj12 ~/khg]# khg list
Using config file: /Users/skiss/.khg.yaml
ConfigLabel          | SourceUrl                                   | KubernetesContext                              | ApiAddress
testvm               | ssh://testvm/etc/rancher/k3s/k3s.yaml       | default@testvm                                 | https://10.0.0.1:6443
20:00   0[skiss@86dfj12 ~/khg]# kubectl config view
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://10.0.0.1:6443
  name: default@testvm
contexts:
- context:
    cluster: default@testvm
    user: default@testvm
  name: default@testvm
current-context: default@testvm
kind: Config
preferences: {}
users:
- name: default@testvm
  user:
    client-certificate-data: REDACTED
    client-key-data: REDACTED
20:00   0[skiss@86dfj12 ~/khg]# 
```