package io.github.engswee.flashpipe.http

import org.slf4j.Logger
import org.slf4j.LoggerFactory

abstract class HTTPExecuter {

    static Logger logger = LoggerFactory.getLogger(HTTPExecuter)

    abstract void setBaseURL(String scheme, String host, int port)

    abstract void setBasicAuth(String user, String password)

    abstract void setOAuthToken(String token)

    abstract void executeRequest(String method, String path, Map headers, Map queryParameters, byte[] requestBytes, String mimeType)

    void executeRequest(String method, String path, Map headers, Map queryParameters, String requestBody, String charset, String mimeType) {
        executeRequest(method, path, headers, queryParameters, requestBody.getBytes(charset), mimeType)
    }

    void executeRequest(String method, String path, Map headers, Map queryParameters) {
        executeRequest(method, path, headers, queryParameters, null, null)
    }

    void executeRequest(String path) {
        executeRequest('GET', path, null, null, null, null)
    }

    void executeRequest(String path, Map headers) {
        executeRequest('GET', path, headers, null, null, null)
    }

    void executeRequest(String path, Map headers, Map queryParameters) {
        executeRequest('GET', path, headers, queryParameters, null, null)
    }

    void logError(String callType) {
        def body = this.getResponseBody().getText('UTF-8')
        if (body)
            logger.error("🛑 Response body = ${body}")
        throw new HTTPExecuterException("${callType} call failed with response code = ${this.getResponseCode()}")
    }

    abstract InputStream getResponseBody()

    abstract Map getResponseHeaders()

    abstract String getResponseHeader(String name)

    abstract int getResponseCode()

}