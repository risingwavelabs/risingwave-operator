# Storages

## Memory

Currently, memory storage is supported for test usage only. We highly discourage you use the memory storage for other
purposes. For now, you can enable the memory metadata and object storage with the following configs:

```yamlex
#...
spec:
  storages:
    meta:
      memory: true
    object:
      memory: true
#...
```

## etcd (meta)

We recommend using the [etcd](https://etcd.io/) to store the metadata. You can specify the connection information of
the `etcd` you'd like to use like the following:

```yamlex
#...
spec:
  storages:
    meta:
      etcd: 
        endpoint: risingwave-etcd:2388
        secret: etcd-credentials      # optional, empty means no authentication 
#...
```

Check the [docs/manifests/risingwave/risingwave-etcd-minio.yaml](/docs/manifests/risingwave/risingwave-etcd-minio.yaml) for how to
provision a simple RisingWave with an `etcd` instance as the metadata storage.

## MinIO

We support using MinIO as the object storage. Check
the [docs/manifests/risingwave/risingwave-etcd-minio.yaml](/docs/manifests/risingwave/risingwave-etcd-minio.yaml) for details. The
YAML structure is like the following:

```yamlex
#...
spec:
  storages:
    object:
      minio:
        secret: minio-credentials
        endpoint: minio-endpoint:2388
        bucket: hummock001
#...
```

## S3

We support using AWS S3 as the object storage. Follow the steps below and check
the [docs/manifests/risingwave/risingwave-etcd-s3.yaml](/docs/manifests/risingwave/risingwave-etcd-s3.yaml) for details:

First, you need to create a `Secret` with the name `s3-credentials`:

```shell
kubectl create secret generic s3-credentials --from-literal AccessKeyID=${ACCESS_KEY} --from-literal SecretAccessKey=${SECRET_ACCESS_KEY} --from-literal Region=${AWS_REGION}
```

Then, you need to create a `bucket` on the console, e.g., `hummock001`.

Finally, you can specify S3 as the object storage in YAML, like the following:

```yamlex
#...
spec:
  storages:
    object:
      s3:
        secret: s3-credentials
        bucket: hummock001
#...
```

## Azure Blob

We support using Azure blob as the object storage. FOllow the steps below and check the [yaml file](/docs/manifests/risingwave/risingwave-azure.yaml) for details:


1. You need to get several parameter values in your Azure account. 
   1. Account Name & Account Key: You need to create a `storage account` in azure blob. Then [storage accounts] -> find your account -> [security + networking] -> [Access keys] -> get your account name and key.
   2. container name: After creating the `storage account`, you need to create a `Container`, use the container name.
   3. root: The risingwave kernel will store data in this folder. For object storage in Azure blob, you do not need to create this folder in advance.
   4. endpoint: When you upload something into your Azure container, each item will have a URL for the download. The prefix of this download URL is the endpoint you need to use. For publicly accessible links, it should be like: `https://StorageAccountA.blob.core.windows.net`. If you use a private link like this `https://StorageAccountA.privatelink.blob.core.windows.net`, you need to do some additional settings in your Azure and make sure your machine can get access to the Azure blob storage.
2. Change the corresponding values in the [yaml file](/docs/manifests/risingwave/risingwave-azure.yaml) and apply it.

    ```yamlex
    stringData:
    AccountName: your-azure-account-name
    AccountKey: your-azure-account-key
    ```
    and 

    ```yamlex
    object:
        azureBlob: 
        secret:  risingwave-azure-blob-credentials
        container: your-azure-container-name
        root: risingwave
        endpoint: https://your-azure-account-name.blob.core.windows.net
    ```
3. After your success launches the risinwgave, execute some query, and flush the data. You should see some files created in your azure blob container folder.


## HDFS
We support using HDFS S3 as the object storage. Follow the steps below and check
the [/docs/manifests/risingwave/risingwave-etcd-hdfs.yaml](/docs/manifests/risingwave/risingwave-etcd-hdfs.yaml) for details. 
1. You need to launch the HDFS service
2. Specific the storage type like below. 
   1. nameNode: If you follow the yaml file we provide, it should be hadoop-hdfs-master:9000
   2. root: the path that storage the data, you can change it.
    ```yamlex
      storages:
        meta:
          memory: true
        object:
          hdfs:
            nameNode: hadoop-hdfs-master:9000
            root: risingwave

    ```
3. Note: you need to pay attention that we only provide HDFS feature in some specific docker image like `ghcr.io/risingwavelabs/risingwave:git-4e93d7061b66f8f12c4e89e097d950dfd3a5963a`. If you use other image, it will error.
4. After launch, you cannot connect to database can execute SQL. After flush, you can find the file written in the `${root}` fold we assigned in HDFS.
5. If you use the above yaml file, you can forward the port and check the HDFS data in `localhost:50070`.
   ```
   kubectl port-forward service/hadoop-hdfs-master 50070:50070
   ```


# Test

## check instance status
When you launch the RW, you can use `kubectl get risingwave` to check the detail. 
1. `RUNNING` will show whether instance launch correctly.
2. `STORAGE(META)`, `STORAGE(OBJECT)` will show the type of your meta storage type and obejct storage type
```
NAME                    RUNNING   STORAGE(META)   STORAGE(OBJECT)   AGE
risingwave-etcd-minio   True      etcd            MinIO             30s
```
## Connect and execute SQL
1. creating a pod
   ```
   kubectl apply -f https://raw.githubusercontent.com/risingwavelabs/risingwave-operator/main/docs/manifests/psql/psql-console.yaml
   ```
2. Attach the pod
   ```
   kubectl exec -it psql-console -- bash
   ```
3. Connect RW via psql. You need to replace `risingwave-frontend` with your rw frontend name
   ```
   psql -h risingwave-frontend -p 4567 -d dev -U root
   ```
4. Execute a simple read/write test
  ```sql  
    /* create a table */
  create table t1(v1 int);

  /* create a materialized view based on the previous table */
  create materialized view mv1 as select sum(v1) as sum_v1 from t1;

  /* insert some data into the source table */
  insert into t1 values (1), (2), (3);

  /* (optional) ensure the materialized view has been updated */
  flush;

  /* the materialized view should reflect the changes in source table */
  select * from mv1;
  ```

  If everything is correct, the result should be:
  ```
   sum_v1
  --------
        6
  (1 row)
  ```
5. In the end, you can check whether the data has already written into your object storage.