//
// Copyright (c) 2017, Cisco Systems
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice, this
//   list of conditions and the following disclaimer in the documentation and/or
//   other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
// ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//

package arc

import (
	"fmt"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"

	"github.com/cisco/arc/pkg/aws"
	"github.com/cisco/arc/pkg/mock"
	//"github.com/cisco/arc/pkg/gcp"
	//"github.com/cisco/arc/pkg/azure"
)

type dataCenter struct {
	*resource.Resources
	*config.DataCenter
	arc     *arc
	network *network
	compute *compute
	dns     *dns
}

// newDataCenter is the constructor for a dataCenter object. It returns a non-nil error upon failure.
func newDataCenter(arc *arc, cfg *config.DataCenter) (*dataCenter, error) {
	if cfg == nil {
		return nil, nil
	}
	log.Debug("Initializing Datacenter")

	// Validate the config.DataCenter object.
	if cfg.Provider == nil {
		return nil, fmt.Errorf("The provider element is missing from the datacenter configuration")
	}

	if cfg.Network == nil && cfg.Compute != nil {
		return nil, fmt.Errorf("The network element is missing from the datacenter configuration")
	}

	d := &dataCenter{
		Resources:  resource.NewResources(),
		DataCenter: cfg,
		arc:        arc,
	}

	vendor := cfg.Provider.Vendor
	var err error
	var p provider.DataCenter

	switch vendor {
	case "mock":
		p, err = mock.NewDataCenterProvider(cfg)
	case "aws":
		p, err = aws.NewDataCenterProvider(cfg)
	//case "azure":
	//	p, err = azure.NewDataCenterProvider(cfg)
	//case "gcp":
	//	p, err = gcp.NewDataCenterProvider(cfg)
	default:
		err = fmt.Errorf("Unknown vendor %q", vendor)
	}
	if err != nil {
		return nil, err
	}

	// The network and compute name field is a convenience field in the config struct and
	// is set at run time to the arc name.
	if cfg.Network != nil {
		cfg.Network.SetName(arc.Name())
	}
	if cfg.Compute != nil {
		cfg.Compute.SetName(arc.Name())
	}

	d.network, err = newNetwork(d, p, cfg.Network)
	if err != nil {
		return nil, err
	}
	if d.network != nil {
		d.Append(d.network)
	}

	d.compute, err = newCompute(d, p, cfg.Compute)
	if err != nil {
		return nil, err
	}
	if d.compute != nil {
		d.Append(d.compute)
	}

	return d, nil
}

// Arc satisfies the resource.DataCenter interface and provides access
// to datacenter's parent.
func (d *dataCenter) Arc() resource.Arc {
	return d.arc
}

// Network satisfies the resource.DataCenter interface and provides access
// to datacenter's network.
func (d *dataCenter) Network() resource.Network {
	if d.network == nil {
		return nil
	}
	return d.network
}

// Compute satisfies the resource.DataCenter interface and provides access
// to datacenter's compuet .
func (d *dataCenter) Compute() resource.Compute {
	if d.compute == nil {
		return nil
	}
	return d.compute
}

// associate the DataCenter resource with the Dns resource.
func (d *dataCenter) associate(r *dns) {
	if r == nil {
		return
	}
	d.dns = r
}

// Dns providess acess to the Dns resource.
func (d *dataCenter) Dns() resource.Dns {
	if d.dns == nil {
		return nil
	}
	return d.dns
}

// Route satisfies the embedded resource.Resource interface in resource.DataCenter.
// DataCenter does not directly terminate a request so only handles load and info
// requests from it's parent.  All other commands are routed to arc's children.
func (d *dataCenter) Route(req *route.Request) route.Response {
	log.Route(req, "DataCenter")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "network":
		if d.Network() == nil {
			msg.Error("Network not defined in the config file")
			return route.OK
		}
		return d.Network().Route(req.Pop())
	case "subnet", "secgroup":
		if d.Network() == nil {
			msg.Error("Network not defined in the config file")
			return route.OK
		}
		return d.Network().Route(req)
	case "compute":
		if d.Compute() == nil {
			msg.Error("Compute not defined in the config file")
			return route.OK
		}
		return d.Compute().Route(req.Pop())
	case "keypair", "cluster", "pod", "instance", "volume", "eip":
		if d.Compute() == nil {
			msg.Error("Compute not defined in the config file")
			return route.OK
		}
		return d.Compute().Route(req)
	default:
		panic("Internal Error: Unknown path " + req.Top())
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		return d.RouteInOrder(req)
	case route.Info:
		d.info(req)
		return route.OK
	case route.Audit:
		return d.RouteInOrder(req)
	default:
		panic("Internal Error: Unknown command " + req.Command().String())
	}
	return route.FAIL
}

func (d *dataCenter) info(req *route.Request) {
	if d.Destroyed() {
		return
	}
	msg.Info("DataCenter")
	msg.IndentInc()
	d.RouteInOrder(req)
	msg.IndentDec()
}
