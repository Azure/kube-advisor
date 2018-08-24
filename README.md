# kube-advisor

kube-advisor is a diagnostic tool for Kubernetes clusters. At the moment, it returns pods that are missing resource and request limits.

![screenshot](https://user-images.githubusercontent.com/1231630/44505638-5a2c0500-a657-11e8-8bf1-1766d69fa2ff.png)

## Running in a Kubernetes cluster without RBAC enabled

Just run the pod by itself:

```bash
# kubectl run --rm -i -t kube-advisor --image=mcr.microsoft.com/kube-advisor --restart=Never
```

## Running in a Kubernetes cluster with RBAC enabled

### Create the service account and cluster role binding

```bash
# kubectl apply -f https://raw.githubusercontent.com/Azure/kube-advisor/master/sa.yaml?token=ABLLDrNcuHMro9jQ0xduCaEbpzLupzQUks5bh3RhwA%3D%3D
```

### Run the pod

```bash
# kubectl run --rm -i -t kube-advisor --image=mcr.microsoft.com/kube-advisor --restart=Never --overrides="{ \"apiVersion\": \"v1\", \"spec\": { \"serviceAccountName\": \"kube-advisor\" } }"
```

### If desired, delete the service account and cluster role binding

```bash
# kubectl delete -f https://raw.githubusercontent.com/Azure/kube-advisor/master/sa.yaml?token=ABLLDrNcuHMro9jQ0xduCaEbpzLupzQUks5bh3RhwA%3D%3D
```


# Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
