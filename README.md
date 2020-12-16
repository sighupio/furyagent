# Furyagent v0.2.2

[![Build Status](http://ci.sighup.io/api/badges/sighupio/furyagent/status.svg?ref=refs/tags/v0.2.2)](http://ci.sighup.io/sighupio/furyagent)

## Install

You can find `furyagent` binaries on the [Releases page](https://github.com/sighupio/furyagent/releases).

Supported architectures are (64 bit):

-   `linux`
-   `darwin`

Download right binary for your architecture and add it to your PATH. Assuming it's downloaded in your
`~/Downloads` folder, you can run following commands (replacing `{arch}` with your architecture):

```shell
chmod +x  ~/Downloads/furyagent-{arch}-amd64 && mv ~/Downloads/furyagent-{arch}-amd64 /usr/local/bin/furyagent
```

If you are using MacOS you can also install `furyagent` using the `brew` package manager:

```shell
brew tap sighupio/furyagent
brew install furyagent
```

## Usage

```shell
Available Commands:
  backup        Executes backups
  configure     Executes configuration
  help          Help about any command
  init          Executes initialization, uploads ca files
  parsed-config Prints the parsed furyagent.yaml file
  restore       Executes restores
  version       Prints the client version information
Flags:
      --config furyagent.yaml   config file (default is furyagent.yaml) (default "furyagent.yml")
  -h, --help                    help for furyagent

furyagent
├── init
│   ├── etcd
│   ├── master
│   ├── openvpn
│   └── ssh-keys
├── configure
│   ├── etcd
│   ├── master
│   ├── openvpn
│   ├── openvpn-client
│   └── ssh-keys
├── backup
│   ├── etcd
│   └── master
└── restore
    ├── etcd
    └── master
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

## Storage

There is going to be one and only one bucket per cluster.

```shell
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

### OpenVPN users management

In order to enable this feature, add the following configuration to the
`furyagent.yml` file:

```yaml
clusterComponent:
    openvpn:
        server:
            - 1.2.3.4
            - 5.6.7.8
```

then you can create an OpenVPN client configuration with the following command:

```shell
furyagent configure openvpn-client --client-name foo --config /etc/fury/furyagent.yml > foo.ovpn
```

the newly created client certificate is saved to the object storage to keep
track of all the certificates issued by the OpenVPN CA in case of revocation.

The resulting `*.ovpn` file can be fed to any OpenVPN client (such as
Tunnelblick) to connect to the OpenVPN server.

If you need to revoke access to any user, you can do it with the following command:

```shell
furyagent config openvpn-client --client-name foo --revoke --config /etc/fury/furyagent.yml
```

### SSH management

In order to enable this feature, you have to add the following configuration to the `furyagent.yml` file:

```yaml
clusterComponent:
    sshKeys:
        adapter:
            name: "github" # you can use also "http" as adapter  name but you'll need to specify also the "uri" field as well because `non github` adapter is not well known
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
    - name: ramiro
      github_id: ralgozino
```

once you've done that, all you have to do is to upload the `ssh-users.yml` to the S3 bucket:

`furyagent init --config ssh/furyagent.yml ssh-keys`

On the nodes, you must create a cron entry like the following:

`*/30 * * * * furyagent configure --config <path>/furyagent.yml ssh-keys --overwrite true`

and it will do the following actions:

1. fetch the `ssh-users.yml` from s3 bucket
2. get the adapter from `furyagent.yml` (GitHub doesn't require an uri, because it's well known, http requires also a uri field to be put in the adapter struct )
3. once it gets the adapter (name, uri) it will fetch from it the same GitHub structure: a `file.keys` for each user
4. create the system user (if doesn't exist) checking on which OS is launched (RedHat based, Debian based) in order to use the correct command flags
5. create a temporary `authorized_keys`
6. if the step 3 goes well, it will override the `authorized_keys` file of the user, otherwise it won't

## furyagent list openvpn client certificate

`furyagent --config path/to/furyagent.yml configure openvpn-client --list`

This will be the output:

```bash
2020-03-19 17:09:00.727031 I | storage.go:146: Item pki/vpn-client/revoked/luca.zecca.crt found [size: 1103]
2020-03-19 17:09:00.727195 I | storage.go:147: Saving item pki/vpn-client/revoked/luca.zecca.crt ...
2020-03-19 17:09:00.830450 I | storage.go:146: Item pki/vpn-client/simone.messina.crt found [size: 1107]
2020-03-19 17:09:00.830470 I | storage.go:147: Saving item pki/vpn-client/simone.messina.crt ...
2020-03-19 17:09:00.948095 I | storage.go:146: Item pki/vpn/ca.crl found [size: 597]
2020-03-19 17:09:00.948113 I | storage.go:147: Saving item pki/vpn/ca.crl ...
2020-03-19 17:09:01.046877 I | storage.go:146: Item pki/vpn/ca.crl found [size: 597]
2020-03-19 17:09:01.046893 I | storage.go:147: Saving item pki/vpn/ca.crl ...
+----------------+------------+------------+---------+--------------------------------+
|      USER      | VALID FROM |  VALID TO  | EXPIRED |            REVOKED             |
+----------------+------------+------------+---------+--------------------------------+
| luca.zecca     | 2020-03-19 | 2021-03-19 | false   | true 2020-03-19 14:47:40 +0000 |
|                |            |            |         | UTC                            |
+----------------+------------+------------+---------+--------------------------------+
| simone.messina | 2020-03-19 | 2021-03-19 | false   | false 0001-01-01 00:00:00      |
|                |            |            |         | +0000 UTC                      |
+----------------+------------+------------+---------+--------------------------------+
```

you can also add `--output=json` to the command above and than you can obtain a json output:

```bash
go run main.go --config=ssh/furyagent.yml configure openvpn-client --list --output=json
2020-03-19 18:37:25.204840 I | storage.go:146: Item pki/vpn-client/revoked/luca.zecca.crt found [size: 1103]
2020-03-19 18:37:25.204988 I | storage.go:147: Saving item pki/vpn-client/revoked/luca.zecca.crt ...
2020-03-19 18:37:25.314691 I | storage.go:146: Item pki/vpn-client/simone.messina.crt found [size: 1107]
2020-03-19 18:37:25.314715 I | storage.go:147: Saving item pki/vpn-client/simone.messina.crt ...
2020-03-19 18:37:25.432634 I | storage.go:146: Item pki/vpn/ca.crl found [size: 597]
2020-03-19 18:37:25.432655 I | storage.go:147: Saving item pki/vpn/ca.crl ...
2020-03-19 18:37:25.537314 I | storage.go:146: Item pki/vpn/ca.crl found [size: 597]
2020-03-19 18:37:25.537341 I | storage.go:147: Saving item pki/vpn/ca.crl ...
[{"User":"luca.zecca","Valid_from":"2020-03-19","Valid_to":"2021-03-19","Expired":false,"Revoked":{"Revoked":true,"RevokeTime":"2020-03-19T14:47:40Z"}},{"User":"simone.messina","Valid_from":"2020-03-19","Valid_to":"2021-03-19","Expired":false,"Revoked":{"Revoked":false,"RevokeTime":"0001-01-01T00:00:00Z"}}]
```

