Make sure kubelet is started with 

```
--enable-controller-attach-detach=false

```

place hugepages binary in 

```
/usr/libexec/kubernetes/kubelet-plugins/volume/exec/intel~hugepages
```

Example Pod manifest:

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: flex-test
  labels:
    app: example
spec:
  containers:
  - name: flex-test
    image: ubuntu
    command:
    - sleep
    - inf
    volumeMounts:
    - name: test
      mountPath: /test
  volumes:
  - name: test
    flexVolume:
      driver: "intel/hugepages"
```
