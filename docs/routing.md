Routing
=======

Arc
----

Request
  - arc help
  - arc <dc> <sub-target> <action> <flag>*
  - arc <dc> <sub-target> <target>+ <action> <flag>*

Target
  - arc

Sub-targets
  - network, subnet, secgroup, compute, keypair, cluster, pod, instance   -> DataCenter (arc/datacenter.go)
  - dns                                                                   -> Dns        (arc/dns.go)

Actions
  - help
  - config
  - info


DataCenter
----------

Target
  - n/a

Sub-targets
  - network, subnet, secgroup                 -> Network (arc/network.go)
  - compute, keypair, cluster, pod, instance  -> Compute (arc/compute.go)

Actions
  - info


Network
-------

Target
  - arc network

Sub-targets
  - raw network       -> <provider>/network (aws/network.go)
      aws:
        vpc           -> <provider>/vpc             (<provider)/vpc.go)
        rt            -> <provider>/routeTables     (<provider>/route_tables.go)
                         <provider>/routeTable      (<provider>/route_table.go)
        igw           -> <provider>/internetGateway (<provider>/internet_gateway.go)

  - subnet            -> SubnetGroups       (arc/subnet_groups.go, arc/subnet_group.go)
  - secgroup          -> SecurityGroups     (arc/security_groups.so, arc/securit_group.go)

  - raw network post  -> <provider>/NetworkPost
      aws:
        ngw           -> <provider>/natGateway (<provider>/nat_gateway.go)

Actions
  - create
  - destroy
  - help
  - config
  - info

Flags
  - create test             Noop
  - destroy test            Noop

AWS: vpc
--------

Target
  - arc network vpc

Actions
  - create
  - destroy
  - help
  - info

Flags
  - create test             Noop
  - destroy test            Noop


AWS: routeTables
----------------

Target
  - arc network routetable

Sub-target
  - [routetable name]     -> <provider>/routeTable (<provider>/route_table.go)

Actions
  - create
  - destroy
  - help
  - info

Flags
  - create test             Noop
  - destroy test            Noop


AWS: routeTable
---------------

Target
  - arc network routetable [name]

Actions
  - create
  - destroy
  - help
  - info

Flags
  - create test             Noop
  - destroy test            Noop


AWS: internetGateway
--------------------

Target
  - arc network internetgateway

Actions
  - create
  - destroy
  - help
  - info

Flags
  - create test             Noop
  - destroy test            Noop


SubnetGroups
------------

Target
  - arc network subnet
  - arc subnet

Sub-target
  - [SubnetGroup name]     -> arc/subnet_group.go

Actions
  - create
  - destroy
  - help
  - config
  - info

Flags
  - create test             Noop
  - destroy test            Noop


SubnetGroup
-----------

Target
  - arc network subnet [name]
  - arc subnet [name]

Actions
  - create
  - destroy
  - help
  - config
  - info

Flags
  - create test             Noop
  - destroy test            Noop


SecurityGroups
--------------

Target
  - arc network secgroup
  - arc secgroup

Sub-target
  - [SecurityGroup name]     -> arc/subnet_group.go

Actions
  - create
  - update/provision
  - audit
  - destroy
  - help
  - config
  - info

Flags
  - create test             Noop
  - create norules          Do not create the security rules
  - destroy test            Noop


SubnetGroup
-----------

Target
  - arc network secgroup [name]
  - arc secgroup [name]

Actions
  - create
  - update/provision
  - audit
  - destroy
  - help
  - config
  - info

Flags
  - create test             Noop
  - destroy test            Noop


Compute
-------

Target
  - arc compute

Actions
  - help
  - config
  - info

