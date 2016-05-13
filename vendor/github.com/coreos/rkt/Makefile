# make "all" a default target
all:

# makelib/inc.mk must be included first!
include makelib/inc.mk
include makelib/verbosity.mk
include makelib/file-ops-prolog.mk
include makelib/variables.mk
include makelib/misc.mk

SHELL := $(BASH_SHELL)
TOPLEVEL_STAMPS :=
TOPLEVEL_CHECK_STAMPS :=
TOPLEVEL_UNIT_CHECK_STAMPS :=
TOPLEVEL_FUNCTIONAL_CHECK_STAMPS :=
TOPLEVEL_SUBDIRS := rkt tests stage1 stage1_fly

$(call inc-one,tools/tools.mk)
$(call inc-many,$(foreach sd,$(TOPLEVEL_SUBDIRS),$(sd)/$(sd).mk))

all: $(TOPLEVEL_STAMPS)

$(TOPLEVEL_CHECK_STAMPS): $(TOPLEVEL_STAMPS)

.INTERMEDIATE: $(TOPLEVEL_CHECK_STAMPS)
.INTERMEDIATE: $(TOPLEVEL_UNIT_CHECK_STAMPS)
.INTERMEDIATE: $(TOPLEVEL_FUNCTIONAL_CHECK_STAMPS)

check: $(TOPLEVEL_CHECK_STAMPS)
unit-check: $(TOPLEVEL_UNIT_CHECK_STAMPS)
functional-check: $(TOPLEVEL_FUNCTIONAL_CHECK_STAMPS)

include makelib/file-ops-epilog.mk

.PHONY: all check unit-check functional-check
