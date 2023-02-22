#!/usr/bin/env bash

CERT_MANAGER_VERSION=${CERT_MANAGER_VERSION:-v1.11.0}
RISINGWAVE_OPERATOR_VERSION=${RISINGWAVE_OPERATOR_VERSION:-latest}

function link_to_risingwave_operator_manifests() {
  local version=$1

  if [[ "${version}" == "latest" ]]; then
    echo "https://github.com/risingwavelabs/risingwave-operator/releases/latest/download/risingwave-operator.yaml"
  else
    echo "https://github.com/risingwavelabs/risingwave-operator/releases/download/${RISINGWAVE_OPERATOR_VERSION}/risingwave-operator.yaml"
  fi
}

function k8s::cert_manager::install() {
  echo "Installing the cert-manager ${CERT_MANAGER_VERSION}..."
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/"${CERT_MANAGER_VERSION}"/cert-manager.yaml >/dev/null
}

function k8s::cert_manager::_wait_validating_webhook_ca_bundle() {
  local certificate
  local threshold=60
  local current_epoch=0
  local interval=5
  while :; do
    certificate=$(kubectl get validatingwebhookconfigurations cert-manager-webhook -o jsonpath='{.webhooks[0].clientConfig.caBundle}')
    if [ -n "$certificate" ]; then
      break
    fi
    if ((current_epoch == threshold)); then
      echo >&2 "Timeout waiting for cert-manager's CA bundle to be ready!"
      return 1
    fi

    ((current_epoch++))
    sleep "${interval}"
  done
}

function k8s::cert_manager::wait() {
  k8s::cert_manager::_wait_validating_webhook_ca_bundle
}

function k8s::risingwave_operator::install() {
  local manifest_link
  manifest_link=$(link_to_risingwave_operator_manifests "${RISINGWAVE_OPERATOR_VERSION}")

  echo "Installing the risingwave-operator ${RISINGWAVE_OPERATOR_VERSION}..."
  kubectl apply -f "${manifest_link}" >/dev/null
}

k8s::cert_manager::install && k8s::cert_manager::wait && k8s::risingwave_operator::install
echo "Done!"
