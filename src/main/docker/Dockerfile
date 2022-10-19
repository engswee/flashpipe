FROM bitnami/java:1.8.292-prod-debian-10-r23
# ----------------------------------------
# Docker base image reference:
# https://hub.docker.com/r/bitnami/java
# ----------------------------------------

# ----------------------------------------
# 1 - Install zip functionality
# ----------------------------------------
RUN install_packages git unzip

# ----------------------------------------
# 2 - Install Maven
# Dockerfile reference:
# https://github.com/carlossg/docker-maven/blob/master/openjdk-8/Dockerfile
# ----------------------------------------
ARG MAVEN_VERSION=3.8.6
ARG USER_HOME_DIR="/root"
ARG SHA=f790857f3b1f90ae8d16281f902c689e4f136ebe584aba45e4b1fa66c80cba826d3e0e52fdd04ed44b4c66f6d3fe3584a057c26dfcac544a60b301e6d0f91c26
ARG BASE_URL=https://apache.osuosl.org/maven/maven-3/${MAVEN_VERSION}/binaries

RUN mkdir -p /usr/share/maven /usr/share/maven/ref \
  && curl -fsSL -o /tmp/apache-maven.tar.gz ${BASE_URL}/apache-maven-${MAVEN_VERSION}-bin.tar.gz \
  && echo "${SHA}  /tmp/apache-maven.tar.gz" | sha512sum -c - \
  && tar -xzf /tmp/apache-maven.tar.gz -C /usr/share/maven --strip-components=1 \
  && rm -f /tmp/apache-maven.tar.gz \
  && ln -s /usr/share/maven/bin/mvn /usr/bin/mvn

ENV MAVEN_HOME /usr/share/maven
ENV MAVEN_CONFIG "$USER_HOME_DIR/.m2"

COPY src/main/docker/settings-docker.xml /usr/share/maven/ref/

# ----------------------------------------
# 3 - Copy Maven POM and download all dependencies & plugin for faster execution
# ----------------------------------------
COPY pom.xml /tmp/pom.xml
RUN mvn -B -f /tmp/pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:go-offline \
  && chmod -R 777 /usr/share/maven/ref/repository
RUN mvn -B -f /tmp/pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:get -Dartifact=io.github.engswee:cpi-mock-message:1.0.0

# ----------------------------------------
# 4 - Copy JAR file used for accessing CPI APIs
# ----------------------------------------
# <FLASHPIPE_VERSION>
ARG FLASHPIPE_VERSION=2.6.1-SNAPSHOT
COPY target/flashpipe-${FLASHPIPE_VERSION}.jar /tmp/flashpipe-${FLASHPIPE_VERSION}.jar
RUN mvn -s /usr/share/maven/ref/settings-docker.xml install:install-file -DgroupId=io.github.engswee -DartifactId=flashpipe -Dversion=${FLASHPIPE_VERSION} -Dpackaging=jar -Dfile=/tmp/flashpipe-${FLASHPIPE_VERSION}.jar \
  && chmod -R 777 /usr/share/maven/ref/repository/io/github

COPY src/main/docker/script/*.sh /usr/bin/
RUN chmod +x /usr/bin/*.sh

RUN mkdir -p /tmp/log4j2-config
COPY src/main/docker/log4j2-config/*.xml /tmp/log4j2-config/