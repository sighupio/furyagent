provider "scaleway" {
  region       = "ams1"
}

resource "scaleway_server" "furyagent-test"{
  name = "furyagent"
  image = "${data.scaleway_image.furyagent-test.id}"
  type = "C2S"
}

data "scaleway_image" "furyagent-test" {
  architecture = "x86_64"
  name         = "Ubuntu Bionic"
}

resource "scaleway_ip" "furyagent-test-ip" {
  server = "${scaleway_server.furyagent-test.id}"
}

resource "scaleway_security_group" "furyagent-test-http" {
  name        = "http"
  description = "allow HTTP and HTTPS traffic"
}

resource "scaleway_security_group_rule" "furyagent-test-http_accept" {
  security_group = "${scaleway_security_group.furyagent-test-http.id}"

  action    = "accept"
  direction = "inbound"
  ip_range  = "0.0.0.0/0"
  protocol  = "TCP"
  port      = 80
}

resource "scaleway_security_group_rule" "furyagent-test-https_accept" {
  security_group = "${scaleway_security_group.furyagent-test-http.id}"

  action    = "accept"
  direction = "inbound"
  ip_range  = "0.0.0.0/0"
  protocol  = "TCP"
  port      = 443
}


locals {
  inventory = <<EOF
etcd_node ansible_host=${scaleway_ip.furyagent-test-ip.ip}

[etcd]
etcd_node

[all:vars]
ansible_user=root
EOF
}

output "inventory" {
  value = "${local.inventory}"
}
