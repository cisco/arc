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

package aws

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type securityRules struct {
	ingressRules []*ec2.IpPermission
	egressRules  []*ec2.IpPermission
}

func newSecurityRules(s *securityGroup) *securityRules {
	log.Debug("Initializing AWS Security Rules for Security Group '%s'", s.Name())
	r := &securityRules{}
	r.clear()
	return r
}

func (r *securityRules) populate(s *securityGroup) error {
	for _, rule := range *s.SecurityRules {
		for _, protocol := range rule.Protocols() {
			for _, port := range rule.Ports() {
				for _, direction := range rule.Directions() {
					for _, remote := range rule.Remotes() {

						toPort, fromPort, err := parsePort(port)
						if err != nil {
							return err
						}

						// If we have multiple availability zones, a single remote can generate multiple ipRanges.
						ipRanges, userIdGroupPairs, err := parseRemote(s.network, remote)
						if err != nil {
							return err
						}

						// Create a partial rule with the to, from and protocol. We will search existing rules
						// looking for a match with the partial rule.
						newRule := &ec2.IpPermission{
							IpProtocol: aws.String(protocol),
							FromPort:   aws.Int64(fromPort),
							ToPort:     aws.Int64(toPort),
						}

						var existingRule *ec2.IpPermission

						switch direction {
						case "ingress":
							existingRule = rulesContain(r.ingressRules, newRule, false)
						case "egress":
							existingRule = rulesContain(r.egressRules, newRule, false)
						}

						// If a rule exists with the same to port, from port and protocol, append the remote.
						if existingRule != nil {
							if ipRanges != nil {
								for _, ipRange := range ipRanges {
									existingRule.IpRanges = append(existingRule.IpRanges, ipRange)
								}
							}
							if userIdGroupPairs != nil {
								for _, userIdGroupPair := range userIdGroupPairs {
									existingRule.UserIdGroupPairs = append(existingRule.UserIdGroupPairs, userIdGroupPair)
								}
							}
							continue
						}

						// Otherwise create a new rule.
						if ipRanges != nil {
							newRule.IpRanges = ipRanges
						}
						if userIdGroupPairs != nil {
							newRule.UserIdGroupPairs = userIdGroupPairs
						}

						switch direction {
						case "ingress":
							r.ingressRules = append(r.ingressRules, newRule)
						case "egress":
							r.egressRules = append(r.egressRules, newRule)
						}
					}
				}
			}
		}
	}
	return nil
}

func (r *securityRules) set(secgroup *ec2.SecurityGroup) {
	if secgroup == nil {
		return
	}
	r.ingressRules = secgroup.IpPermissions
	r.egressRules = secgroup.IpPermissionsEgress
}

func (r *securityRules) clear() {
	r.ingressRules = nil
	r.egressRules = nil
}

func (r *securityRules) create(req *route.Request, s *securityGroup) error {
	log.Debug("created:\n%+v\n%+v", r.ingressRules, r.egressRules)
	// Ingress rules
	if r.ingressRules != nil && len(r.ingressRules) > 0 {
		msg.Info("Ingress SecurityRules Creation: %s", s.Name())
		ingressParams := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(s.Id()),
			IpPermissions: r.ingressRules,
		}
		if _, err := s.ec2.AuthorizeSecurityGroupIngress(ingressParams); err != nil {
			return err
		}
		msg.Detail("Created.")
		aaa.Accounting("Ingress SecurityRules created: %s", s.Id())
	}
	// Egress rules
	if r.egressRules != nil && len(r.egressRules) > 0 {
		msg.Info("Egress SecurityRules Creation: %s", s.Name())
		egressParams := &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       aws.String(s.Id()),
			IpPermissions: r.egressRules,
		}
		if _, err := s.ec2.AuthorizeSecurityGroupEgress(egressParams); err != nil {
			return err
		}
		msg.Detail("Created.")
		aaa.Accounting("Egress SecurityRules created: %s", s.Id())
	}
	return nil
}

func (r *securityRules) destroy(req *route.Request, s *securityGroup) error {
	// Ingress rules
	if r.ingressRules != nil && len(r.ingressRules) > 0 {
		msg.Info("Ingress SecurityRules Destruction: %s", s.Name())
		ingressParams := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       aws.String(s.Id()),
			IpPermissions: r.ingressRules,
		}
		if _, err := s.ec2.RevokeSecurityGroupIngress(ingressParams); err != nil {
			return err
		}
		msg.Detail("Destroyed.")
		aaa.Accounting("Ingress SecurityRules destroyed: %s", s.Id())
	}
	// Egress rules
	if r.egressRules != nil && len(r.egressRules) > 0 {
		msg.Info("Egress SecurityRules Destruction: %s", s.Name())
		egressParams := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       aws.String(s.Id()),
			IpPermissions: r.egressRules,
		}
		if _, err := s.ec2.RevokeSecurityGroupEgress(egressParams); err != nil {
			return err
		}
		msg.Detail("Destroyed.")
		aaa.Accounting("Egress SecurityRules destroyed: %s", s.Id())
	}
	return nil
}

