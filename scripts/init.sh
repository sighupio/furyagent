#!/bin/bash

rm -r etcd | true
rm -r master | true

mkdir -p etcd
mkdir -p master

# Generating etcd init CA
cfssl gencert -initca ca-csr.json | cfssljson -bare etcd/ca
mv etcd/ca.pem etcd/ca.crt
mv etcd/ca-key.pem etcd/ca.key
rm etcd/ca.csr

# Generating master init CA
kubeadm alpha phase certs all --cert-dir=$(pwd)/master 
rm master/apiserver*
rm master/front-proxy-client*
rm -r master/etcd