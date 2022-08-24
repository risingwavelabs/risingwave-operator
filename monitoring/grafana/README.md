## Initialize Project
```shell
cd ./monitoring/grafana/jsonnet
make vendor
```


## Run TPCH Bench Kube
Clone [https://github.com/singularity-data/kube-bench](https://github.com/singularity-data/kube-bench) 
Run the benching script and port-forward grafana to localhost:3000
```shell
./start.sh 
kubectl port-forward svc/prometheus-stack-grafana 3000:http-web
```

## Update  Panal
Make your changes in ./monitoring/grafana/jsonnet/panels
Register changes by running update command
```shell
cd ./monitoring/grafana
make update
```