func (r *securityRules) compareEqual(cfg *securityRules, quiet bool, name string, flags ...string) bool {
	found := false
	doAudit := (len(flags) == 0 || flags[0] == "")
	a := &aaa.Audit{}
	if !doAudit {
		a = aaa.AuditBuffer[flags[0]]
	}
	for _, rule := range cfg.ingressRules {
		if rulesContain(r.ingressRules, rule, true) == nil {
			if !quiet {
				msg.Detail("Security Group: %q\nIngress rule is configured but not deployed\n%s", name, indentRule(rule))
			}
			if !doAudit {
				a.Audit(aaa.Mismatched, "\n> Security Group: %q - Ingress rule is configured but not deployed\n%s", name, indentRule(rule))
			}
			found = true
		}
	}
	for _, rule := range r.ingressRules {
		if rulesContain(cfg.ingressRules, rule, true) == nil {
			if !quiet {
				msg.Detail("Security Group: %q\nIngress rule is deployed but not configured\n%s", name, indentRule(rule))
			}
			if !doAudit {
				a.Audit(aaa.Mismatched, "\n> Security Group: %q - Ingress rule is deployed but not configured\n%s", name, indentRule(rule))
			}
			found = true
		}
	}
	for _, rule := range cfg.egressRules {
		if rulesContain(r.egressRules, rule, true) == nil {
			if !quiet {
				msg.Detail("Security Group: %q\nEgress rule is configured but not deployed\n%s", name, indentRule(rule))
			}
			if !doAudit {
				a.Audit(aaa.Mismatched, "\n> Security Group: %q - Egress rule is configured but not deployed\n%s", name, indentRule(rule))
			}
			found = true
		}
	}
	for _, rule := range r.egressRules {
		if rulesContain(cfg.egressRules, rule, true) == nil {
			if !quiet {
				msg.Detail("Security Group: %q\nEgress rule is deployed but not configured\n%s", name, indentRule(rule))
			}
			if !doAudit {
				a.Audit(aaa.Mismatched, "\n> Security Group: %q - Egress rule is deployed but not configured\n%s", name, indentRule(rule))
			}
			found = true
		}
	}
	if !found {
		if !quiet {
			msg.Detail("No inconsistencies found")
		}
		return true
	}
	return false
}

func (r *securityRules) info() {
	if len(r.ingressRules) > 0 {
		r.infoIngress()
	}
	if len(r.egressRules) > 0 {
		r.infoEgress()
	}
}

func (r *securityRules) infoIngress() {
	msg.Info("Ingress Rules")
	msg.IndentInc()
	for _, ingressRule := range r.ingressRules {
		msg.Info("Ingress Rule")
		if ingressRule.IpProtocol != nil {
			msg.Detail("%-20s\t%s", "protocol", *ingressRule.IpProtocol)
		}
		for _, ipRange := range ingressRule.IpRanges {
			if ipRange.CidrIp != nil {
				msg.Detail("%-20s\t%s", "cidr", *ipRange.CidrIp)
			}
		}
		if ingressRule.ToPort != nil {
			msg.Detail("%-20s\t%d", "to", *ingressRule.ToPort)
		}
		if ingressRule.FromPort != nil {
			msg.Detail("%-20s\t%d", "from", *ingressRule.FromPort)
		}
	}
	msg.IndentDec()
}

func (r *securityRules) infoEgress() {
	msg.Info("Egress Rules")
	msg.IndentInc()
	for _, egressRule := range r.egressRules {
		msg.Info("Egress Rule")
		if egressRule.IpProtocol != nil {
			msg.Detail("%-20s\t%s", "protocol", *egressRule.IpProtocol)
		}
		for _, ipRange := range egressRule.IpRanges {
			if ipRange.CidrIp != nil {
				msg.Detail("%-20s\t%s", "cidr", *ipRange.CidrIp)
			}
		}
		if egressRule.ToPort != nil {
			msg.Detail("%-20s\t%d", "to", *egressRule.ToPort)
		}
		if egressRule.FromPort != nil {
			msg.Detail("%-20s\t%d", "from", *egressRule.FromPort)
		}
	}
	msg.IndentDec()
}

