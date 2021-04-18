module github.com/service-mesh-patterns/agentless-proxy-app

go 1.15

require (
	github.com/hashicorp/consul v1.8.0
	//github.com/hashicorp/consul v0.0.0-20210331183933-6e69829edbde
	github.com/hashicorp/consul/api v1.8.0
	github.com/hashicorp/go-hclog v0.14.1
	github.com/nicholasjackson/env v0.6.0
)

//replace github.com/hashicorp/consul/api v1.8.0 => github.com/hashicorp/consul/api v0.0.0-20210331183933-6e69829edbde
