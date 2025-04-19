#!/bin/bash

# Open a new terminal and run backend
code --remote-terminal -e "workbench.action.terminal.new" -e "workbench.action.terminal.sendSequence" -a "make run-be\u000D"

# Open another terminal and run frontend
code --remote-terminal -e "workbench.action.terminal.new" -e "workbench.action.terminal.sendSequence" -a "make run-fe\u000D"
