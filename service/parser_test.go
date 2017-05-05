package service

import (
	"testing"

	etcd "github.com/coreos/etcd/client"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	sampleNode = &etcd.Node{
		Key: "/services/test/example.org",
		Value: `
		{
			"Name": "example.org",
			"User": "user",
			"Password": "password",
			"Ports": {
				"http": "111"
			}
		}
		`,
	}
	sampleInfoNode = &etcd.Node{
		Key: "/services/test/service_infos",
		Value: `
		{
			"critical": true
		}
		`,
	}
	sampleNodes = etcd.Nodes{sampleNode, sampleNode, sampleInfoNode}
)

var (
	sampleResult, _ = NewHost("example.org", Ports{"http": "111"}, "user", "password")
)

func TestBuildHostsFromNodes(t *testing.T) {
	service := buildServiceFromNodes(sampleNodes)
	Convey("Given a sample response with 2 nodes, we got 2 hosts", t, func() {
		hosts := service.Hosts
		So(len(hosts), ShouldEqual, 2)
		So(hosts[0], ShouldResemble, sampleResult)
		So(hosts[1], ShouldResemble, sampleResult)
	})
}

func TestBuildHostFromNode(t *testing.T) {
	host := buildHostFromNode(sampleNode)
	Convey("Given a sample response, we got a filled Host", t, func() {
		So(host, ShouldResemble, sampleResult)
	})
}

func TestBuildInfosFromNode(t *testing.T) {
	infos := buildInfosFromNode(sampleInfoNode)
	Convey("Given a sample response, we got a filled Infos", t, func() {
		So(infos.Critical, ShouldBeTrue)
	})
}
