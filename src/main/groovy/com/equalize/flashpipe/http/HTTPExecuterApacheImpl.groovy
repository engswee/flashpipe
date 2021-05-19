package com.equalize.flashpipe.http

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

class HTTPExecuterApacheImpl extends HTTPExecuter {

    String scheme
    String host
    int port
    HttpClientContext context
    int responseCode
    Header[] headers
    byte[] responseBytes

    @Override
    void setBaseURL(String scheme, String host, int port) {
        this.scheme = scheme
        this.host = host
        this.port = port
        this.context = HttpClientContext.create()
    }

    @Override
    void setBasicAuth(String user, String password) {
        final HttpHost targetHost = new HttpHost(this.scheme, this.host, this.port)
        final BasicScheme basicAuth = new BasicScheme()
        basicAuth.initPreemptive(new UsernamePasswordCredentials(user, password.toCharArray()))

        this.context.resetAuthExchange(targetHost, basicAuth)
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