# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Duplicate resources
resource "aws_instance" "foo" {}
resource "aws_instance" "foo" {}
