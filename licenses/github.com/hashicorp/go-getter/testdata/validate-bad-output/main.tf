# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

module "child" {
    source = "./child"
}

resource "aws_instance" "foo" {
    memory = "${module.child.memory}"
}
