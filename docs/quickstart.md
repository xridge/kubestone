title: Kubestone - Quickstart guide

# Quickstart Guide



Welcome to Kubestone, the benchmarking operator for Kubernetes.

This guide will walk you through on the one the installation process and will show you how to create a benchmark run on your cluster.



## Installation

**Requirements**

- [Kubernetes](https://kubernetes.io/) v1.13 (or newer)

- [Kustomize v3.1.0](https://kustomize.io/)

- Cluster admin privileges



Deploy Kubestone to `kubestone-system` namespace with the following command:

```bash
$ kustomize build github.com/xridge/kubestone/config/default?ref= | sed "s/kubestone:latest/kubestone:/" | kubectl create -f -
```

Once deployed, Kubestone will listen for Custom Resources created with the `kubestone.xridge.io` group.


## Benchmarking

Benchmarks can be executed via Kubestone by creating Custom Resources in your cluster.

### Namespace

It is recommended to create a dedicated namespace for benchmarking.

```bash
$ kubectl create namespace kubestone
```

After the namespace is created you can use it to post a benchmark request to the cluster.

The resulting benchmark executions will reside in this namespace.



### Custom Resource rendering

We will be using [kustomize](https://kustomize.io/) to render the Custom Resource from the [github repository](https://github.com/xridge/kubestone/tree/master/config/samples/fio/).

Kustomize takes a [base yaml](https://github.com/xridge/kubestone/blob/master/config/samples/fio/base/fio_cr.yaml) and patches with an [overlay file](https://github.com/xridge/kubestone/blob/master/config/samples/fio/overlays/pvc/patch.yaml) to render the final yaml file, which describes the benchmark.

```bash
$ kustomize build github.com/xridge/kubestone/config/samples/fio/overlays/pvc
```

*Note: Kustomize has [extensive documentation](https://github.com/kubernetes-sigs/kustomize/tree/master/docs) about it's operation. It is advised to have a basic understanding of how Kustomize works as Kubestone uses it's features for both deployment and benchmark execution.*



The Custom Resource (rendered yaml) looks as follows:

```yaml
apiVersion: perf.kubestone.xridge.io/v1alpha1
kind: Fio
metadata:
  name: fio-sample
spec:
  cmdLineArgs: --name=randwrite --iodepth=1 --rw=randwrite --bs=4m --size=256M
  image:
    name: xridge/fio:3.13
  volume:
    persistentVolumeClaimSpec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
    volumeSource:
      persistentVolumeClaim:
        claimName: GENERATED
```



When we create this resource in Kubernetes, the operator interprets it and creates the associated benchmark. The fields of the Custom Resource controls how the benchmark will be executed:

- `metadata.name`: Identifies the Custom Resource. Later, this can be used to query or delete the benchmark in the cluster.
- `cmdLineArgs`: Arguments passed to the benchmark. In this case we are providing the arguments to fio (a filesystem benchmark). It instructs the benchmark to execute a random write test with 4Mb of block size with an overall transfer size of 256 MB.
- `image.name`: Describes the Docker Image of the benchmark. In case of [Fio](https://fio.readthedocs.io/) we are using [xridge's fio Docker Image](https://cloud.docker.com/u/xridge/repository/docker/xridge/fio), which is built from [this repository](https://github.com/xridge/fio-docker/).
- `volume.persistentVolumeClaimSpec`: Given that Fio is a disk benchmark we can set a PersistentVolumeClaim for the benchmark to be executed. The above setup instructs Kubernetes to take 1GB of space from the default StorageClass and use it for the benchmark.



### Running the benchmark

Now, as we understand the definition of the benchmark we can try to execute it.

*Note: Make sure you installed the kubestone operator and have it running before executing this step.*

```bash
$ kustomize build github.com/xridge/kubestone/config/samples/fio/overlays/pvc | kubectl create --namespace kubestone -f -
```

Since we pipe the output of the `kustomize build` command into `kubectl create`, it will create the object in our Kubernetes cluster.



The resulting object can be queried using the object's type (`fio`) and it's name (`fio-sample`):

```bash
$ kubectl describe --namespace kubestone fio fio-sample
Name:         fio-sample
Namespace:    kubestone
Labels:       <none>
Annotations:  <none>
API Version:  perf.kubestone.xridge.io/v1alpha1
Kind:         Fio
Metadata:
  Creation Timestamp:  2019-09-14T11:31:02Z
  Generation:          1
  Resource Version:    31488293
  Self Link:           /apis/perf.kubestone.xridge.io/v1alpha1/namespaces/kubestone/fios/fio-sample
  UID:                 21cdbe92-d6e3-11e9-ba70-4439c4920abc
Spec:
  Cmd Line Args:  --name=randwrite --iodepth=1 --rw=randwrite --bs=4m --size=256M
  Image:
    Name:  xridge/fio:3.13
  Volume:
    Persistent Volume Claim Spec:
      Access Modes:
        ReadWriteOnce
      Resources:
        Requests:
          Storage:  1Gi
    Volume Source:
      Persistent Volume Claim:
        Claim Name:  GENERATED
Status:
  Completed:  true
  Running:    false
Events:
  Type    Reason           Age   From       Message
  ----    ------           ----  ----       -------
  Normal  Created  11s   kubestone  Created /api/v1/namespaces/kubestone/configmaps/fio-sample
  Normal  Created  11s   kubestone  Created /api/v1/namespaces/kubestone/persistentvolumeclaims/fio-sample
  Normal  Created  11s   kubestone  Created /apis/batch/v1/namespaces/kubestone/jobs/fio-sample
```



As the `Events` section shows, Kubestone has created a `ConfigMap`, a `PersistentVolumeClaim` and a` Job` for the provided Custom Resource. The `Status` field tells us that the benchmark has completed.



### Inspecting the benchmark

The created objects related to the benchmark can be listed using `kubectl` command:

```bash
$ kubectl get pods,jobs,configmaps,pvc --namespace kubestone
NAME                   READY   STATUS      RESTARTS   AGE
pod/fio-sample-bqqmm   0/1     Completed   0          54s

NAME                   COMPLETIONS   DURATION   AGE
job.batch/fio-sample   1/1           15s        54s

NAME                   DATA   AGE
configmap/fio-sample   0      54s

NAME                               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
persistentvolumeclaim/fio-sample   Bound    pvc-b3898236-c698-11e9-8071-4439c4920abc   1Gi        RWO            rook-ceph-block   54s
```



As shown above, fio controller has created a PersistentVolumeClaim and a ConfigMap which is used by the Fio Job during benchmark execution. The Fio Job has an associated Pod which contains our test execution. The results of the run can be shown with the `kubectl logs` command:

```bash
$ kubectl logs --namespace kubestone fio-sample-bqqmm
randwrite: (g=0): rw=randwrite, bs=(R) 4096KiB-4096KiB, (W) 4096KiB-4096KiB, (T) 4096KiB-4096KiB, ioengine=psync, iodepth=1
fio-3.13
Starting 1 process
randwrite: Laying out IO file (1 file / 256MiB)

randwrite: (groupid=0, jobs=1): err= 0: pid=47: Sat Aug 24 17:58:10 2019
  write: IOPS=470, BW=1882MiB/s (1974MB/s)(256MiB/136msec); 0 zone resets
    clat (usec): min=1887, max=2595, avg=2042.76, stdev=136.56
     lat (usec): min=1953, max=2688, avg=2107.35, stdev=142.94
    clat percentiles (usec):
     |  1.00th=[ 1893],  5.00th=[ 1926], 10.00th=[ 1926], 20.00th=[ 1958],
     | 30.00th=[ 1991], 40.00th=[ 2008], 50.00th=[ 2024], 60.00th=[ 2040],
     | 70.00th=[ 2057], 80.00th=[ 2073], 90.00th=[ 2114], 95.00th=[ 2409],
     | 99.00th=[ 2606], 99.50th=[ 2606], 99.90th=[ 2606], 99.95th=[ 2606],
     | 99.99th=[ 2606]
  lat (msec)   : 2=34.38%, 4=65.62%
  cpu          : usr=2.22%, sys=97.78%, ctx=1, majf=0, minf=9
  IO depths    : 1=100.0%, 2=0.0%, 4=0.0%, 8=0.0%, 16=0.0%, 32=0.0%, >=64=0.0%
     submit    : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.0%, 64=0.0%, >=64=0.0%
     complete  : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.0%, 64=0.0%, >=64=0.0%
     issued rwts: total=0,64,0,0 short=0,0,0,0 dropped=0,0,0,0
     latency   : target=0, window=0, percentile=100.00%, depth=1

Run status group 0 (all jobs):
  WRITE: bw=1882MiB/s (1974MB/s), 1882MiB/s-1882MiB/s (1974MB/s-1974MB/s), io=256MiB (268MB), run=136-136msec

Disk stats (read/write):
  rbd7: ios=0/0, merge=0/0, ticks=0/0, in_queue=0, util=0.00%
```



### Listing benchmarks

We have learned that Kubestone uses Custom Resources to define benchmarks. We can list the installed custom resources using the `kubectl get crds` command:

```bash
$ kubectl get crds | grep kubestone
drills.perf.kubestone.xridge.io         2019-09-08T05:51:26Z
fios.perf.kubestone.xridge.io           2019-09-08T05:51:26Z
iopings.perf.kubestone.xridge.io        2019-09-08T05:51:26Z
iperf3s.perf.kubestone.xridge.io        2019-09-08T05:51:26Z
pgbenches.perf.kubestone.xridge.io      2019-09-08T05:51:26Z
sysbenches.perf.kubestone.xridge.io     2019-09-08T05:51:26Z
```

Using the CRD names above, we can list the executed benchmarks in the system.

Kubernetes provides a convenience feature regarding CRDs: one can use the shortened name of the CRD, which is the singular part of the fully qualified CRD name. In our case, `fios.perf.kubestone.xridge.io` can be shortened to `fio`. Hence, we can list the executed fio benchmark using the following command:

```bash
$ kubectl get --namespace kubestone fios.perf.kubestone.xridge.io
NAME         RUNNING   COMPLETED
fio-sample   false     true
```



### Cleaning up

After a successful benchmark run the resulting objects are stored in the Kubernetes cluster.
Given that Kubernetes can hold a limited number of pods in the system it is advised that the user cleans up the benchmark runs time to time. This can be achieved by deleting the Custom Resource, which initiated the benchmark:

```bash
$ kubectl delete --namespace kubestone fio fio-sample
```

Since the Custom Resource has ownership on the created resources, the underlying pods, jobs, configmaps, pvcs, etc. are also removed by this operation.



## Next steps

Now you are familiar with the key concepts of Kubestone, it is time to explore and benchmark.

You can play around with Fio Benchmark via it's `cmdLineArgs`, Persistent Volume and Scheduling related settings. You can find more information about that in Fio's benchmark page.

If you are interested in other benchmarks, please refer to our [benchmark suite](benchmarks-index.md).



## Tips

In the commands executed above, we needed to specify the namespace.

This step can be avoided if the namespace is specified in your kubernetes config file. One convenient tool to do so is `kubens`, which is part of the [kubectx program suite](https://kubectx.dev).
