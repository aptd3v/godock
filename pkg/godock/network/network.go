package network

import (
	"github.com/aptd3v/godock/pkg/godock/networkoptions"
	"github.com/docker/docker/api/types/network"
)

type NetworkConfig struct {
	Id      string
	Name    string
	Options *network.CreateOptions
}

func (n *NetworkConfig) String() string {
	return n.Name
}

func (n *NetworkConfig) SetOptions(setNOFns ...networkoptions.SetNetworkOptions) {
	for _, set := range setNOFns {
		set(n.Options)
	}
}
