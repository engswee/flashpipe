package io.github.engswee.flashpipe.cpi.util

class ScriptCollection {

    final Map collections

    static ScriptCollection newInstance(String collectionMap) {
        return new ScriptCollection(collectionMap)
    }

    private ScriptCollection() {
    }

    private ScriptCollection(String collectionMap) {
        // TODO - error handling on incorrect input values
        if (collectionMap && collectionMap.trim()) {
            this.collections = collectionMap.split(',')?.toList()?.collectEntries {
                String[] pair = it.split('=')
                [(pair[0]): pair[1]]
            }
        } else {
            this.collections = [:]
        }
    }

    List getTargetCollectionValues() {
        return this.collections.collect { it.value }
    }
}