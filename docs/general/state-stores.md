# State Store Backends

RisingWave supports various state store backends including the AWS S3, so does the RisingWave operator. You can
customize the state store backend of RisingWave cluster by modifying the manifest YAML file. For more API details,
please refer to the API reference in the [api.md](api.md) file.

The state store backends are defined in the `spec.stateStore` section of the RisingWave manifest YAML file. Currently,
the following state store backends are supported:

- [Memory (for test only)](#memory-for-test-only)
- [Local File System (for test only)](#local-file-system-for-test-only)
- [MinIO](#minio)
- [AWS S3](#aws-s3)
- [S3 compatible object storages](#s3-compatible-object-storages)
- [Google Cloud Storage](#google-cloud-storage)
- [Azure Blob Storage](#azure-blob-storage)
- [Apache HDFS / WebHDFS](#apache-hdfs--webhdfs)

## Memory (for test only)

```yamlex
spec:
  stateStore:
    memory: true
```

## Local File System (for test only)

```yamlex
spec:
  stateStore:
    localDisk:
      # Path to the directory of the local file system.
      path: /tmp/risingwave
```

## MinIO

```yamlex
spec:
  stateStore:
    # Prefix to objects in the object stores or directory in file system. Default to "hummock".
    dataDirectory: hummock
    
    # Declaration of the MinIO state store backend.
    minio:
      # Endpoint of the MinIO service.
      endpoint: risingwave-minio:9301
      
      # Name of the MinIO bucket.
      bucket: hummock001
      
      # Credentials to access the MinIO bucket.
      credentials:
        # Name of the Kubernetes secret that stores the credentials.
        secretName: minio-credentials
        
        # Key of the username ID in the secret.
        usernameKeyRef: username
        
        # Key of the password key in the secret.
        passwordKeyRef: password 
```

## AWS S3

```yamlex
spec:
  stateStore:
    # Prefix to objects in the object stores or directory in file system. Default to "hummock".
    dataDirectory: hummock
    
    # Declaration of the S3 state store backend.
    s3:
      # Region of the S3 bucket.
      region: us-east-1
      
      # Name of the S3 bucket.
      bucket: risingwave
      
      # Credentials to access the S3 bucket.
      credentials:
        # Name of the Kubernetes secret that stores the credentials.
        secretName: s3-credentials
        
        # Key of the access key ID in the secret.
        accessKeyRef: AWS_ACCESS_KEY_ID
        
        # Key of the secret access key in the secret.
        secretAccessKeyRef: AWS_SECRET_ACCESS_KEY
        
        # Optional, set it to true when the credentials can be retrieved 
        # with the service account token, e.g., running inside the EKS.
        # 
        # useServiceAccount: true 
```

## S3 compatible object storages

RisingWave also supports S3 compatible object storages, such as the Tencent Cloud Object Storage (COS), Aliyun Object
Storage Service (OSS), MinIO, etc. The configuration is similar to the AWS S3 backend, except that the endpoint is
different.

```yamlex
spec:
  stateStore:
    # Prefix to objects in the object stores or directory in file system. Default to "hummock".
    dataDirectory: hummock
    
    # Declaration of the S3 compatible state store backend.
    s3:
      # Endpoint of the S3 compatible object storage. Two variables are supported:
      # - ${BUCKET}: name of the S3 bucket.
      # - ${REGION}: name of the region.
      endpoint: ${BUCKET}.cos.${REGION}.myqcloud.com
      
      # Region of the S3 compatible bucket.
      region: ap-guangzhou
      
      # Name of the S3 compatible bucket.
      bucket: risingwave
      
      # Credentials to access the S3 compatible bucket.
      credentials:
        # Name of the Kubernetes secret that stores the credentials.
        secretName: cos-credentials
        
        # Key of the access key ID in the secret.
        accessKeyRef: ACCESS_KEY_ID
        
        # Key of the secret access key in the secret.
        secretAccessKeyRef: SECRET_ACCESS_KEY
```

## Google Cloud Storage

```yamlex
spec:
  stateStore:
    # Prefix to objects in the object stores or directory in file system. Default to "hummock".
    dataDirectory: hummock
    
    # Declaration of the Google Cloud Storage state store backend.
    gcs:
      # Name of the Google Cloud Storage bucket.
      bucket: risingwave
      
      # Root directory of the Google Cloud Storage bucket.
      root: risingwave
    
      # Credentials to access the Google Cloud Storage bucket.
      credentials:
        # Name of the Kubernetes secret that stores the credentials.
        secretName: gcs-credentials
        
        # Key of the service account credentials in the secret.
        serviceAccountCredentialsKeyRef: ServiceAccountCredentials
        
        # Optional, set it to true when the credentials can be retrieved.
        # useWorkloadIdentity: true
```

## Azure Blob Storage

```yamlex
spec:
  stateStore:
    # Prefix to objects in the object stores or directory in file system. Default to "hummock".
    dataDirectory: hummock
    
    # Declaration of the Azure Blob Storage state store backend.
    azureBlob:
      # Endpoint of the Azure Blob service.
      endpoint: https://you-blob-service.blob.core.windows.net
      
      # Working directory root of the Azure Blob service.
      root: risingwave
      
      # Container name of the Azure Blob service.
      container: risingwave
    
      # Credentials to access the Azure Blob Storage container.
      credentials:
        # Name of the Kubernetes secret that stores the credentials.
        secretName: azure-credentials
        
        # Key of the account name in the secret.
        accountNameRef: AccountName
        
        # Key of the account name in the secret.
        accountKeyRef: AccountKey
```

## Apache HDFS / WebHDFS

(Note: The standard image do not support HDFS. Please get in touch with us to get the latest available image tag)

```yamlex
spec:
  stateStore:
    # Prefix to objects in the object stores or directory in file system. Default to "hummock".
    dataDirectory: hummock
    
    # Declaration of the Apache HDFS state store backend.
    hdfs:
      # Endpoint of the Apache HDFS.
      nameNode: hadoop-hdfs-master:9000
      
      # Working directory root of the Apache HDFS service.
      root: risingwave
```

```yamlex
spec:
  stateStore:
    # Prefix to objects in the object stores or directory in file system. Default to "hummock".
    dataDirectory: hummock
    
    # Declaration of the Apache WebHDFS state store backend.
    webhdfs:
      # Endpoint of the Apache HDFS.
      nameNode: hadoop-hdfs-master:9000
      
      # Working directory root of the Apache HDFS service.
      root: risingwave
```
