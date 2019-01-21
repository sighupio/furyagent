# Furyagent

## Install
Get the right binary for you in the [latest release](https://git.incubator.sh/sighup/furyagent/tags)

## Usage

```
Available Commands:
  backup        Executes backups
  configure     Executes configuration
  help          Help about any command
  init          Executes initialization, uploads ca files
  parsed-config Prints the parsed furyagent.yaml file
  printDefault  Prints a basic Furyfile used to generate an INFRA project
  restore       Executes restores
  version       Prints the client version information
Flags:
      --config furyagent.yaml   config file (default is furyagent.yaml) (default "furyagent.yml")
  -h, --help                    help for furyagent

furyagent
├── init
│   ├── node
│   ├── etcd
│   ├── openVPN
│   └── master
├── configure
│   ├── etcd
│   ├── openVPN
│   └── master
├── backup
│   └── etcd
└── restore
    └── etcd
```

## Workflow
1. Write a [`furyagent.yml`](furyagent.yml)
2. Generate certificates (by hand at the moment)
3. `furyagent init -d /path/to/cert/dir --config /path/to/furyagent.yml [etcd|master]` to upload certificates
4. Then on the nodes: `furyagent configure --config /path/to/furyagent.yml [etcd|master]` to download the certificates to the correct directory specified in the config file
5. if needed: to backup the state of etcd through `furyagent backup --config /path/to/furyagent.yml etcd`
6. if needed: to restore the state of etcd, stop etcd, run `furyagent restore --config /path/to/furyagent.yml etcd`, restart etcd


## Contributing
We still use `go mod` as golang package manager. Once you have that installed you can run `go mod vendor` and `go build` or `go install` should run without problems

# Storage
There is going to be one and only one bucket per cluster.

```
S3 bucket
├── etcd
│   ├── node-1
│   │   └── snapshot.db
│   ├── node-2
│   └── node-3
├── cluster-backup
│   ├── full-20181002120049
│   │   ├── full-20181002120049.tar.gz
│   │   ├── full-20181002120049-logs.gz
│   │   └── ark-backup.json
│   ├── full-20181003120049
│   │   ├── full-20181003120049.tar.gz
│   │   ├── full-20181003120049-logs.gz
│   │   └── ark-backup.json
│   └── full-20181004120049
│       ├── full-20181004120049.tar.gz
│       ├── full-20181004120049-logs.gz
│       └── ark-backup.json
├── nodes
│   ├── discovery.txt
│   └── token.txt
├── users
│   ├── giacomo.conf
│   ├── jacopo.conf
│   ├── luca.conf
│   ├── philippe.conf
│   └── berat.conf
├── configurations
│   ├── kustomization.yaml
│   ├── audit.yaml
│   ├── nodeSelector.yaml
│   └── kubeadm.yml
└── pki
    ├── etcd
    │   ├── ca.crt
    │   └── ca.key
    ├── master
    │   ├── sa.key
    │   ├── sa.pub
    │   ├── front-proxy-ca.crt
    │   ├── front-proxy-ca.key
    │   ├── ca.crt
    │   └── ca.key
    └── vpn
        ├── ca.crt
        ├── ca.key
        ├── server.crt
        └── server.key

```

For ARK volume backup using restic backup is necessary a different bucket then this one.

