#!/bin/bash
cd /Users/jacksonc/i/kitten-crypto-mining-ventures/.worktrees/sprint-1
PROMPT=$(cat /Users/jacksonc/i/kitten-crypto-mining-ventures/.autonomous/sprint-prompt.md)
exec claude --dangerously-skip-permissions "$PROMPT"
