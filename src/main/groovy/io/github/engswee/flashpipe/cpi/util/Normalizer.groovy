package io.github.engswee.flashpipe.cpi.util

class Normalizer {
    static String normalize(String input, String normalizeAction, String normalizePrefixOrSuffix) {
        switch (normalizeAction) {
            case 'ADD_PREFIX':
                return "${normalizePrefixOrSuffix}${input}"
            case 'ADD_SUFFIX':
                return "${input}${normalizePrefixOrSuffix}"
            case 'DELETE_PREFIX':
                return (input.startsWith(normalizePrefixOrSuffix)) ? input.replaceFirst(normalizePrefixOrSuffix, '') : input
            case 'DELETE_SUFFIX':
                if ((input.endsWith(normalizePrefixOrSuffix))) {
                    return input.substring(0, input.size() - normalizePrefixOrSuffix.size())
                } else {
                    return input
                }
            default:
                return input
        }
    }
}
