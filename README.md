# Furyctl


git::https://gitlab+deploy-token-1:zU92yQ4reCQyqeE_i1KL@git.incubator.sh/sighup/fury.git//test-base-machine


# Go-discover

go get -u github.com/hashicorp/go-discover/cmd/discover

discover -q addrs provider=aws tag_key=KubernetesCluster tag_value=flyer-tech-cluster