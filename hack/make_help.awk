#!/usr/bin/awk -f
#
# Copyright 2023 OK Ryoko
# SPDX-License-Identifier: Apache-2.0
#
# make_help - Generate help text for a makefile
#
# USAGE: awk --exec make_help.awk MAKEFILE
#
# Read a Makefile, searching for all lines with this form:
#
#     .PHONY: target #? documentation comment
#
# For each match, print the target and the contents of the comment

BEGIN {
    printf "%s\n\n", "Available targets:"
}

$1 == ".PHONY:" && NF > 3 && $3 == "#?" {
    printf "\t%-10s", $2;
    for (i = 4; i < NF; i++) { printf "%s ", $i }
    printf "%s\n", $NF
}

END {
    printf "\n"
}
