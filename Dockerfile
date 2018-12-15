FROM golang:alpine

 RUN apk add --update \
 python \
 curl \
 which \
 bash

 RUN curl -sSL https://sdk.cloud.google.com | bash

 RUN gcloud components install pubsub-emulator

 RUN gcloud components update

 RUN gcloud beta emulators pubsub start &
 
 ENV PATH $PATH:/root/google-cloud-sdk/bin

 ADD . $GOPATH/src/PubSubEmulatorApp

 WORKDIR $GOPATH/src/PubSubEmulatorApp

 RUN cd  $GOPATH/src/PubSubEmulatorApp; go build -o pubsubapp

 ENTRYPOINT [ "./pubsubapp" ]
