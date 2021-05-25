package io.github.engswee.flashpipe.http

import org.apache.hc.client5.http.auth.UsernamePasswordCredentials
import org.apache.hc.client5.http.impl.auth.BasicScheme
import org.apache.hc.client5.http.impl.classic.HttpClients
import org.apache.hc.client5.http.protocol.HttpClientContext
import org.apache.hc.core5.http.ClassicHttpRequest
import org.apache.hc.core5.http.ContentType
import org.apache.hc.core5.http.Header
import org.apache.hc.core5.http.HttpHost
import org.apache.hc.core5.http.io.support.ClassicRequestBuilder
import org.apache.hc.core5.net.URIBuilder
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class HTTPExecuterApacheImpl extends HTTPExecuter {

    String scheme
    String host
    int port
    HttpClientContext context
    int responseCode
    Header[] headers
    byte[] responseBytes
    String token

    static Logger logger = LoggerFactory.getLogger(HTTPExecuterApacheImpl)

    private HTTPExecuterApacheImpl() {
    }

    static HTTPExecuter newInstance(String scheme, String host, int port, String user, String password) {
        if (!host || !scheme || !port)
            throw new HTTPExecuterException('Mandatory input scheme/host/port is missing')
        def httpExecuter = new HTTPExecuterApacheImpl()
        httpExecuter.setBaseURL(scheme, host, port)
        logger.info("Using Basic Authentication for ${scheme}://${host}:${port}")
        httpExecuter.setBasicAuth(user, password)
        return httpExecuter
    }

    static HTTPExecuter newInstance(String scheme, String host, int port, String token) {
        if (!host || !scheme || !port)
            throw new HTTPExecuterException('Mandatory input scheme/host/port is missing')
        def httpExecuter = new HTTPExecuterApacheImpl()
        httpExecuter.setBaseURL(scheme, host, port)
        logger.info("Using OAuth 2.0 Authentication ${scheme}://${host}:${port}")
        httpExecuter.setOAuthToken(token)
        return httpExecuter
    }

    static HTTPExecuter newInstance(String scheme, String host, int port, String user, String password, String token) {
        if (token)
            return newInstance(scheme, host, port, token)
        else
            return newInstance(scheme, host, port, user, password)
    }

    @Override
    void setBaseURL(String scheme, String host, int port) {
        this.scheme = scheme
        this.host = host
        this.port = port
        this.context = HttpClientContext.create()
    }

    @Override
    void setBasicAuth(String user, String password) {
        if (!user || !password)
            throw new HTTPExecuterException('Mandatory input user/password is missing')
        final HttpHost targetHost = new HttpHost(this.scheme, this.host, this.port)
        final BasicScheme basicAuth = new BasicScheme()
        basicAuth.initPreemptive(new UsernamePasswordCredentials(user, password.toCharArray()))

        this.context.resetAuthExchange(targetHost, basicAuth)
    }

    @Override
    void setOAuthToken(String token) {
        if (!token)
            throw new HTTPExecuterException('Mandatory input token is missing')
        this.token = token
    }

    @Override
    void executeRequest(String method, String path, Map headers, Map queryParameters, byte[] requestBytes, String mimeType) {
        ClassicRequestBuilder builder
        switch (method) {
            case 'GET':
                builder = ClassicRequestBuilder.get()
                break
            case 'POST':
                builder = ClassicRequestBuilder.post()
                break
            case 'DELETE':
                builder = ClassicRequestBuilder.delete()
                break
            default:
                builder = ClassicRequestBuilder.create(method)
        }
        builder = builder.setUri(createURI(path, queryParameters))
        if (headers) {
            headers.each { key, value ->
                builder.setHeader(key, value)
            }
        }
        // If an OAuth token is set, use it in the Authorization header
        if (this.token) {
            builder.setHeader('Authorization', "Bearer ${this.token}")
        }

        if (requestBytes)
            builder.setEntity(requestBytes, ContentType.create(mimeType))
        ClassicHttpRequest request = builder.build()

        HttpClients.createDefault().withCloseable { client ->
            client.execute(request, this.context).withCloseable { response ->
                this.responseCode = response.getCode()
                this.headers = response.getHeaders()
                this.responseBytes = response.getEntity().getContent().getBytes()
            }
        }
    }

    @Override
    InputStream getResponseBody() {
        return new ByteArrayInputStream(this.responseBytes)
    }

    @Override
    Map getResponseHeaders() {
        // TODO
        return null
    }

    @Override
    String getResponseHeader(String name) {
        String headerValue
        for (Header header : this.headers) {
            if (header.getName().toUpperCase() == name.toUpperCase()) {
                headerValue = header.getValue()
            }
        }
        return headerValue
    }

    @Override
    int getResponseCode() {
        return this.responseCode
    }

    private URI createURI(String path, Map queryParameters) {
        URIBuilder builder = new URIBuilder()
        builder.setScheme(this.scheme)
                .setHost(this.host)
                .setPort(this.port)
                .setPath(path)
        if (queryParameters) {
            queryParameters.each { k, v ->
                builder.setParameter(k, v)
            }
        }
        return builder.build()
    }
}