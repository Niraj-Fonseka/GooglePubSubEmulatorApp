FROM golang:alpine

 RUN apk add --update \
 python \
 curl \
 which \
 bash

 RUN curl -sSL https://sdk.cloud.google.com | bash

 ENV PATH $PATH:/root/google-cloud-sdk/bin


 RUN echo $PATH