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

func (i *Instance) Audit(flags ...string) error {
	if i.Derived().Pod().Cluster().AuditIgnore() {
		return nil
	}
	if err := i.Derived().PreAudit(flags...); err != nil {
		return err
	}
	if err := i.Derived().MidAudit(flags...); err != nil {
		return err
	}
	if err := i.Derived().PostAudit(flags...); err != nil {
		return err
	}
	return nil
}

func (i *Instance) PreAudit(flags ...string) error {
	return nil
}

func (i *Instance) MidAudit(flags ...string) error {
	if err := i.providerInstance.Audit(flags...); err != nil {
		return err
	}
	if !i.Created() {
		return nil
	}
	if err := i.volumes.Audit(flags...); err != nil {
		return err
	}
	return nil
}

func (i *Instance) PostAudit(flags ...string) error {
	return nil
}
