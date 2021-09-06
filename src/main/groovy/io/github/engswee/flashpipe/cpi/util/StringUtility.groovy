package io.github.engswee.flashpipe.cpi.util

class StringUtility {

    static List extractDelimitedValues(String input, String delimiter) {
        return input ? input.split(delimiter).toList()*.trim() : []
    }
}