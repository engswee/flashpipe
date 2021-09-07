package io.github.engswee.flashpipe.cpi.util

import spock.lang.Specification

class StringUtilitySpec extends Specification {

    def 'No input value'() {
        when:
        List entries = StringUtility.extractDelimitedValues('', ',')

        then:
        entries.size() == 0
    }

    def 'Single value extracted'() {
        when:
        List entries = StringUtility.extractDelimitedValues('ABC', ',')
        then:
        verifyAll {
            entries.size() == 1
            entries[0] == 'ABC'
        }
    }

    def 'Multiple values extracted'() {
        when:
        List entries = StringUtility.extractDelimitedValues('ABC, 123,XYZ ', ',')
        then:
        verifyAll {
            entries.size() == 3
            entries[0] == 'ABC'
            entries[1] == '123'
            entries[2] == 'XYZ'
        }
    }
}