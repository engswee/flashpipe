package io.github.engswee.flashpipe.http

import spock.lang.Specification

class HTTPExecuterApacheImplSpec extends Specification {

    final static String LOCALHOST = 'localhost'

    def 'Missing mandatory scheme/host/port exception during instantiation'() {
        when:
        HTTPExecuterApacheImpl.newInstance('', LOCALHOST, 9443, '', '')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Mandatory input scheme/host/port is missing'
    }

    def 'Missing mandatory user/password exception during instantiation'() {
        when:
        HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, '', '')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Mandatory input user/password is missing'
    }

    def 'Missing mandatory token exception during instantiation'() {
        when:
        HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, '')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Mandatory input token is missing'
    }

    def 'Missing mandatory scheme/host/port exception during generic instantiation'() {
        when:
        HTTPExecuterApacheImpl.newInstance('http', '', 9443, '', '', 'token')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Mandatory input scheme/host/port is missing'
    }

    def 'No exception thrown when mandatory input provided during generic instantiation with token'() {
        when:
        HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, '', '', 'token')

        then:
        noExceptionThrown()
    }

    def 'No exception thrown when mandatory input provided during generic instantiation without token'() {
        when:
        HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, 'user', 'password', '')

        then:
        noExceptionThrown()
    }
}