# Furyagent

## Install

You can find `furyagent` binaries on the [Releases page](https://github.com/sighup-io/furyagent/releases).

Supported architectures are (64 bit):

-   `linux`
-   `darwin`

Download right binary for your architecture and add it to your PATH. Assuming it's downloaded in your
`~/Downloads` folder, you can run following commands (replacing `{arch}` with your architecture):

```
chmod +x  ~/Downloads/furyagent-{arch}-amd64 && mv ~/Downloads/furyagent-{arch}-amd64 /usr/local/bin/furyagent
```

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


### SSH management

In order to enable this feature, you have to add this configuration to the `furyagent.yml` file

```yaml
clusterComponent:
    sshKeys:
        adapter:
            name: "github"  # you can use also "http" as adapter  name but you'll need to specify also the "uri" field as well because `non github` adapter is not well known 
        user: "sighup" # the user that will be created on the system for storing public keys
        tempDir: "/tmp" # the temp dir that will be used to put the downloaded file
        localDirConfigs: "secrets/ssh" # where the code will look for searching the file ssh-users.yml
```


`ssh-users.yml` should have the following structure:

```yaml
users:
    - name: lucazecca
      github_id: lzecca78
    - name: philippe
      github_id: phisco
    - name: samuele
      github_id: nutellinoit
    - name: lucanovara
      github_id: lnovara
```

once do that, all you have to do is 

to put the `ssh-users.yml` on the bucket s3:

`furyagent init --config ssh/furyagent.yml ssh-keys`

on the nodes, just create a cron entry like the following:

`*/30 * * * * furyagent configure --config <path>/furyagent.yml ssh-keys --overwrite true`

And it will do the following actions: 

1. fetch the ssh-users.yml from s3 bucket
2. get the adapter from furyagent.yml (github doesn't require uri, because is well known, http require also a uri field to be put in the adapter struct )
3. once get the adapter (name, uri) it will fetch from it the same github structure: so 1 file.keys for each user
4. create the system user (if doesn't exist) checking on which os is launched (redhat bases, debian based) in order to use the correct command flags
5. create a temporary authorized_keys
6. if the step 3 goes well, it will override the authorized_keys file of the user, otherwise it won't
of course the steps 4 is be ignored if the user already exists