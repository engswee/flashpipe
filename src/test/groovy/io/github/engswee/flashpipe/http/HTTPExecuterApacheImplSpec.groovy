package io.github.engswee.flashpipe.http

import spock.lang.Specification

class HTTPExecuterApacheImplSpec extends Specification {

    final static String LOCALHOST = 'localhost'

    def 'Missing mandatory input exception during instantiation'() {
        when:
        HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, '', '')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Mandatory input user/password is missing'
    }
}