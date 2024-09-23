FROM golang:1.23

ARG KIND_VERSION=v0.18.0
ARG SHELLCHECK_VERSION=v0.10.0

RUN apt-get update && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash && \
    apt-get install -y nodejs ca-certificates gnupg xz-utils && \
    npm install -g cspell@latest && \
    install -m 0755 -d /etc/apt/keyrings && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && chmod a+r /etc/apt/keyrings/docker.gpg && \
    echo "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt-get update && apt-get install -y docker-ce-cli && \
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64 && chmod +x ./kind && mv ./kind /usr/local/bin/kind && \
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x ./kubectl && mv ./kubectl /usr/local/bin/kubectl && \
    wget -qO- https://github.com/koalaman/shellcheck/releases/download/${SHELLCHECK_VERSION}/shellcheck-${SHELLCHECK_VERSION}.linux.x86_64.tar.xz | tar -xJf - && cd shellcheck-${SHELLCHECK_VERSION} && cp shellcheck /usr/local/bin && cd .. && rm -r shellcheck-${SHELLCHECK_VERSION}  && \
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash && \
    curl -Os https://uploader.codecov.io/latest/linux/codecov && chmod +x codecov && mv codecov /usr/local/bin/codecov

RUN apt-get install -y gettext-base jq etcd-client

WORKDIR /workspace

ENTRYPOINT ["/bin/bash"]
