# kube-advisor

kube-advisor is a diagnostic tool for Kubernetes clusters. At the moment, it returns pods that are missing resource and request limits.

![screenshot](https://user-images.githubusercontent.com/1231630/44505638-5a2c0500-a657-11e8-8bf1-1766d69fa2ff.png)

## Running in a Kubernetes cluster without RBAC enabled

Just run the pod by itself:

```bash
# kubectl run --rm -i -t kube-advisor --image=mcr.microsoft.com/aks/kube-advisor --restart=Never
```

## Running in a Kubernetes cluster with RBAC enabled

### Create the service account and cluster role binding

```bash
# kubectl apply -f https://raw.githubusercontent.com/Azure/kube-advisor/master/sa.yaml?token=ABLLDqUpCcBLHrAoMNOCwSahn4b-hwKKks5bl-0QwA%3D%3D
```

### Run the pod

```bash
# kubectl run --rm -i -t kube-advisor --image=mcr.microsoft.com/aks/kube-advisor --restart=Never --overrides="{ \"apiVersion\": \"v1\", \"spec\": { \"serviceAccountName\": \"kube-advisor\" } }"
```

### If desired, delete the service account and cluster role binding

```bash
# kubectl delete -f https://raw.githubusercontent.com/Azure/kube-advisor/master/sa.yaml?token=ABLLDqUpCcBLHrAoMNOCwSahn4b-hwKKks5bl-0QwA%3D%3D
```
