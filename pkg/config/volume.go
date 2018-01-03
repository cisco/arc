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

package config

import "github.com/cisco/arc/pkg/msg"

type Volumes []*Volume

func (v *Volumes) Print() {
	msg.Info("Volumes Config")
	msg.IndentInc()
	for _, volume := range *v {
		volume.Print()
	}
	msg.IndentDec()
}

type Volume struct {
	Device_     string `json:"device"`
	Type_       string `json:"type"`
	Size_       int64  `json:"size"`
	Keep_       bool   `json:"keep"`
	Boot_       bool   `json:"boot"`
	FsType_     string `json:"fstype"`
	Inodes_     int    `json:"inodes"`
	MountPoint_ string `json:"mount_point"`
	Preserve_   bool   `json:"preserve"`
}

func (v *Volume) Device() string {
	return v.Device_
}

func (v *Volume) Size() int64 {
	return v.Size_
}

func (v *Volume) Type() string {
	return v.Type_
}

func (v *Volume) Keep() bool {
	return v.Keep_
}

func (v *Volume) Boot() bool {
	return v.Boot_
}

func (v *Volume) FsType() string {
	return v.FsType_
}

func (v *Volume) Inodes() int {
	return v.Inodes_
}

func (v *Volume) MountPoint() string {
	return v.MountPoint_
}

func (v *Volume) Preserve() bool {
	return v.Preserve_
}

func (v *Volume) Print() {
	msg.Info("Volume Config")
	msg.Detail("%-20s\t%s", "device", v.Device())
	msg.Detail("%-20s\t%s", "type", v.Type())
	msg.Detail("%-20s\t%d", "size", v.Size())
	msg.Detail("%-20s\t%t", "keep", v.Keep())
	msg.Detail("%-20s\t%t", "boot", v.Boot())
	msg.Detail("%-20s\t%t", "preserve", v.Preserve())
	if v.FsType() != "" {
		msg.Detail("%-20s\t%s", "fstype", v.FsType())
		msg.Detail("%-20s\t%s", "mount point", v.MountPoint())
	}
}
