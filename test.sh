echo 1 > /etc/etcd/pki/ca.crt
echo 1 > /etc/etcd/pki/ca.key
furyctl backup etcd
cat /etc/etcd/pki/ca.crt
cat /etc/etcd/pki/ca.key
echo 2 > /etc/etcd/pki/ca.crt
echo 2 > /etc/etcd/pki/ca.key
cat /etc/etcd/pki/ca.crt
cat /etc/etcd/pki/ca.key
furyctl restore etcd
cat /etc/etcd/pki/ca.crt
cat /etc/etcd/pki/ca.key
