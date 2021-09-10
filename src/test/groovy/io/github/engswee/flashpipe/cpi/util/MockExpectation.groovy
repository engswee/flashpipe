package io.github.engswee.flashpipe.cpi.util

import org.mockserver.client.MockServerClient
import org.mockserver.model.HttpRequest
import org.mockserver.model.HttpResponse

class MockExpectation {

    final MockServerClient mockServerClient

    private MockExpectation() {}

    private MockExpectation(String host, int port) {
        this.mockServerClient = new MockServerClient(host, port)
    }

    static MockExpectation newInstance(String host, int port) {
        return new MockExpectation(host, port)
    }

    void set(String method, String path, Integer responseCode, String responseBody) {
        set(method, path, [:], [:], responseCode, responseBody, [:])
    }

    void set(String method, String path, Map<String, String> requestHeaders, Integer responseCode, String responseBody) {
        set(method, path, requestHeaders, [:], responseCode, responseBody, [:])
    }

    void set(String method, String path, Map<String, String> requestHeaders, Map<String, String> requestQueryParameters, Integer responseCode, String responseBody, Map<String, String> responseHeaders) {
        // Request
        HttpRequest httpRequest = HttpRequest.request()
                .withMethod(method)
                .withPath(path)
        requestHeaders.each { String key, String value ->
            httpRequest.withHeader(key, value)
        }
        requestQueryParameters.each { String key, String value ->
            httpRequest.withQueryStringParameter(key, value)
        }
        // Response    
        HttpResponse httpResponse = HttpResponse.response()
                .withStatusCode(responseCode)
                .withBody(responseBody)
        responseHeaders.each { String key, String value ->
            httpResponse.withHeader(key, value)
        }
        this.mockServerClient.when(httpRequest).respond(httpResponse)
    }

    void setCSRFTokenExpectation(String path, String token, Integer httpResponseStatusCode) {
        set('GET', path, ['x-csrf-token': 'fetch'], [:], httpResponseStatusCode, '', ['x-csrf-token': token])
    }

    void setCSRFTokenExpectation(String path, String token) {
        setCSRFTokenExpectation(path, token, 200)
    }
}