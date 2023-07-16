// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import "github.com/ok-ryoko/turret/pkg/linux"

type OpenSUSETurretBuilder struct {
	TurretBuilder
}

func (b *OpenSUSETurretBuilder) Distro() linux.Distro {
	return linux.OpenSUSE
}
