/usr/local/bin/etcd --name s1 --data-dir /etcd-data --listen-client-urls http://0.0.0.0:2379 --advertise-client-urls http://0.0.0.0:2379 --listen-peer-urls http://0.0.0.0:2380 --initial-advertise-peer-urls http://0.0.0.0:2380 --initial-cluster s1=http://0.0.0.0:2380 --initial-cluster-token tkn --initial-cluster-state new 2> /dev/null &
sleep 10

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
killall -KILL etcd
sleep 10

furyctl restore etcd

echo "read after restoring: "
cat /etc/etcd/pki/ca.crt
cat /etc/etcd/pki/ca.key
/usr/local/bin/etcd --name s1 --data-dir /etcd-data --listen-client-urls http://0.0.0.0:2379 --advertise-client-urls http://0.0.0.0:2379 --listen-peer-urls http://0.0.0.0:2380 --initial-advertise-peer-urls http://0.0.0.0:2380 --initial-cluster s1=http://0.0.0.0:2380 --initial-cluster-token tkn --initial-cluster-state new 2> /dev/null &
sleep 10
etcdctl get foo
