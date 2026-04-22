#!/bin/bash
cd /Users/jacksonc/i/kitten-crypto-mining-ventures/.worktrees/sprint-4
PROMPT=$(cat /Users/jacksonc/i/kitten-crypto-mining-ventures/.autonomous/worker-prompt.md)
exec claude --dangerously-skip-permissions "$PROMPT"
