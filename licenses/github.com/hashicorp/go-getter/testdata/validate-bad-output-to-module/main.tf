# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

module "child" {
    source = "./child"
}

module "child2" {
    source = "./child"
    memory = "${module.child.memory_max}"
}
