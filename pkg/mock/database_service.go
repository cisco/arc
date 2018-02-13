//
// Copyright (c) 2018, Cisco Systems
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

package mock

import (
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type databaseService struct {
	*config.DatabaseService
	opt options
}

func newDatabaseService(cfg *config.DatabaseService, p *databaseServiceProvider) (resource.ProviderDatabaseService, error) {
	log.Info("Initializing Mock Database Service")
	dbs := &databaseService{
		DatabaseService: cfg,
		opt:             options{p.Provider.Data},
	}
	if dbs.opt.err("dbs.New") {
		return nil, err{"dbs.New"}
	}
	return dbs, nil
}

func (dbs *databaseService) Load() error {
	log.Info("Loading Mock Database Service")
	if dbs.opt.err("dbs.Load") {
		return err{"dbs.Load"}
	}
	return nil
}

func (dbs *databaseService) Audit(flags ...string) error {
	msg.Info("Auditing Mock DatabaseService")
	if dbs.opt.err("dbs.Audit") {
		return err{"dbs.Audit"}
	}
	return nil
}

func (dbs *databaseService) Info() {
	msg.Info("Mock DatabaseService")
	msg.Detail("...")
}
