package model

import "k8s.io/apimachinery/pkg/api/resource"

type Model struct {
	Workloads []Workload
	Services  []Service
	Summary   Summary
}

type Workload struct {
	Kind       string
	Name       string
	Namespace  string
	Replicas   int64
	Containers []Container
}

type Container struct {
	Name         string
	Image        string
	CPURequest   *resource.Quantity
	CPULimit     *resource.Quantity
	MemRequest   *resource.Quantity
	MemLimit     *resource.Quantity
	HasReadiness bool
	HasLiveness  bool
}

type Service struct {
	Name      string
	Namespace string
	Type      string
	Ports     []ServicePort
}

type ServicePort struct {
	Name       string
	Port       int64
	TargetPort string
	Protocol   string
}

type Summary struct {
	CPURequests resource.Quantity
	CPULimits   resource.Quantity
	MemRequests resource.Quantity
	MemLimits   resource.Quantity
}
