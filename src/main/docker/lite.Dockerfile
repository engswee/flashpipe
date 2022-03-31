FROM bitnami/java:1.8.292-prod-debian-10-r23
# ----------------------------------------
# Docker base image reference:
# https://hub.docker.com/r/bitnami/java
# ----------------------------------------

# ----------------------------------------
# 1 - Install unzip functionality
# ----------------------------------------
RUN install_packages git unzip

# ----------------------------------------
# 2 - Copy third-party JAR file used for accessing CPI APIs
# ----------------------------------------
ARG MAVEN_REPO_DIR="/usr/share/maven/ref/repository"
ARG BASE_URL=https://repo1.maven.org/maven2

ARG GROOVY_VER=2.4.21
RUN mkdir -p       ${MAVEN_REPO_DIR}/org/codehaus/groovy/groovy-all/${GROOVY_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/org/codehaus/groovy/groovy-all/${GROOVY_VER}/groovy-all-${GROOVY_VER}.jar \
                         ${BASE_URL}/org/codehaus/groovy/groovy-all/${GROOVY_VER}/groovy-all-${GROOVY_VER}.jar
ARG APACHE_HTTP_VER=5.0.4
RUN mkdir -p       ${MAVEN_REPO_DIR}/org/apache/httpcomponents/core5/httpcore5/${APACHE_HTTP_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/org/apache/httpcomponents/core5/httpcore5/${APACHE_HTTP_VER}/httpcore5-${APACHE_HTTP_VER}.jar \
                         ${BASE_URL}/org/apache/httpcomponents/core5/httpcore5/${APACHE_HTTP_VER}/httpcore5-${APACHE_HTTP_VER}.jar
RUN mkdir -p       ${MAVEN_REPO_DIR}/org/apache/httpcomponents/client5/httpclient5/${APACHE_HTTP_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/org/apache/httpcomponents/client5/httpclient5/${APACHE_HTTP_VER}/httpclient5-${APACHE_HTTP_VER}.jar \
                         ${BASE_URL}/org/apache/httpcomponents/client5/httpclient5/${APACHE_HTTP_VER}/httpclient5-${APACHE_HTTP_VER}.jar
ARG COMMONS_CODEC_VER=1.15
RUN mkdir -p       ${MAVEN_REPO_DIR}/commons-codec/commons-codec/${COMMONS_CODEC_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/commons-codec/commons-codec/${COMMONS_CODEC_VER}/commons-codec-${COMMONS_CODEC_VER}.jar \
                         ${BASE_URL}/commons-codec/commons-codec/${COMMONS_CODEC_VER}/commons-codec-${COMMONS_CODEC_VER}.jar
ARG SLF4J_VER=1.7.25
RUN mkdir -p       ${MAVEN_REPO_DIR}/org/slf4j/slf4j-api/${SLF4J_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/org/slf4j/slf4j-api/${SLF4J_VER}/slf4j-api-${SLF4J_VER}.jar \
                         ${BASE_URL}/org/slf4j/slf4j-api/${SLF4J_VER}/slf4j-api-${SLF4J_VER}.jar
ARG LOG4J_VER=2.17.1
RUN mkdir -p       ${MAVEN_REPO_DIR}/org/apache/logging/log4j/log4j-slf4j-impl/${LOG4J_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/org/apache/logging/log4j/log4j-slf4j-impl/${LOG4J_VER}/log4j-slf4j-impl-${LOG4J_VER}.jar \
                         ${BASE_URL}/org/apache/logging/log4j/log4j-slf4j-impl/${LOG4J_VER}/log4j-slf4j-impl-${LOG4J_VER}.jar
RUN mkdir -p       ${MAVEN_REPO_DIR}/org/apache/logging/log4j/log4j-api/${LOG4J_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/org/apache/logging/log4j/log4j-api/${LOG4J_VER}/log4j-api-${LOG4J_VER}.jar \
                         ${BASE_URL}/org/apache/logging/log4j/log4j-api/${LOG4J_VER}/log4j-api-${LOG4J_VER}.jar
RUN mkdir -p       ${MAVEN_REPO_DIR}/org/apache/logging/log4j/log4j-core/${LOG4J_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/org/apache/logging/log4j/log4j-core/${LOG4J_VER}/log4j-core-${LOG4J_VER}.jar \
                         ${BASE_URL}/org/apache/logging/log4j/log4j-core/${LOG4J_VER}/log4j-core-${LOG4J_VER}.jar
ARG ZT_ZIP_VER=1.14
RUN mkdir -p       ${MAVEN_REPO_DIR}/org/zeroturnaround/zt-zip/${ZT_ZIP_VER} \
  && curl -fsSL -o ${MAVEN_REPO_DIR}/org/zeroturnaround/zt-zip/${ZT_ZIP_VER}/zt-zip-${ZT_ZIP_VER}.jar \
                         ${BASE_URL}/org/zeroturnaround/zt-zip/${ZT_ZIP_VER}/zt-zip-${ZT_ZIP_VER}.jar

# ----------------------------------------
# 3 - Copy FlashPipe JAR file used for accessing CPI APIs
# ----------------------------------------
ARG FLASHPIPE_VERSION=2.6.0
RUN mkdir -p ${MAVEN_REPO_DIR}/io/github/engswee/flashpipe/${FLASHPIPE_VERSION}
COPY target/flashpipe-${FLASHPIPE_VERSION}.jar ${MAVEN_REPO_DIR}/io/github/engswee/flashpipe/${FLASHPIPE_VERSION}/flashpipe-${FLASHPIPE_VERSION}.jar

COPY src/main/docker/script/*.sh /usr/bin/
RUN chmod +x /usr/bin/*.sh

RUN mkdir -p /tmp/log4j2-config
COPY src/main/docker/log4j2-config/*.xml /tmp/log4j2-config/