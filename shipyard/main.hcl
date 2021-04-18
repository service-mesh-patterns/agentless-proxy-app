variable "consul_k8s_cluster" {
  default = "dc1"
}

variable "consul_k8s_network" {
  default = "dc1"
}

k8s_cluster "dc1" {
  driver  = "k3s"

  nodes = 1

  network {
    name = "network.dc1"
  }
}

network "dc1" {
  subnet = "10.5.0.0/16"
}
 
module "consul" {
  source = "github.com/shipyard-run/blueprints/modules/kubernetes-consul"
}

k8s_config "app" {
  depends_on = ["module.consul"]

  cluster = "k8s_cluster.dc1"
  paths = [
    "./k8s_config/",
  ]

  wait_until_ready = true
}

output "KUBECONFIG" {
  value = k8s_config("dc1")
}