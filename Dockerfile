From ubuntu:18.04

COPY ./bin/fetch-k8s-resource /usr/local/bin/fetch-k8s-resource
ENTRYPOINT ["fetch-k8s-resource"]
