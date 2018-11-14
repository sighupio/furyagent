#!/bin/bash

rm -r etcd | true
rm -r master | true

mkdir -p etcd
mkdir -p master

# Generating etcd init CA
cfssl gencert -initca files/ca-csr.json | cfssljson -bare etcd/ca
mv etcd/ca.pem etcd/ca.crt
mv etcd/ca-key.pem etcd/ca.key
rm etcd/ca.csr

# Generating master init CA
kubeadm alpha phase certs ca --cert-dir=$(pwd)/master
kubeadm alpha phase certs front-proxy-ca --cert-dir=$(pwd)/master
kubeadm alpha phase certs sa --cert-dir=$(pwd)/master