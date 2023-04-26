FROM golang:1.20

RUN apt-get update && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash && \
    apt-get install -y docker.io nodejs && \
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.18.0/kind-linux-amd64 && chmod +x ./kind && mv ./kind /usr/local/bin/kind && \
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x ./kubectl && mv ./kubectl /usr/local/bin/kubectl && \
    npm install -g cspell@latest && \
    wget -qO- https://github.com/koalaman/shellcheck/releases/download/v0.9.0/shellcheck-v0.9.0.linux.x86_64.tar.xz | tar -xJf - && cd shellcheck-v0.9.0 && cp shellcheck /usr/local/bin && cd .. && rm -r shellcheck-v0.9.0  && \
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash


WORKDIR /workspace

ENTRYPOINT ["/bin/bash"]