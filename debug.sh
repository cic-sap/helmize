#!/usr/bin/env bash
set -x
mkdir -p bin
go build   -o bin/helmize .

helm plugin uninstall helmize
helm plugin install   .

function clean_helm() {
  helm delete -n default demo3
  helm delete -n default error-result
}

clean_helm

helm helmize upgrade demo3 pkg/testdata/charts -n default
echo must be successful
helm upgrade -i demo3 pkg/testdata/charts -n default
echo must be fail
helm upgrade -i error-result  pkg/testdata/charts -n default
echo must be successful
helm helmize upgrade error-result   pkg/testdata/charts -n default
helm upgrade -i error-result   pkg/testdata/charts -n default

clean_helm