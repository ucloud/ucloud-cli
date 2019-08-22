FROM ubuntu:18.04
RUN apt-get update && apt-get install wget -y
RUN wget https://github.com/ucloud/ucloud-cli/releases/download/0.1.22/ucloud-cli-linux-0.1.22-amd64.tgz
RUN tar -zxf ucloud-cli-linux-0.1.22-amd64.tgz -C /usr/local/bin/
RUN echo "complete -C $(which ucloud) ucloud" >> ~/.bashrc
