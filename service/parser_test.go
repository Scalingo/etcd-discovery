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
			"name": "public.dev",
			"service_name": "test-service",
			"User": "user",
			"password": "password",
			"ports": {
				"http": "10000"
			},
			"public": true,
			"critical": true,
			"private_hostname": "test-private.dev",
			"private_ports": {
				"http": "20000"
			},
			"uuid": "1234"
		}
		`,
	}
	sampleInfoNode = &etcd.Node{
		Key: "/services_infos/test",
		Value: `
		{
			"critical": true
		}
		`,
	}
	sampleNodes = etcd.Nodes{sampleNode, sampleNode}
)

var (
	sampleResult = genHost("test")
)

func TestBuildHostsFromNodes(t *testing.T) {
	Convey("Given a sample response with 2 nodes, we got 2 hosts", t, func() {
		hosts, err := buildHostsFromNodes(sampleNodes)
		So(err, ShouldBeNil)
		So(len(hosts), ShouldEqual, 2)
		So(hosts[0], ShouldResemble, sampleResult)
		So(hosts[1], ShouldResemble, sampleResult)
	})
}

func TestBuildHostFromNode(t *testing.T) {
	Convey("Given a sample response, we got a filled Host", t, func() {
		host, err := buildHostFromNode(sampleNode)
		So(err, ShouldBeNil)
		So(host, ShouldResemble, sampleResult)
	})
}

func TestBuildServiceFromNode(t *testing.T) {
	Convey("Given a sample response, we got a filled Infos", t, func() {
		infos, err := buildServiceFromNode(sampleInfoNode)
		So(err, ShouldBeNil)
		So(infos.Critical, ShouldBeTrue)
	})
}
