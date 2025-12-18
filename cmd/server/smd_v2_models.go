// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
package main

import (
	"github.com/user/smd2/pkg/resources/component"
)

type ComponentArray struct {
	Components []*component.ComponentSpec `json:"Components"`
}