func parsePort(port string) (toPort, fromPort int64, err error) {
	ports := strings.Split(port, ":")
	if ports[0] == "" {
		err = fmt.Errorf("Malformed port: %s", port)
		return
	}

	if fromPort, err = strconv.ParseInt(strings.TrimSpace(ports[0]), 10, 64); err != nil {
		return
	}
	toPort = fromPort

	if len(ports) > 1 {
		if toPort, err = strconv.ParseInt(strings.TrimSpace(ports[1]), 10, 64); err != nil {
			return
		}
	}
	return
}

func parseRemote(network resource.Network, remote string) ([]*ec2.IpRange, []*ec2.UserIdGroupPair, error) {
	ipRanges := []*ec2.IpRange{}
	userIdGroupPair := []*ec2.UserIdGroupPair{}

	s := strings.Split(remote, ":")
	if len(s) != 2 {
		return nil, nil, fmt.Errorf("Malformed remote: %s", remote)
	}
	protocol := s[0]
	dest := s[1]

	switch protocol {
	case "subnet_group":
		subnetGroup := network.SubnetGroups().Find(dest)
		if subnetGroup == nil {
			return nil, nil, fmt.Errorf("Unknown subnet_group: %s", dest)
		}
		for _, subnet := range subnetGroup.Subnets() {
			ipRanges = append(ipRanges, &ec2.IpRange{CidrIp: aws.String(subnet.CidrBlock())})
		}
		return ipRanges, nil, nil
	case "cidr":
		ip := network.CidrAlias(dest)
		if ip != "" {
			dest = ip
		}
		ipRanges = append(ipRanges, &ec2.IpRange{CidrIp: aws.String(dest)})
		return ipRanges, nil, nil
	case "cidr_group":
		group := network.CidrGroup(dest)
		if group == nil {
			return nil, nil, fmt.Errorf("Cidr group %s does not exist", dest)
		}
		for _, v := range group {
			ip := v
			if ip != "" {
				dest = ip
			}
			ipRanges = append(ipRanges, &ec2.IpRange{CidrIp: aws.String(dest)})
		}
		return ipRanges, nil, nil
	case "security_group":
		securityGroup := network.SecurityGroups().Find(dest)
		if securityGroup == nil {
			return nil, nil, fmt.Errorf("Unknown security_group: %s", dest)
		}
		userIdGroupPair = append(userIdGroupPair, &ec2.UserIdGroupPair{GroupId: aws.String(securityGroup.Id())})
		return nil, userIdGroupPair, nil
	}
	return nil, nil, fmt.Errorf("Unknown remote: %s", protocol)
}

func rulesContain(rules []*ec2.IpPermission, rule *ec2.IpPermission, checkIpRange bool) *ec2.IpPermission {
	for _, r := range rules {
		if ruleEqual(r, rule, checkIpRange) {
			return r
		}
	}
	return nil
}

func protocolEqual(r *string, l *string) bool {
	if r == nil && l == nil {
		return true
	}
	if r == nil || l == nil {
		return false
	}
	return *r == *l
}

func portEqual(r *int64, l *int64) bool {
	if r == nil && l == nil {
		return true
	}
	if r == nil && *l == -1 {
		return true
	}
	if l == nil && *r == -1 {
		return true
	}
	return *r == *l
}

func ipRangeEqual(r *ec2.IpRange, l *ec2.IpRange) bool {
	if r == nil && l == nil {
		return true
	}
	if r == nil || l == nil {
		return false
	}
	return *r.CidrIp == *l.CidrIp
}

func ipRangesEqual(r []*ec2.IpRange, l []*ec2.IpRange) bool {
	if r == nil && l == nil {
		return true
	}
	if len(r) != len(l) {
		return false
	}
	for _, rIp := range r {
		match := false
		for _, lIp := range l {
			if ipRangeEqual(rIp, lIp) {
				match = true
			}
		}
		if !match {
			return false
		}
	}
	return true
}

func ruleEqual(r *ec2.IpPermission, l *ec2.IpPermission, checkIpRanges bool) bool {
	if !protocolEqual(r.IpProtocol, l.IpProtocol) {
		return false
	}
	if !portEqual(r.FromPort, l.FromPort) {
		return false
	}
	if !portEqual(r.ToPort, l.ToPort) {
		return false
	}
	if !checkIpRanges {
		return true
	}
	if !ipRangesEqual(r.IpRanges, l.IpRanges) {
		return false
	}
	return true
}

func indentRule(rule *ec2.IpPermission) string {
	var result string
	for _, s := range strings.Split(fmt.Sprintf("%+v", rule), "\n") {
		result += "\t" + s + "\n"
	}
	return result
}
