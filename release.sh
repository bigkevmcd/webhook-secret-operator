#!/bin/sh
for file in deploy/*.yaml; do
  echo "---" >> release.yaml
  cat $file >> release.yaml
done
echo "---" >> release.yaml
cat deploy/crds/apps.bigkevmcd.com_webhooksecrets_crd.yaml >> release.yaml
