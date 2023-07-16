// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import "github.com/ok-ryoko/turret/pkg/linux"

type FedoraTurretBuilder struct {
	TurretBuilder
}

func (b *FedoraTurretBuilder) Distro() linux.Distro {
	return linux.Fedora
}
