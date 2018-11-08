echo 1 > /etc/etcd/pki/ca.crt
echo 1 > /etc/etcd/pki/ca.key
echo "ci sono" > /etcd-data/member/prova
echo "written before backup: "
etcdctl set foo v1 

echo "read before backup: "
cat /etc/etcd/pki/ca.crt
cat /etc/etcd/pki/ca.key
cat /etcd-data/member/prova
etcdctl get foo

furyctl backup etcd

echo "written before restoring: "
echo 2 > /etc/etcd/pki/ca.crt
echo 2 > /etc/etcd/pki/ca.key
etcdctl set foo v2

echo "read before restoring: "
cat /etc/etcd/pki/ca.crt
cat /etc/etcd/pki/ca.key
etcdctl get foo
cat /etcd-data/member/prova

furyctl restore etcd

echo "read after restoring: "
cat /etc/etcd/pki/ca.crt
cat /etc/etcd/pki/ca.key
cat /etcd-data/member/prova
etcdctl get foo
