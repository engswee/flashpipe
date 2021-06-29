package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.http.HTTPExecuter
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class CSRFToken {

    final HTTPExecuter httpExecuter
    String token

    static Logger logger = LoggerFactory.getLogger(CSRFToken)

    CSRFToken(HTTPExecuter httpExecuter) {
        this.httpExecuter = httpExecuter
    }

    String get() {
        if (this.token) {
            return this.token
        } else {
            logger.debug('Get CSRF Token')
            httpExecuter.executeRequest('/api/v1/', ['x-csrf-token': 'fetch'])
            def code = httpExecuter.getResponseCode()
            if (code == 200) {
                this.token = httpExecuter.getResponseHeader('x-csrf-token')
                logger.debug("Received CSRF Token - ${this.token}")
                return this.token
            } else
                this.httpExecuter.logError('Get CSRF Token')
        }
    }
}