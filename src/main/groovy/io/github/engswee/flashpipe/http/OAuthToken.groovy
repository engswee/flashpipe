package io.github.engswee.flashpipe.http

import groovy.json.JsonSlurper
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class OAuthToken {

    HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(OAuthToken)

    static String get(String scheme, String host, int port, String user, String password) {
        def instance = new OAuthToken(scheme, host, port, user, password)
        return instance.getToken()
    }

    private OAuthToken() {
    }

    private OAuthToken(String scheme, String host, int port, String user, String password) {
        this.httpExecuter = HTTPExecuterApacheImpl.newInstance(scheme, host, port, user, password)
    }

    private String getToken() {
        logger.debug('Get OAuth token')
        this.httpExecuter.executeRequest('/oauth/token', [:], ['grant_type': 'client_credentials'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.access_token
        } else
            this.httpExecuter.logError('Get OAuth token')
    }
}