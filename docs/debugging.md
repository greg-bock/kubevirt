# Debugging

When debugging KubeVirt in the development environment, you have to be aware
that we have a proxy for ThirdPartyResource related preprocessing in front of
the apiserver. By default the `cluster/kubectl.sh` script assumes that you want
to talk to the cluster through that proxy. So if

```bash
cluster/kubectl.sh version
```

is run, `cluster/kubectl.sh` connects to the proxy. However when something is
not right with the proxy, or KubeVirt is not even yet deployed, it is possible
to connect to the apiserver directly by adding `--core` as the first argument.
So

```bash
cluster/kubectl.sh --core version
```

will try to connect to the apiserver.

## Retrieving Logs

To investigate the logs of a container, you can view the logs via
`cluster/kubectl.sh logs`.  To view the logs of `virt-api`, type

```bash
cluster/kubectl.sh logs virt-api -f
```

Sometimes a container is crashlooping because of an application error inside
it. In this case, you normally can't see any logs, because the container is
already gone, and so are the logs. To get the logs from the last run attempt,
the `--previous` flag can be used. To view the logs of `virt-api` from the
previous run, type

```bash
cluster/kubectl.sh logs virt-api -f -p
```

## Watching Events

Both, Kubernetes and KubeVirt are creating events, which can be viewed via

```bash
cluster/kubectl.sh get events --watch
```

This way it is pretty easy to detect if a Pod or a VM got started.

## Entering Containers

It can be very valuable to enter a container and do some investigations there,
to see what is going wrong. In this case the kubectl `exec` command can be
used. To enter `virt-api` with an interactive shell, type

```bash
cluster/kubectl.sh exec virt-api -c virt-api -i -t -- sh
```

## Kubelet Logs

After all you might not see errors in the logs provided by Kubernetes. In that case
you can take a look at the logs of the `kubelet` on the host where the issue is
appearing. Depending on the error it is getting logged to either the system logs or
to the kubelet logs, you can use the following commands to view them:

```bash
journalctl
# or
journalctl -u kubelet
```

## References

 - [kubectl overview](https://kubernetes.io/docs/user-guide/kubectl-overview/)
 - [kubectl reference](https://kubernetes.io/docs/user-guide/kubectl/v1.5/)

# Using a Debugger (delve)

This shows the basic principle on how remote debugging can be done.

 - Add delve to the container
 - Start delve on a specific port ( `dlv attach <pid> --headless --listen=0.0.0.0:1234`)
 - Use kube-proxy to forward the port to your machine