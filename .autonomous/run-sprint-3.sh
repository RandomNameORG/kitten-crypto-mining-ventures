#!/bin/bash
cd /Users/jacksonc/i/kitten-crypto-mining-ventures/.worktrees/sprint-3
PROMPT=$(cat /Users/jacksonc/i/kitten-crypto-mining-ventures/.autonomous/sprint-prompt.md)
exec claude --dangerously-skip-permissions "$PROMPT"
