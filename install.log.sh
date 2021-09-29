set -ex
pwd
env
bash $HELM_PLUGIN_DIR/install-binary.sh &> /tmp/helm.log
