package service

import (
	"github.com/coreos/go-etcd/etcd"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	sampleNode = &etcd.Node {
		Key: "/services/test/example.org",
		Nodes: etcd.Nodes{
			etcd.Node{
				Key: "/services/test/example.org/user",
				Value: "user",
			},
			etcd.Node{
				Key: "/services/test/example.org/password",
				Value: "password",
			},
			etcd.Node{
				Key: "/services/test/example.org/port",
				Value: "port",
			},
		},
	}
	sampleNodes = etcd.Nodes{*sampleNode, *sampleNode}
)

func TestBuildHostsFromNodes(t *testing.T) {
	hosts := buildHostsFromNodes(sampleNodes)
	Convey("Given a sample response with 2 nodes, we got 2 hosts", t, func() {
		So(len(hosts), ShouldEqual, 2)
	})
}

func TestBuildHostFromNode(t *testing.T) {
	host := buildHostFromNode(sampleNode)
	Convey("Given a sample response, we got a filled Host", t, func() {
		So(host.User, ShouldEqual, "user")
		So(host.Password, ShouldEqual, "password")
		So(host.Port, ShouldEqual, "port")
	})
}
