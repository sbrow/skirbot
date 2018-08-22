#!/bin/bash
pkg=github.com/sbrow/skirmish 

godoc2md -template .doc.template $pkg > README.md
branch=$(git rev-parse --abbrev-ref HEAD)
sed -i -r "s/branch=master/branch=$branch/g" README.md