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

package resource

import "github.com/cisco/arc/pkg/route"

// Resources provides a collection of resource.Resource objects while
// implementing the Resource interface. It is meant to be used as an
// embedded field in compostive resource type.
type Resources struct {
	resources []Resource
}

// NewResources constructs a Resources object.
func NewResources() *Resources {
	return &Resources{
		resources: []Resource{},
	}
}

// Append adds a resource to the collection.
func (r *Resources) Append(rsrc Resource) {
	if r == nil {
		return
	}
	r.resources = append(r.resources, rsrc)
}

// Length provides the number of resources in the collection.
func (r *Resources) Length() int {
	if r == nil {
		return 0
	}
	return len(r.resources)
}

// Get returns the list of Resources in the collection.
func (r *Resources) Get() []Resource {
	if r == nil {
		return nil
	}
	return r.resources
}

// Created indicates true if all resources in the collection have been created.
// It is possible for the collection to fail both Created and Destroyed tests if
// for some reason the collection of resources has been partially created.
func (r *Resources) Created() bool {
	if r == nil || len(r.resources) == 0 {
		return false
	}
	for _, rsrc := range r.resources {
		if rsrc == nil || !rsrc.Created() {
			return false
		}
	}
	return true
}

// Destroyed indicates true if all resources in the collection have been destroyed.
// It is possible for the collection to fail both Created and Destroyed tests if
// for some reason the collection of resources has been partially created.
func (r *Resources) Destroyed() bool {
	if r == nil {
		return true
	}
	for _, rsrc := range r.resources {
		if rsrc != nil && !rsrc.Destroyed() {
			return false
		}
	}
	return true
}

// RouteInOrder routes requests to each resouce in the collection in the
// order they were added to the collection. Most requests should use
// this method to route requests.
func (r *Resources) RouteInOrder(req *route.Request) route.Response {
	if r == nil {
		return route.FAIL
	}
	for _, rsrc := range r.resources {
		if rsrc == nil {
			continue
		}
		if resp := rsrc.Route(req); resp != route.OK {
			return resp
		}
	}
	return route.OK
}

// RouteReverseOrder routes requests to each resouce in the collection in the
// reverse order they were added to the collection. This is useful for the
// destroy request where the resources need to be destroyed in the reverse
// order that they where created.
func (r *Resources) RouteReverseOrder(req *route.Request) route.Response {
	if r == nil {
		return route.FAIL
	}
	for i := len(r.resources) - 1; i >= 0; i-- {
		rsrc := r.resources[i]
		if rsrc == nil {
			continue
		}
		if resp := rsrc.Route(req); resp != route.OK {
			return resp
		}
	}
	return route.OK
}
