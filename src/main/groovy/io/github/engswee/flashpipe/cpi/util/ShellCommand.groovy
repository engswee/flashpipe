package io.github.engswee.flashpipe.cpi.util

import org.slf4j.Logger
import org.slf4j.LoggerFactory

class ShellCommand {
    static Logger logger = LoggerFactory.getLogger(ShellCommand)

    final String shell
    Process process

    ShellCommand(String shell) {
        this.shell = shell
    }
    
    void execute(String command) {
        logger.info("Executing shell command: ${command}")
        ProcessBuilder processBuilder = new ProcessBuilder()
        processBuilder.command(this.shell, '-c', command)
        this.process = processBuilder.start()
        this.process.waitFor()
    }
    
    int getExitValue() {
        return this.process.exitValue()
    }
    
    String getOutputText() {
        return this.process.getText()
    }
    
    String getErrorText() {
        return this.process.err.text
    }
}
